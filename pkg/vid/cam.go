package vid

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"math/big"
	"time"

	"github.com/aamcrae/webcam"
)

// Skip that many frames immediately after opening the camera.
const skipInitialFrames = 10

type CamConfig struct {
	DeviceFile string `arg:"env:PIC_DEV,--pic-dev" default:"/dev/video3" help:"camera video device file path" placeholder:"DEV"`

	// Format is the image format FourCC to request from the camera, for example "MJPG".
	// To list available formats and frame sizes:
	//
	//   v4l2-ctl --list-formats-ext --device /dev/video2
	FormatFourCC           string
	FrameSizeX, FrameSizeY uint32
}

type CamSrc struct {
	c            CamConfig
	cam          *webcam.Webcam
	stride, size uint32
}

// compile time interface check
var _ Src = (*CamSrc)(nil)

// fourCCToStr converts a FourCC code to string, e.g. 1448695129 to YUYV.
func fourCCToStr(f webcam.PixelFormat) (string, error) {
	i := big.NewInt(int64(uint32(f)))
	b := i.Bytes()

	if len(b) != 4 {
		return "", fmt.Errorf("unable to convert '%d' to a FourCC string", uint32(f))
	}

	for i := 0; i < len(b)/2; i++ {
		b[i], b[len(b)-i-1] = b[len(b)-i-1], b[i]
	}

	return string(b), nil
}

func checkFormatAvailable(cam *webcam.Webcam, c CamConfig) (webcam.PixelFormat, error) {
	fmap := cam.GetSupportedFormats()
	for f := range fmap {
		fourCC, err := fourCCToStr(f)
		if err != nil {
			return 0, err
		}

		if fourCC == c.FormatFourCC {
			fsizes := cam.GetSupportedFrameSizes(f)
			for _, sz := range fsizes {
				if sz.MaxHeight == c.FrameSizeY && sz.MinHeight == c.FrameSizeY &&
					sz.MaxWidth == c.FrameSizeX && sz.MinWidth == c.FrameSizeX {
					return f, nil
				}
			}
		}
	}

	return 0, errors.New("unable to find the requested format and/or frame size on the given device")
}

func NewCamSrc(c CamConfig) (ret *CamSrc, fps float64, err error) {
	cam, err := webcam.Open(c.DeviceFile)
	if err != nil {
		return nil, 0, err
	}

	format, err := checkFormatAvailable(cam, c)
	if err != nil {
		cam.Close()
		return nil, 0, err
	}

	f, w, h, stride, size, err := cam.SetImageFormat(format, uint32(c.FrameSizeX), uint32(c.FrameSizeY))
	if err != nil {
		cam.Close()
		return nil, 0, err
	}
	if f != format || w != c.FrameSizeX || h != c.FrameSizeY {
		cam.Close()
		return nil, 0, errors.New("was not able to set the desired format and/or frame size")
	}

	err = cam.StartStreaming()
	if err != nil {
		cam.Close()
		return nil, 0, err
	}

	ret = &CamSrc{
		c:      c,
		cam:    cam,
		stride: stride,
		size:   size,
	}

	// We now skip some initial frames, because
	// 1. We need some (mediocre) way to estimate FPS.
	// 2. Some cameras will return garbage in the first frame(s), so let's skip over that.
	// Initially, we retrieve a frame without measuring time because the camera takes a bit to spin up.
	ret.getFrame()
	t0 := time.Now()
	for i := 0; i < skipInitialFrames; i++ {
		_, _, err := ret.getFrame()
		if err != nil {
			cam.Close()
			return nil, 0, err
		}
	}
	fps = float64(skipInitialFrames) / time.Since(t0).Seconds()

	return ret, fps, nil
}

func (s *CamSrc) getFrame() ([]byte, *time.Time, error) {
	err := s.cam.WaitForFrame(uint32(time.Second))
	if err != nil {
		return nil, nil, err
	}
	ts := time.Now()

	frame, err := s.cam.ReadFrame()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read frame: %w", err)
	}
	if len(frame) == 0 {
		return nil, nil, errors.New("received empty frame")
	}

	return frame, &ts, nil
}

func (s *CamSrc) convertFrame(frame []byte) (image.Image, error) {
	switch s.c.FormatFourCC {
	case "MJPG":
		b := bytes.NewBuffer(frame)
		return jpeg.Decode(b)
	case "YUYV":
		rect := image.Rect(0, 0, int(s.c.FrameSizeX), int(s.c.FrameSizeY))
		buf := make([]byte, len(frame))
		copy(buf, frame)
		return &YCbCr{
			rect: rect,
			buf:  buf,
		}, nil
	default:
		return nil, errors.New("unsupported format")
	}
}

func (s *CamSrc) GetFrame() (image.Image, *time.Time, error) {
	frame, ts, err := s.getFrame()
	if err != nil {
		return nil, nil, err
	}

	img, err := s.convertFrame(frame)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to convert frame: %w", err)
	}

	return img, ts, nil
}

func (s *CamSrc) Close() error {
	return s.cam.Close()
}
