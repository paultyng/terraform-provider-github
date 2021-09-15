package github

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v39/github"
)

func TestEtagTransport(t *testing.T) {
	ts := githubApiMock([]*mockResponse{
		{
			ExpectedUri: "/repos/test/blah",
			ExpectedHeaders: map[string]string{
				"If-None-Match": "something",
			},

			ResponseBody: `{"id": 1234}`,
			StatusCode:   200,
		},
	})
	defer ts.Close()

	httpClient := http.DefaultClient
	httpClient.Transport = NewEtagTransport(http.DefaultTransport)

	client := github.NewClient(httpClient)
	u, _ := url.Parse(ts.URL + "/")
	client.BaseURL = u

	ctx := context.WithValue(context.Background(), ctxEtag, "something")
	r, _, err := client.Repositories.Get(ctx, "test", "blah")
	if err != nil {
		t.Fatal(err)
	}

	if r.GetID() != 1234 {
		t.Fatalf("Expected ID to be 1234, got: %d", r.GetID())
	}
}

func githubApiMock(responseSequence []*mockResponse) *httptest.Server {
	position := github.Int(0)
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Server", "GitHub.com")

		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("[ERROR] %s", err)
		}
		log.Printf("[DEBUG] Mock server received %s request to %q; headers:\n%s\nrequest body: %q\n",
			r.Method, r.RequestURI, r.Header, string(bodyBytes))

		i := *position
		if len(responseSequence) < i+1 {
			w.WriteHeader(400)
			return
		}

		tc := responseSequence[i]

		headersMatch := func(h http.Header, expectedHeaders map[string]string) bool {
			for key, value := range expectedHeaders {
				if h.Get(key) != value {
					return false
				}
			}
			return true
		}

		if r.RequestURI != tc.ExpectedUri {
			log.Printf("[ERROR] Expected URI: %q, given: %q", tc.ExpectedUri, r.RequestURI)
			w.WriteHeader(400)
			return
		}
		if !headersMatch(r.Header, tc.ExpectedHeaders) {
			log.Printf("[ERROR] Expected headers: %q, given: %q", tc.ExpectedHeaders, r.Header)
			w.WriteHeader(400)
			return
		}
		if tc.ExpectedMethod != "" && r.Method != tc.ExpectedMethod {
			log.Printf("[ERROR] Expected method: %q, given: %q", tc.ExpectedMethod, r.Method)
			w.WriteHeader(400)
			return
		}
		if len(tc.ExpectedBody) > 0 && string(bodyBytes) != string(tc.ExpectedBody) {
			log.Printf("[ERROR] Expected body: %q, given: %q",
				string(tc.ExpectedBody), string(bodyBytes))
			w.WriteHeader(400)
			return
		}

		for key, value := range tc.ResponseHeaders {
			w.Header().Set(key, value)
		}
		w.WriteHeader(tc.StatusCode)
		fmt.Fprintln(w, tc.ResponseBody)

		// Treat response as disposable
		position = github.Int(i + 1)
	}))
}

func TestRateLimitTransport_abuseLimit_get(t *testing.T) {
	ts := githubApiMock([]*mockResponse{
		{
			ExpectedUri: "/repos/test/blah",
			ResponseBody: `{
  "message": "You have triggered an abuse detection mechanism and have been temporarily blocked from content creation. Please retry your request again later.",
  "documentation_url": "https://developer.github.com/v3/#abuse-rate-limits"
}`,
			StatusCode: 403,
			ResponseHeaders: map[string]string{
				"Retry-After": "0.1",
			},
		},
		{
			ExpectedUri: "/repos/test/blah",
			ResponseBody: `{
  "message": "You have triggered an abuse detection mechanism and have been temporarily blocked from content creation. Please retry your request again later.",
  "documentation_url": "https://developer.github.com/v3/#abuse-rate-limits"
}`,
			StatusCode: 403,
			ResponseHeaders: map[string]string{
				"Retry-After": "0.1",
			},
		},
		{
			ExpectedUri:  "/repos/test/blah",
			ResponseBody: `{"id": 1234}`,
			StatusCode:   200,
		},
	})
	defer ts.Close()

	httpClient := http.DefaultClient
	httpClient.Transport = NewRateLimitTransport(http.DefaultTransport)

	client := github.NewClient(httpClient)
	u, _ := url.Parse(ts.URL + "/")
	client.BaseURL = u

	ctx := context.WithValue(context.Background(), ctxId, t.Name())
	r, _, err := client.Repositories.Get(ctx, "test", "blah")
	if err != nil {
		t.Fatal(err)
	}

	if r.GetID() != 1234 {
		t.Fatalf("Expected ID to be 1234, got: %d", r.GetID())
	}
}

func TestRateLimitTransport_abuseLimit_post(t *testing.T) {
	ts := githubApiMock([]*mockResponse{
		{
			ExpectedUri:    "/orgs/tada/repos",
			ExpectedMethod: "POST",
			ExpectedBody: []byte(`{"name":"radek-example-48","description":""}
`),
			ResponseBody: `{
  "message": "You have triggered an abuse detection mechanism and have been temporarily blocked from content creation. Please retry your request again later.",
  "documentation_url": "https://developer.github.com/v3/#abuse-rate-limits"
}`,
			StatusCode: 403,
			ResponseHeaders: map[string]string{
				"Retry-After": "0.1",
			},
		},
		{
			ExpectedUri:    "/orgs/tada/repos",
			ExpectedMethod: "POST",
			ExpectedBody: []byte(`{"name":"radek-example-48","description":""}
`),
			ResponseBody: `{"id": 1234}`,
			StatusCode:   200,
		},
	})
	defer ts.Close()

	httpClient := http.DefaultClient
	httpClient.Transport = NewRateLimitTransport(http.DefaultTransport)

	client := github.NewClient(httpClient)
	u, _ := url.Parse(ts.URL + "/")
	client.BaseURL = u

	ctx := context.WithValue(context.Background(), ctxId, t.Name())
	r, _, err := client.Repositories.Create(ctx, "tada", &github.Repository{
		Name:        github.String("radek-example-48"),
		Description: github.String(""),
	})
	if err != nil {
		t.Fatal(err)
	}

	if r.GetID() != 1234 {
		t.Fatalf("Expected ID to be 1234, got: %d", r.GetID())
	}
}

func TestRateLimitTransport_abuseLimit_post_error(t *testing.T) {
	ts := githubApiMock([]*mockResponse{
		{
			ExpectedUri:    "/orgs/tada/repos",
			ExpectedMethod: "POST",
			ExpectedBody: []byte(`{"name":"radek-example-48","description":""}
`),
			ResponseBody: `{
  "message": "You have triggered an abuse detection mechanism and have been temporarily blocked from content creation. Please retry your request again later.",
  "documentation_url": "https://developer.github.com/v3/#abuse-rate-limits"
}`,
			StatusCode: 403,
			ResponseHeaders: map[string]string{
				"Retry-After": "0.1",
			},
		},
		{
			ExpectedUri:    "/orgs/tada/repos",
			ExpectedMethod: "POST",
			ExpectedBody: []byte(`{"name":"radek-example-48","description":""}
`),
			ResponseBody: `{
  "message": "Repository creation failed.",
  "errors": [
    {
      "resource": "Repository",
      "code": "custom",
      "field": "name",
      "message": "name already exists on this account"
    }
  ],
  "documentation_url": "https://developer.github.com/v3/repos/#create"
}
`,
			StatusCode: 422,
		},
	})
	defer ts.Close()

	httpClient := http.DefaultClient
	httpClient.Transport = NewRateLimitTransport(http.DefaultTransport)

	client := github.NewClient(httpClient)
	u, _ := url.Parse(ts.URL + "/")
	client.BaseURL = u

	ctx := context.WithValue(context.Background(), ctxId, t.Name())
	_, _, err := client.Repositories.Create(ctx, "tada", &github.Repository{
		Name:        github.String("radek-example-48"),
		Description: github.String(""),
	})
	if err == nil {
		t.Fatal("Expected 422 error, got nil")
	}

	ghErr, ok := err.(*github.ErrorResponse)
	if !ok {
		t.Fatalf("Expected github.ErrorResponse, got: %#v", err)
	}

	expectedMessage := "Repository creation failed."
	if ghErr.Message != expectedMessage {
		t.Fatalf("Expected message %q, got: %q", expectedMessage, ghErr.Message)
	}
}

type mockResponse struct {
	ExpectedUri     string
	ExpectedMethod  string
	ExpectedHeaders map[string]string
	ExpectedBody    []byte

	StatusCode      int
	ResponseHeaders map[string]string
	ResponseBody    string
}
