package bases

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmbedCRDs(t *testing.T) {
	entries, err := CRDsFS.ReadDir(".")
	assert.NoError(t, err)
	assert.Equal(t, 6, len(entries))
}
