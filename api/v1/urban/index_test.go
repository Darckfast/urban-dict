package urban_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomTermSearch(t *testing.T) {
	assert := assert.New(t)

	res, err := http.Get("http://localhost:3000/api/urban?term=&channel=test")

	assert.Nil(err)

	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	stringConent := string(body)

	assert.Equal(res.StatusCode, 200, "Should return 200")
	assert.NotEmpty(stringConent, "Should not be empty")
}

func TestTermSearch(t *testing.T) {
	assert := assert.New(t)

	res, err := http.Get("http://localhost:3000/api/urban?term=glizzy&channel=test")

	assert.Nil(err)

	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	stringConent := string(body)

	assert.Equal(res.StatusCode, 200, "Should return 200")
	assert.NotEmpty(stringConent, "Should not be empty")
}

func TestNoTermSearch(t *testing.T) {
	assert := assert.New(t)

	res, err := http.Get("http://localhost:3000/api/urban")

	assert.Nil(err)

	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	stringConent := string(body)

	assert.Equal(res.StatusCode, 200, "Should return 200")
	assert.NotEmpty(stringConent, "Should not be empty")
}
