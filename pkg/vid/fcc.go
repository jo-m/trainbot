package vid

import (
	"math/big"

	"github.com/vladimirvivien/go4vl/v4l2"
)

// FourCC is a FourCC pixel format.
type FourCC int32

var (
	// FourCCMJPEG means Motion-JPEG.
	FourCCMJPEG FourCC = FourCC(v4l2.PixelFmtMJPEG)
	// FourCCYUYV means YUYV 4:2:2.
	FourCCYUYV FourCC = FourCC(v4l2.PixelFmtYUYV)
)

// String converts a FourCC code to string, e.g. 1448695129 to YUYV.
func (f FourCC) String() string {
	i := big.NewInt(int64(uint32(f)))
	b := i.Bytes()

	if len(b) != 4 {
		return ""
	}

	for i := 0; i < len(b)/2; i++ {
		b[i], b[len(b)-i-1] = b[len(b)-i-1], b[i]
	}

	return string(b)
}

func FourCCFromString(fcc string) FourCC {
	if len(fcc) != 4 {
		return 0
	}

	b := []byte(fcc)
	return FourCC(int32(b[0]) + int32(b[1])<<8 + int32(b[2])<<16 + int32(b[3])<<24)
}
