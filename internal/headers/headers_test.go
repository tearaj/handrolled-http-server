package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidHeaders(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
}

func TestValidSingleHeaderWithSpace(t *testing.T) {
	headers := NewHeaders()
	data := []byte("         Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 32, n)
	assert.False(t, done)
}

func TestValidMultipleHeader(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n Content-length: 1234\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	n2, done, err := headers.Parse(data[23:])
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "1234", headers["content-length"])
	assert.Equal(t, 23, n2)
	assert.False(t, done)

	n3, done, err := headers.Parse(data[46:])
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 2, n3)
	assert.True(t, done)
}

func TestValidDone(t *testing.T) {
	headers := NewHeaders()
	data := []byte("         Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 32, n)
	assert.False(t, done)
	n, done, err = headers.Parse(data[32:])
	assert.True(t, done, true)
}

func TestInvalidHeaderNameWithSpace(t *testing.T) {
	headers := NewHeaders()
	data := []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestInvalidHeaderNameWithUnsupportedCharacters(t *testing.T) {
	headers := NewHeaders()
	data := []byte("       Hos@t: localhost:42069       \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	require.ErrorContains(t, err, "header name contains unsupported characters")
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
