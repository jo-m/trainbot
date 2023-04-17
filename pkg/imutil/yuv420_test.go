package imutil

import (
	"os"
	"testing"

	"github.com/jo-m/trainbot/internal/pkg/testutil"

	"github.com/stretchr/testify/require"
)

const (
	w = 512
	h = 256
)

func Test_NewYuv420(t *testing.T) {
	buf, err := os.ReadFile("testdata/512x256.yuv420p.data")
	require.NoError(t, err)
	require.Equal(t, len(buf), w*h*12/8)

	im := NewYuv420(buf, w, h)
	truth, err := Load("testdata/512x256.yuv420p.jpg")
	require.NoError(t, err)
	testutil.AssertImagesAlmostEqual(t, truth, im)
}
