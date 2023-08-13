package vid

import (
	"bytes"
	"errors"
	"image/jpeg"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testImage = "testdata/frame.jpg"

func Test_JPEGScanner_Scan(t *testing.T) {
	f, err := os.Open(testImage)
	require.NoError(t, err)
	defer f.Close()

	stat, err := os.Stat(testImage)
	require.NoError(t, err)

	scanner := NewJPEGScanner(f)
	data, err := scanner.Scan()
	assert.NoError(t, err)
	assert.Equal(t, int64(len(data)), stat.Size())

	buf := bytes.NewBuffer(data)
	_, err = jpeg.Decode(buf)
	assert.NoError(t, err)
}

func Test_JPEGScanner_Scan_Multiple(t *testing.T) {
	img, err := os.ReadFile(testImage)
	require.NoError(t, err)

	buf := bytes.Buffer{}
	for i := 0; i < 5; i++ {
		n, err := buf.Write(img)
		require.NoError(t, err)
		require.Len(t, img, n)
	}

	scanner := NewJPEGScanner(&buf)
	for i := 0; i < 5; i++ {
		data, err := scanner.Scan()
		assert.NoError(t, err)
		assert.Len(t, data, len(img))
	}

	data, err := scanner.Scan()
	assert.True(t, errors.Is(err, io.EOF))
	assert.Len(t, data, 0)
}
