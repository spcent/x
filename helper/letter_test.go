package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUpper(t *testing.T) {
	// Test IsUpper cases
	assert.True(t, IsLetterUpper('A'))
	assert.True(t, IsLetterUpper('Z'))
	assert.False(t, IsLetterUpper('a'))
	assert.False(t, IsLetterUpper('z'))
	assert.False(t, IsLetterUpper('0'))
	assert.False(t, IsLetterUpper('9'))
}

func TestIsLower(t *testing.T) {
	// Test IsLower cases
	assert.False(t, IsLetterLower('A'))
	assert.False(t, IsLetterLower('Z'))
	assert.True(t, IsLetterLower('a'))
	assert.True(t, IsLetterLower('z'))
	assert.False(t, IsLetterLower('0'))
}
