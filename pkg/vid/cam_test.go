package vid

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vladimirvivien/go4vl/v4l2"
)

func Test_FourCC_String(t *testing.T) {
	s, err := FourCC(v4l2.PixelFmtMJPEG).String()
	assert.NoError(t, err)
	assert.Equal(t, "MJPG", s)
}

func Test_FourCCFromString(t *testing.T) {
	f := FourCCFromString("MJPG")
	assert.Equal(t, FourCC(v4l2.PixelFmtMJPEG), f)
}
