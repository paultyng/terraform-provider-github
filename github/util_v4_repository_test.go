package github

import (
	"bytes"
	"github.com/shurcooL/githubv4"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"
)

// Heavily based on https://github.com/shurcooL/githubv4/blob/master/githubv4_test.go#L114-L144

const nodeMatchTmpl = `{
  "data": {
    "node": {
      "id": "{{.Provided}}"
    }
  }
}`
const nodeNoMatchTmpl = `{
  "data": {
    "node": null
  },
  "errors": [
    {
      "type": "NOT_FOUND",
      "path": [
        "node"
      ],
      "locations": [
        {
          "line": 2,
          "column": 3
        }
      ],
      "message": "Could not resolve to a node with the global id of '{{.Provided}}'"
    }
  ]
}`

const repoMatchTmpl = `{
  "data": {
    "repository": {
      "id": "{{.Expected}}"
    }
  }
}`

const repoNoMatchTmpl = `{
  "data": {
    "repository": null
  },
  "errors": [
    {
      "type": "NOT_FOUND",
      "path": [
        "repository"
      ],
      "locations": [
        {
          "line": 1,
          "column": 36
        }
      ],
      "message": "Could not resolve to a Repository with the name '{{.Owner}}/{{.Provided}}'."
    }
  ]
}`

func TestGetRepositoryIDPositiveMatches(t *testing.T) {
	cases := []struct {
		Provided string
		Expected string
		Owner    string
	}{
		{
			// Old style Node ID
			Provided: "MDEwOlJlcG9zaXRvcnk5MzQ0NjA5OQ==",
			Expected: "MDEwOlJlcG9zaXRvcnk5MzQ0NjA5OQ==",
		},
		{
			// Resolve a new style Node ID
			Provided: "terraform-provider-github",
			Expected: "MDEwOlJlcG9zaXRvcnk5MzQ0NjA5OQ==",
			Owner:    "integrations",
		},
		{
			// New style Node ID
			Provided: "R_kgDOGGmaaw",
			Expected: "R_kgDOGGmaaw",
		},
		{
			// Resolve a new style Node ID
			Provided: "actions-docker-build",
			Expected: "R_kgDOGGmaaw",
			Owner:    "hashicorp",
		},

		// Ensure We don't get any unexpected results
		{
			Provided: "testrepo8",
			Owner:    "testowner",
		},
		{
			Provided: "R_NOPE",
		},
		{
			Provided: "RkFJTCBIRVJFCg==",
		},
	}

	responses := template.Must(template.New("node_match").Parse(nodeMatchTmpl))
	_, err := responses.New("node_no_match").Parse(nodeNoMatchTmpl)
	if err != nil {
		panic(err)
	}
	_, err = responses.New("repo_match").Parse(repoMatchTmpl)
	if err != nil {
		panic(err)
	}
	_, err = responses.New("repo_no_match").Parse(repoNoMatchTmpl)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/graphql", func(w http.ResponseWriter, req *http.Request) {
		body := mustRead(req.Body)
		var action string
		for _, tc := range cases {
			if strings.Contains(body, tc.Provided) || strings.Contains(body, tc.Expected) {
				var out bytes.Buffer
				w.Header().Set("Content-Type", "application/json")
				if strings.Contains(body, "repository(owner:$owner, name:$name)") {
					if tc.Expected == tc.Provided {
						t.Fatalf("Attempted to use node_id=%s as a repo name", tc.Provided)
					} else if tc.Expected == "" {
						action = "repo_no_match"
					} else {
						action = "repo_match"
					}
				} else if strings.Contains(body, "node(id:$id)") {
					if tc.Expected == tc.Provided {
						action = "node_match"
					} else {
						action = "node_no_match"
					}
				} else {
					t.Fatalf("Unknown GraphQL Call on %s", tc.Provided)
				}
				err := responses.ExecuteTemplate(&out, action, tc)
				if err != nil {
					t.Fatalf("Failed Templating %s", err)
				}
				payload := out.String()
				mustWrite(w, payload)
				break
			}
		}
		if action == "" {
			t.Fatalf("Unknown query %s", body)
		}
	})

	meta := Owner{
		v4client: githubv4.NewClient(&http.Client{Transport: localRoundTripper{handler: mux}}),
		name:     "care-dot-com",
	}

	for _, tc := range cases {
		got, err := getRepositoryID(tc.Provided, &meta)
		if err != nil {
			// We expect to error out on these repos
			if tc.Expected != "" {
				t.Fatalf("unexpected error(s): %s (%s)", err, tc.Provided)
			}
			t.Logf("Got expected error in %s: %s", tc.Provided, err)
		}
		if (tc.Expected != "") && (tc.Expected != got) {
			t.Fatalf("%s got %s expected %s", tc.Provided, got, tc.Expected)
		}
		if (tc.Expected == "") && (got != nil) {
			t.Fatalf("%s should have failed, instead got  %s", tc.Provided, got)
		}
	}
}

// localRoundTripper is an http.RoundTripper that executes HTTP transactions
// by using handler directly, instead of going over an HTTP connection.
type localRoundTripper struct {
	handler http.Handler
}

func (l localRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	l.handler.ServeHTTP(w, req)
	return w.Result(), nil
}

func mustRead(r io.Reader) string {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func mustWrite(w io.Writer, s string) {
	_, err := io.WriteString(w, s)
	if err != nil {
		panic(err)
	}
}
