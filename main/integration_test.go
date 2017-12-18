package main

import (
	"testing"

	"dinowernli.me/almanac/appender"
	st "dinowernli.me/almanac/storage"

	"github.com/stretchr/testify/assert"
)

func TestFullSystem(t *testing.T) {
	storage := st.NewInMemoryStorage()

	entriesPerChunk := 3
	_, err := appender.New("appender0", storage, entriesPerChunk)
	assert.NoError(t, err)
}
