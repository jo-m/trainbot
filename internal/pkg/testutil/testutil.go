package testutil

import (
	"image"
	"testing"

	"github.com/jo-m/trainbot/internal/pkg/imutil"
	"github.com/stretchr/testify/require"
)

func LoadImg(t *testing.T, path string) image.Image {
	img, err := imutil.Load(path)
	require.NoError(t, err)

	return img
}
