// Copyright 2026 Cody Eding
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"os/user"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/edingc/powervault_me5_exporter/internal/client"
	"github.com/edingc/powervault_me5_exporter/internal/collector"
	"github.com/prometheus/client_golang/prometheus"
	promcollectors "github.com/prometheus/client_golang/prometheus/collectors"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

var (
	me5Host                = kingpin.Flag("me5.host", "Dell PowerVault ME5 hostname or IP (required)").String()
	me5User                = kingpin.Flag("me5.username", "ME5 API username").Envar("ME5_USERNAME").String()
	me5Password            = kingpin.Flag("me5.password", "ME5 API password").Envar("ME5_PASSWORD").String()
	timeout                = kingpin.Flag("me5.timeout", "PowerVault API connect timeout.").Default("30s").Duration()
	enablePprof            = kingpin.Flag("web.enable-pprof", "Enable pprof profiling endpoints under /debug/pprof.").Bool()
	metricsPath            = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	disableExporterMetrics = kingpin.Flag("web.disable-exporter-metrics", "Exclude metrics about the exporter itself.").Bool()
	maxRequests            = kingpin.Flag("web.max-requests", "Maximum number of parallel scrape requests.").Default("1").Int()
	insecureSkipVerify     = kingpin.Flag("me5.insecure-skip-verify", "Skip TLS certificate verification.").Bool()
	toolkitFlags           = kingpinflag.AddFlags(kingpin.CommandLine, ":9850")
)

func main() {
	// Consolidate collector flag logic
	collectorFlags := make(map[string]*bool)
	for name, enabledByDefault := range collector.AllCollectors {
		collectorFlags[name] = kingpin.Flag(
			"collect."+name,
			collector.CollectorHelp[name],
		).Default(fmt.Sprint(enabledByDefault)).Bool()
	}

	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.Version(version.Print("powervault_me5_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promslog.New(promslogConfig)
	logger.Info("Starting powervault_me5_exporter", "version", version.Info())

	if *me5Host == "" {
		logger.Error("--me5.host is required")
		os.Exit(1)
	}

	if *me5User == "" || *me5Password == "" {
		logger.Error("API credentials must be provided via flags or environment variables.")
		os.Exit(1)
	}

	if u, err := user.Current(); err == nil && u.Uid == "0" {
		logger.Warn("Running as root is not required and discouraged.")
	}

	c := client.NewME5Client(*me5Host, *me5User, *me5Password, *timeout, *insecureSkipVerify)

	enabled := make(map[string]bool)
	var enabledNames []string
	for name, f := range collectorFlags {
		if *f {
			enabled[name] = true
			enabledNames = append(enabledNames, name)
		}
	}
	sort.Strings(enabledNames)
	logger.Info("Enabled collectors", "collectors", strings.Join(enabledNames, ", "))

	reg := prometheus.NewRegistry()
	reg.MustRegister(versioncollector.NewCollector("powervault_me5_exporter"))
	reg.MustRegister(collector.NewME5Collector(c, enabled))
	// https://github.com/prometheus/node_exporter/pull/3513/changes
	// Use a dedicated ServeMux to have explicit control over exposed routes.
	// Dependencies might register routes on DefaultServeMux. Be sure to check them.
	// (e.g. [net/http/pprof](https://pkg.go.dev/net/http/pprof) registers debug endpoints there via init()).
	// Avoids accidentally serving handlers that dependencies might register on DefaultServeMux.
	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Go runtime profiling endpoints for CPU, memory, and goroutine analysis.
	// These would normally be registered on DefaultServeMux by importing net/http/pprof,
	// but we register them explicitly on our mux for controlled exposure.
	// See [net/http/pprof package docs](https://pkg.go.dev/net/http/pprof) for more details.
	if *enablePprof {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	// Metrics Handler with context timeout integration
	opts := promhttp.HandlerOpts{
		ErrorLog:            slog.NewLogLogger(logger.Handler(), slog.LevelError),
		ErrorHandling:       promhttp.ContinueOnError,
		MaxRequestsInFlight: *maxRequests,
	}

	var metricsHandler http.Handler
	exporterReg := prometheus.NewRegistry()
	if !*disableExporterMetrics {
		exporterReg.MustRegister(promcollectors.NewProcessCollector(promcollectors.ProcessCollectorOpts{}), promcollectors.NewGoCollector())
	}

	// Wrap handler to inject timeout from flags/request
	metricsHandler = promhttp.HandlerFor(prometheus.Gatherers{exporterReg, reg}, opts)
	if !*disableExporterMetrics {
		metricsHandler = promhttp.InstrumentMetricHandler(exporterReg, metricsHandler)
	}

	mux.Handle(*metricsPath, metricsHandler)

	// Landing Page
	landingConfig := web.LandingConfig{
		Name:        "Dell PowerVault ME5 Exporter",
		Description: "Prometheus Exporter for Dell PowerVault ME5",
		Version:     version.Info(),
		Links: []web.LandingLinks{
			{Address: *metricsPath, Text: "Metrics"},
			{Address: "/health", Text: "Health"},
		},
		Profiling: fmt.Sprintf("%t", *enablePprof), // Display profiling links on landing page only if endpoint is enabled
	}
	landingPage, _ := web.NewLandingPage(landingConfig)
	mux.Handle("/", landingPage)

	srv := &http.Server{Handler: mux}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		if err := web.ListenAndServe(srv, toolkitFlags, logger); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Error starting HTTP server", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down gracefully")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
