package urban_test

import (
	"io"
	"net/http/httptest"
	"testing"

	"urban-dict/api/v1/urban"

	"github.com/stretchr/testify/assert"
)

func FuzzRandomTermSearch(f *testing.F) {
	testcases := []string{"test", "", "random"}
	for _, tc := range testcases {
		f.Add(tc) // Use f.Add to provide a seed corpus
	}

	f.Fuzz(func(t *testing.T, qParam string) {
		req := httptest.NewRequest("GET", "http://localhost:3000/api/v1/urban", nil)
		q := req.URL.Query()
		q.Add("term", qParam)

		req.URL.RawQuery = q.Encode()
		w := httptest.NewRecorder()
		urban.Handler(w, req)

		res := w.Result()
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)

		assert.Equal(t, res.Header.Get("content-type"), "text/plain")
		assert.Equal(t, res.StatusCode, 200, "Should return 200")
		assert.NotEmpty(t, body, "Should not be empty")
	})
}
