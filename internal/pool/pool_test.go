package pool_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bynov/multiplexer/internal/pool"
)

func TestDo(t *testing.T) {
	var tests = map[string]struct {
		paths     []string
		responses []string
	}{
		"valid case": {
			paths: []string{
				"/one",
				"/two",
				"/three",
				"/four",
				"/five",
				"/six",
			},
			responses: []string{
				"111",
				"222",
				"333",
				"444",
				"555",
				"666",
			},
		},
		"empty case": {
			paths:     []string{},
			responses: []string{},
		},
	}

	for k, test := range tests {
		if len(test.paths) != len(test.responses) {
			t.Fatal("paths count is not equal to responses count", k)
		}

		mux := http.NewServeMux()

		// Register different responses on different paths
		for i, path := range test.paths {
			mux.HandleFunc(path, func(resp string) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte(resp))
				}
			}(test.responses[i]))
		}

		srv := httptest.NewServer(mux)

		// Build urls
		var urls = make([]string, 0, len(test.paths))
		for _, path := range test.paths {
			urls = append(urls, srv.URL+path)
		}

		response, err := pool.New(4).Do(context.Background(), urls)
		if err != nil {
			t.Fatal("err", err, k)
		}

		srv.Close()

		assert.Equal(t, test.responses, response)
	}
}
