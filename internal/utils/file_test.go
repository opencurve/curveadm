package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVariantName(t *testing.T) {
	assert := assert.New(t)
	vname := NewVariantName("test")
	assert.Equal("test", vname.Name)
	assert.Equal("test.tar.gz", vname.CompressName)
	assert.Equal("test.local.tar.gz", vname.LocalCompressName)
	assert.Equal("test-encrypted.tar.gz", vname.EncryptCompressName)
}
