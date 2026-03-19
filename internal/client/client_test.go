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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
)

// loginOK writes a successful ME5 login response with the given session key.
func loginOK(w http.ResponseWriter, key string) {
	json.NewEncoder(w).Encode(struct {
		Status []struct {
			Response            string `json:"response"`
			ResponseTypeNumeric int    `json:"response-type-numeric"`
		} `json:"status"`
	}{
		Status: []struct {
			Response            string `json:"response"`
			ResponseTypeNumeric int    `json:"response-type-numeric"`
		}{{Response: key, ResponseTypeNumeric: 0}},
	})
}

// loginFail writes a failed ME5 login response.
func loginFail(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(struct {
		Status []struct {
			Response            string `json:"response"`
			ResponseTypeNumeric int    `json:"response-type-numeric"`
		} `json:"status"`
	}{
		Status: []struct {
			Response            string `json:"response"`
			ResponseTypeNumeric int    `json:"response-type-numeric"`
		}{{Response: "Authentication Failed", ResponseTypeNumeric: 1}},
	})
}

// newTLSClient starts a TLS test server and returns a matching ME5Client.
// The client is configured with insecureSkipVerify so it trusts the test cert.
func newTLSClient(t *testing.T, handler http.Handler) (*httptest.Server, *ME5Client) {
	t.Helper()
	ts := httptest.NewTLSServer(handler)
	t.Cleanup(ts.Close)
	u, _ := url.Parse(ts.URL)
	c := NewME5Client(u.Host, "user", "pass", 5*time.Second, true)
	return ts, c
}

func TestHashCredentials(t *testing.T) {
	// Verify the username_password concatenation format and SHA-256 encoding.
	h := sha256.New()
	h.Write([]byte("user_pass"))
	want := fmt.Sprintf("%x", h.Sum(nil))

	if got := hashCredentials("user", "pass"); got != want {
		t.Errorf("hashCredentials = %s, want %s", got, want)
	}
}

func TestGet_Success(t *testing.T) {
	const sessionKey = "test-session-key"
	type payload struct{ Value string }

	mux := http.NewServeMux()
	mux.HandleFunc("/api/login/", func(w http.ResponseWriter, r *http.Request) {
		loginOK(w, sessionKey)
	})
	mux.HandleFunc("/api/show/thing", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("sessionKey") != sessionKey {
			http.Error(w, "wrong session key", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("datatype") != "json" {
			http.Error(w, "missing datatype header", http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(payload{Value: "ok"})
	})

	_, c := newTLSClient(t, mux)

	var result payload
	if err := c.Get(context.Background(), "/show/thing", &result); err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if result.Value != "ok" {
		t.Errorf("result.Value = %q, want %q", result.Value, "ok")
	}
}

func TestGet_LoginFails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login/", func(w http.ResponseWriter, r *http.Request) {
		loginFail(w)
	})

	_, c := newTLSClient(t, mux)

	if err := c.Get(context.Background(), "/show/thing", nil); err == nil {
		t.Error("expected error on login failure, got nil")
	}
}

func TestGet_Non200(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login/", func(w http.ResponseWriter, r *http.Request) {
		loginOK(w, "key")
	})
	mux.HandleFunc("/api/show/thing", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	})

	_, c := newTLSClient(t, mux)

	if err := c.Get(context.Background(), "/show/thing", nil); err == nil {
		t.Error("expected error on HTTP 500, got nil")
	}
}

func TestGet_SessionReuse(t *testing.T) {
	var loginCalls atomic.Int32

	mux := http.NewServeMux()
	mux.HandleFunc("/api/login/", func(w http.ResponseWriter, r *http.Request) {
		loginCalls.Add(1)
		loginOK(w, "key")
	})
	mux.HandleFunc("/api/show/thing", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(struct{}{})
	})

	_, c := newTLSClient(t, mux)

	for range 3 {
		if err := c.Get(context.Background(), "/show/thing", &struct{}{}); err != nil {
			t.Fatalf("Get returned error: %v", err)
		}
	}

	if n := loginCalls.Load(); n != 1 {
		t.Errorf("expected 1 login call for 3 Gets, got %d", n)
	}
}

func TestGet_401Retry(t *testing.T) {
	var dataCalls atomic.Int32

	mux := http.NewServeMux()
	mux.HandleFunc("/api/login/", func(w http.ResponseWriter, r *http.Request) {
		loginOK(w, "new-key")
	})
	mux.HandleFunc("/api/show/thing", func(w http.ResponseWriter, r *http.Request) {
		if dataCalls.Add(1) == 1 {
			// Simulate server-side session invalidation on the first call.
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(struct{}{})
	})

	_, c := newTLSClient(t, mux)

	// Pre-seed a session so the first Get skips the initial login.
	c.sessionKey = "old-key"
	c.sessionExp = time.Now().Add(25 * time.Minute)

	if err := c.Get(context.Background(), "/show/thing", &struct{}{}); err != nil {
		t.Fatalf("expected successful retry after 401, got: %v", err)
	}
	if n := dataCalls.Load(); n != 2 {
		t.Errorf("expected 2 data calls (original + retry), got %d", n)
	}
}

func TestLogin_BuildRequestError(t *testing.T) {
	// A null byte in the host makes url.Parse fail inside http.NewRequestWithContext.
	c := NewME5Client("host\x00", "user", "pass", time.Second, true)
	if err := c.Get(context.Background(), "/show/thing", nil); err == nil {
		t.Error("expected error with invalid host, got nil")
	}
}

func TestLogin_NetworkError(t *testing.T) {
	// Nothing is listening on localhost:1; the TCP connection is refused immediately.
	c := NewME5Client("localhost:1", "user", "pass", time.Second, true)
	if err := c.Get(context.Background(), "/show/thing", nil); err == nil {
		t.Error("expected network error, got nil")
	}
}

func TestLogin_JSONDecodeError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	})

	_, c := newTLSClient(t, mux)
	if err := c.Get(context.Background(), "/show/thing", nil); err == nil {
		t.Error("expected error on invalid login JSON, got nil")
	}
}

func TestGet_BuildRequestError(t *testing.T) {
	// A newline in the path makes url.Parse fail inside http.NewRequestWithContext.
	c := NewME5Client("localhost:9999", "user", "pass", time.Second, true)
	c.sessionKey = "key"
	c.sessionExp = time.Now().Add(25 * time.Minute)

	if err := c.Get(context.Background(), "/show/\nthing", nil); err == nil {
		t.Error("expected error with invalid path, got nil")
	}
}

func TestGet_NetworkError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login/", func(w http.ResponseWriter, r *http.Request) {
		loginOK(w, "key")
	})

	ts, c := newTLSClient(t, mux)

	// Pre-seed a session then shut down the server so the data request fails.
	c.sessionKey = "key"
	c.sessionExp = time.Now().Add(25 * time.Minute)
	ts.Close()

	if err := c.Get(context.Background(), "/show/thing", nil); err == nil {
		t.Error("expected network error, got nil")
	}
}

func TestGet_401NoInfiniteRetry(t *testing.T) {
	var dataCalls atomic.Int32

	mux := http.NewServeMux()
	mux.HandleFunc("/api/login/", func(w http.ResponseWriter, r *http.Request) {
		loginOK(w, "key")
	})
	mux.HandleFunc("/api/show/thing", func(w http.ResponseWriter, r *http.Request) {
		dataCalls.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
	})

	_, c := newTLSClient(t, mux)

	if err := c.Get(context.Background(), "/show/thing", nil); err == nil {
		t.Error("expected error after exhausting 401 retries, got nil")
	}
	if n := dataCalls.Load(); n != 2 {
		t.Errorf("expected exactly 2 data calls (original + one retry), got %d", n)
	}
}
