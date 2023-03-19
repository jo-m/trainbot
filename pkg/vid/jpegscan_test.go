package vid

import (
	"bytes"
	"image/jpeg"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testImage = "testdata/frame.jpg"

func Test_ScanJPEG(t *testing.T) {
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
