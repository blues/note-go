package notehub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// Exercise the actual localhost callback server: a denied login must release
// the caller, while unrelated requests and wrong states must not consume it.
func TestBrowserLoginDenialReturns(t *testing.T) {
	const name = "Notehub CLI - Ray's Mac & 日本語 + <test>"
	done := make(chan error, 1)
	go func() {
		_, err := initiateBrowserBasedLogin("notehub.example", func(authorize string) error {
			u, err := url.Parse(authorize)
			if err != nil {
				return err
			}
			if u.Query().Get("signin_name") != name {
				t.Errorf("lost sign-in name: %s", u.Query().Get("signin_name"))
			}
			callback := u.Query().Get("redirect_uri")
			// Don't follow the redirect to the (nonexistent) Notehub completion page.
			client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
			for _, suffix := range []string{"/favicon.ico", "?code=unrelated&state=wrong", "?error=access_denied&error_description=" + url.QueryEscape("The user decided not to sign in.") + "&state=" + u.Query().Get("state")} {
				resp, err := client.Get(callback + suffix)
				if err != nil {
					return err
				}
				_, err = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				if err != nil {
					return fmt.Errorf("read callback response: %w", err)
				}
				if strings.HasPrefix(suffix, "?error=") {
					if location := resp.Header.Get("Location"); resp.StatusCode != http.StatusSeeOther || location != "https://notehub.example/login-complete?result=denied" {
						t.Errorf("denial was not sent to the completion page: %d %s", resp.StatusCode, location)
					}
				}
			}
			return nil
		}, BrowserLoginOptions{SignInName: name})
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "access_denied") {
			t.Fatalf("expected denial, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("denied login did not return")
	}
}

func TestPollingLoginTerminalResults(t *testing.T) {
	for _, status := range []string{"denied", "failed", "expired", "code_available"} {
		t.Run(status, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer private-token" {
					t.Error("missing private polling credential")
				}
				if status == "expired" {
					w.WriteHeader(http.StatusGone)
				}
				io.WriteString(w, `{"status":"`+status+`","code":"test-code","error":"access_denied","error_description":"The user decided not to sign in."}`)
			}))
			defer server.Close()
			called := false
			token, err := awaitPollingLogin(context.Background(), server.Client(), server.URL, pollLoginResponse{PollToken: "private-token", Interval: 1}, func(code string) (*AccessToken, error) {
				called = true
				if code != "test-code" {
					t.Errorf("unexpected code %q", code)
				}
				return &AccessToken{AccessToken: "token"}, nil
			})
			if status == "code_available" {
				if err != nil || token == nil || !called {
					t.Fatalf("code exchange failed: %v", err)
				}
			} else {
				if err == nil || called {
					t.Fatalf("terminal %s did not stop polling", status)
				}
				if status == "denied" && err.Error() != authorizationError("access_denied", "The user decided not to sign in.").Error() {
					t.Fatalf("lost cancellation description: %v", err)
				}
			}
		})
	}
}

func TestPollingLoginCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := awaitPollingLogin(ctx, http.DefaultClient, "http://unused", pollLoginResponse{Interval: 5}, func(string) (*AccessToken, error) { t.Fatal("unexpected exchange"); return nil, nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}

func TestAuthorizationCodeExchange(t *testing.T) {
	for _, redirect := range []string{"http://localhost:58766", "https://notehub.example/auth/poll"} {
		t.Run(redirect, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/oauth2/token":
					if err := r.ParseForm(); err != nil {
						t.Errorf("parse token request form: %v", err)
						return
					}
					if r.Form.Get("redirect_uri") != redirect || r.Form.Get("code_verifier") != "private-verifier" || r.Form.Get("grant_type") != "authorization_code" {
						t.Errorf("unexpected token request: %v", r.Form)
					}
					io.WriteString(w, `{"access_token":"issued-token","expires_in":3600}`)
				case "/userinfo":
					if r.Header.Get("Authorization") != "Bearer issued-token" {
						t.Error("missing issued token")
					}
					io.WriteString(w, `{"email":"user@example.com"}`)
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
				}
			}))
			defer server.Close()
			original := http.DefaultTransport
			http.DefaultTransport = server.Client().Transport
			defer func() { http.DefaultTransport = original }()
			host := strings.TrimPrefix(server.URL, "https://")
			token, err := exchangeAuthorizationCode(context.Background(), host, host, "notehub_cli", "code", "private-verifier", redirect)
			if err != nil || token == nil || token.Email != "user@example.com" || token.AccessToken != "issued-token" {
				t.Fatalf("exchange failed: %v, %v", token, err)
			}
		})
	}
}

func TestPollingLoginBackoff(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "12")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()
	result, err := requestPollLogin(context.Background(), server.Client(), server.URL, "private-token", map[string]any{"mode": "poll"})
	if err != nil || result.Status != "pending" || result.Interval != 12 {
		t.Fatalf("expected backoff, got %+v, %v", result, err)
	}
}

type authTestTransport func(*http.Request) (*http.Response, error)

func (f authTestTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestPollingLoginExpiresIn(t *testing.T) {
	for _, seconds := range []int{0, 30} {
		client := &http.Client{Transport: authTestTransport(func(r *http.Request) (*http.Response, error) {
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			got, present := payload["expires_in"]
			if present != (seconds != 0) || (present && got != float64(seconds)) {
				t.Errorf("expires_in = %v (present=%v), expected %d", got, present, seconds)
			}
			body := `{"browser_url":"https://notehub.example/oauth2/auth","poll_token":"private-token","expires_in":10,"interval":2}`
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
		})}
		delivered := errors.New("URL delivered")
		_, err := initiatePollingLogin(context.Background(), "api.notehub.example", client, io.Discard,
			func(string) error { t.Fatal("unexpected browser launch"); return nil },
			PollingLoginOptions{ExpiresInSeconds: seconds, OnURL: func(url string, granted int) error {
				if granted != 10 {
					t.Errorf("granted = %d, expected the hub's 10", granted)
				}
				return delivered
			}})
		if !errors.Is(err, delivered) {
			t.Fatalf("expected the URL to be delivered, got %v", err)
		}
	}
}

func TestPollingLoginSignInName(t *testing.T) {
	for _, name := range []string{"", "Notehub CLI - Ray's Mac & 日本語 + <test>"} {
		t.Run(name, func(t *testing.T) {
			client := &http.Client{Transport: authTestTransport(func(r *http.Request) (*http.Response, error) {
				var payload map[string]string
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Fatal(err)
				}
				got, present := payload["signin_name"]
				if got != name || present != (name != "") {
					t.Errorf("signin_name = %q (present=%v), expected %q", got, present, name)
				}
				body := `{"browser_url":"https://notehub.example/oauth2/auth","poll_token":"private-token","expires_in":10,"interval":2}`
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
			})}
			delivered := errors.New("URL delivered")
			_, err := initiatePollingLogin(context.Background(), "api.notehub.example", client, io.Discard,
				func(string) error { t.Fatal("unexpected browser launch"); return nil },
				PollingLoginOptions{SignInName: name, OnURL: func(string, int) error { return delivered }})
			if !errors.Is(err, delivered) {
				t.Fatalf("expected URL delivery, got %v", err)
			}
		})
	}
}

func TestPollingLoginInteractiveAndAgent(t *testing.T) {
	const longURL = "https://notehub.example/oauth2/auth?state=browser-state"
	const shortURL = "https://notehub.example/auth/poll?login=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	for _, tc := range []struct {
		agent bool
		url   string
	}{
		{false, longURL}, {true, longURL},
		{false, shortURL}, {true, shortURL},
	} {
		agent, browserURL := tc.agent, tc.url
		t.Run(fmt.Sprintf("agent=%v/url=%s", agent, browserURL), func(t *testing.T) {
			var stderr strings.Builder
			client := &http.Client{Transport: authTestTransport(func(r *http.Request) (*http.Response, error) {
				if r.URL.String() != "https://notehub.example/auth/poll" {
					t.Errorf("wrong configured hub: %s", r.URL)
				}
				body := `{"browser_url":"` + browserURL + `","poll_token":"private-token","expires_in":10,"interval":1}`
				if r.Header.Get("Authorization") != "" {
					if r.Header.Get("Authorization") != "Bearer private-token" {
						t.Error("wrong polling credential")
					}
					body = `{"status":"denied","error":"access_denied","error_description":"The user decided not to sign in."}`
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
			})}
			opened, delivered := "", ""
			options := PollingLoginOptions{}
			if agent {
				options.OnURL = func(url string, expiresIn int) error { delivered = url; return nil }
			}
			_, err := initiatePollingLogin(context.Background(), "api.notehub.example", client, &stderr, func(url string) error { opened = url; return nil }, options)
			if err == nil || !strings.Contains(err.Error(), "The user decided not to sign in.") {
				t.Fatalf("lost cancellation description: %v", err)
			}
			if agent {
				if opened != "" || delivered != browserURL || stderr.Len() != 0 {
					t.Fatalf("agent opened browser or printed human messages: %q %q %q", opened, delivered, stderr.String())
				}
			} else {
				if opened != browserURL || delivered != "" || stderr.String() != browserLoginMessage {
					t.Fatalf("interactive browser/output mismatch: %q %q %q", opened, delivered, stderr.String())
				}
			}
		})
	}
}

func TestPollingLoginTenSecondProgress(t *testing.T) {
	// Keep a poll request blocked: progress must still arrive at ten and twenty
	// seconds, independently of network latency and the server's poll interval.
	ctx, cancel := context.WithTimeout(context.Background(), 21*time.Second)
	defer cancel()
	client := &http.Client{Transport: authTestTransport(func(r *http.Request) (*http.Response, error) { <-r.Context().Done(); return nil, r.Context().Err() })}
	started := time.Now()
	var progress []time.Duration
	var remaining []int
	_, err := awaitPollingLoginWithProgress(ctx, client, "https://notehub.example/auth/poll", pollLoginResponse{Interval: 1}, func(string) (*AccessToken, error) { t.Fatal("unexpected token exchange"); return nil, nil }, func(seconds int) error {
		progress = append(progress, time.Since(started))
		remaining = append(remaining, seconds)
		return nil
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected timeout, got %v", err)
	}
	if len(progress) != 2 {
		t.Fatalf("expected two progress messages, got %v", progress)
	}
	for i, at := range progress {
		if remaining[i] < 0 || remaining[i] > 11-i*10 || remaining[i] < 9-i*10 {
			t.Fatalf("incorrect remaining timeout at %s: %d", at, remaining[i])
		}
		expected := time.Duration(i+1) * 10 * time.Second
		if at < expected || at > expected+2*time.Second {
			t.Fatalf("progress %d at %s, expected %s", i, at, expected)
		}
	}
}

func TestPollingLoginTwoSecondInterval(t *testing.T) {
	var requests []time.Time
	client := &http.Client{Transport: authTestTransport(func(r *http.Request) (*http.Response, error) {
		requests = append(requests, time.Now())
		body := `{"status":"pending"}`
		if len(requests) == 3 {
			body = `{"status":"code_available","code":"code"}`
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	// The new CLI also polls every two seconds against servers still advertising five.
	code, err := pollAuthorizationCode(ctx, client, "https://notehub.example/auth/poll", pollLoginResponse{Interval: 5})
	if err != nil || code != "code" || len(requests) != 3 {
		t.Fatalf("polling did not finish: %q %v %v", code, err, requests)
	}
	for i := 1; i < len(requests); i++ {
		elapsed := requests[i].Sub(requests[i-1])
		if elapsed < 2*time.Second || elapsed > 2500*time.Millisecond {
			t.Fatalf("poll interval was %s", elapsed)
		}
	}
}

// randString feeds the PKCE code verifier, which RFC 7636 limits to
// A-Z a-z 0-9 and - . _ ~, and which must be at least 43 characters.
func TestRandStringLengthAndCharset(t *testing.T) {
	const allowed = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

	for _, n := range []int{16, 43, 50} {
		got := randString(n)
		if len(got) != n {
			t.Errorf("randString(%d) returned %d characters", n, len(got))
		}
		for _, r := range got {
			if !strings.ContainsRune(allowed, r) {
				t.Errorf("randString(%d) returned %q, which RFC 7636 does not allow in a code verifier", n, r)
			}
		}
	}
}

// A generator that returns the same value twice would hand every login the same
// code verifier.
func TestRandStringDoesNotRepeat(t *testing.T) {
	seen := make(map[string]bool, 100)
	for i := 0; i < 100; i++ {
		s := randString(32)
		if seen[s] {
			t.Fatalf("randString(32) returned %q twice in 100 calls", s)
		}
		seen[s] = true
	}
}
