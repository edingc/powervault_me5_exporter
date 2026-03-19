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

package client

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// ME5Client handles authenticated communication with the Dell PowerVault ME5 REST API.
type ME5Client struct {
	host       string
	httpClient *http.Client
	password   string
	username   string

	mu         sync.Mutex
	sessionExp time.Time
	sessionKey string
}

// NewME5Client creates a new ME5 API client.
func NewME5Client(host, username, password string, timeout time.Duration, insecureSkipVerify bool) *ME5Client {
	return &ME5Client{
		host:     host,
		password: password,
		username: username,
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
			},
		},
	}
}

func (c *ME5Client) baseURL() string {
	return fmt.Sprintf("https://%s/api", c.host)
}

func hashCredentials(username, password string) string {
	h := sha256.New()

	h.Write([]byte(username))
	h.Write([]byte("_"))
	h.Write([]byte(password))

	return fmt.Sprintf("%x", h.Sum(nil))
}

// login authenticates against the ME5 API and stores the session key.
func (c *ME5Client) login(ctx context.Context) error {
	loginURL := fmt.Sprintf("%s/login/%s", c.baseURL(), hashCredentials(c.username, c.password))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, loginURL, nil)
	if err != nil {
		return fmt.Errorf("building login request: %w", err)
	}
	req.Header.Set("datatype", "json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request: %w", err)
	}
	defer resp.Body.Close()

	var body struct {
		Status []struct {
			Response            string `json:"response"`
			ResponseTypeNumeric int    `json:"response-type-numeric"`
		} `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("decoding login response: %w", err)
	}

	for _, s := range body.Status {
		if s.ResponseTypeNumeric == 0 { // 0 = success
			c.sessionKey = s.Response
			// Sessions expire after 30 min; we expire locally at 25 min to be safe.
			c.sessionExp = time.Now().Add(25 * time.Minute)
			return nil
		}
	}
	return fmt.Errorf("login failed: %+v", body.Status)
}

// ensureSession returns a valid session key, logging in first if necessary.
func (c *ME5Client) ensureSession(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sessionKey == "" || time.Now().After(c.sessionExp) {
		if err := c.login(ctx); err != nil {
			return "", err
		}
	}
	return c.sessionKey, nil
}

// Get performs an authenticated GET against the given ME5 API path.
// It will re-authenticate and retry once on 401 Unauthorized status.
func (c *ME5Client) Get(ctx context.Context, path string, dest any) error {
	return c.get(ctx, path, dest, false)
}

func (c *ME5Client) get(ctx context.Context, path string, dest any, retried bool) error {
	key, err := c.ensureSession(ctx)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL()+path, nil)
	if err != nil {
		return fmt.Errorf("building request for %s: %w", path, err)
	}
	req.Header.Set("sessionKey", key)
	req.Header.Set("datatype", "json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized && !retried {
		slog.Warn("ME5 session invalidated server-side, re-authenticating", "path", path)
		c.mu.Lock()
		c.sessionExp = time.Time{} // Force re-login on next attempt
		c.mu.Unlock()
		return c.get(ctx, path, dest, true)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GET %s: HTTP %d: %s", path, resp.StatusCode, body)
	}

	return json.NewDecoder(resp.Body).Decode(dest)
}
