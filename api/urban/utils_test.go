package urban

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseUrlArgs(t *testing.T) {
	term, atUser := ParseUrlArgs("test%20test%20@username")

	assert.Equal(t, "test test", term)
	assert.Equal(t, "@username", atUser)
}
