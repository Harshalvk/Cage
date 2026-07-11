package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAPIKey(t *testing.T) {
	rawKey, keyHash, err := GenerateAPIKey()
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(rawKey, "cage_"), "key should have cage_ prefix")
	assert.NotEmpty(t, keyHash)
	assert.NotEqual(t, rawKey, keyHash, "hash should differ from raw key")
}

func TestHashKey_Deterministic(t *testing.T) {
	hash1 := HashKey("some-key")
	hash2 := HashKey("some-key")
	assert.Equal(t, hash1, hash2, "hashing the same key twice should prodcue the same hash")
}

func TestHashKey_DifferentInputDifferentHashes(t *testing.T) {
	hash1 := HashKey("key-a")
	hash2 := HashKey("key-b")
	assert.NotEqual(t, hash1, hash2)
}
