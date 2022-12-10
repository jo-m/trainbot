package vid

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"math/big"
	"path/filepath"
	"sort"
	"time"

	"github.com/aamcrae/webcam"
)

// Skip that many frames immediately after opening the camera.
const skipInitialFrames = 10

// FourCC is a FourCC pixel format string.
type FourCC string

const (
	// Motion-JPEG.
	FourCCMJPEG FourCC = "MJPG"
	// YUYV 4:2:2
	FourCCYUYV FourCC = "YUYV"
)

// CamConfig describes an available camera device with a given pixel format and frame size.
type CamConfig struct {
	// For example /dev/video0.
	DeviceFile string

	// Format is the image format FourCC to request from the camera, for example "MJPG".
	// To list available formats and frame sizes:
	//
	//   v4l2-ctl --list-formats-ext --device /dev/video2
	Format    FourCC
	FrameSize image.Point
}

// CamSrc is a video frame source which supports video4linux.
type CamSrc struct {
	c            CamConfig
	cam          *webcam.Webcam
	stride, size uint32
	fpsGuessed   float64
}

// compile time interface check
var _ Src = (*CamSrc)(nil)

// fourCCToStr converts a FourCC code to string, e.g. 1448695129 to YUYV.
func fourCCToStr(f webcam.PixelFormat) (FourCC, error) {
	i := big.NewInt(int64(uint32(f)))
	b := i.Bytes()

	if len(b) != 4 {
		return "", fmt.Errorf("unable to convert '%d' to a FourCC string", uint32(f))
	}

	for i := 0; i < len(b)/2; i++ {
		b[i], b[len(b)-i-1] = b[len(b)-i-1], b[i]
	}

	return FourCC(string(b)), nil
}

func checkFormatAvailable(cam *webcam.Webcam, c CamConfig) (webcam.PixelFormat, error) {
	fmap := cam.GetSupportedFormats()
	for f := range fmap {
		fourCC, err := fourCCToStr(f)
		if err != nil {
			return 0, err
		}

		if fourCC == c.Format {
			fsizes := cam.GetSupportedFrameSizes(f)
			for _, sz := range fsizes {
				if sz.MaxHeight == uint32(c.FrameSize.Y) && sz.MinHeight == uint32(c.FrameSize.Y) &&
					sz.MaxWidth == uint32(c.FrameSize.X) && sz.MinWidth == uint32(c.FrameSize.X) {
					return f, nil
				}
			}
		}
	}

	return 0, errors.New("unable to find the requested format and/or frame size on the given device")
}

// NewCamSrc tries to open the specified frame source for frame reading.
// It tries to guess the frame rate.
func NewCamSrc(c CamConfig) (ret *CamSrc, err error) {
	cam, err := webcam.Open(c.DeviceFile)
	if err != nil {
		return nil, err
	}

	format, err := checkFormatAvailable(cam, c)
	if err != nil {
		cam.Close()
		return nil, err
	}

	f, w, h, stride, size, err := cam.SetImageFormat(format, uint32(c.FrameSize.X), uint32(c.FrameSize.Y))
	if err != nil {
		cam.Close()
		return nil, err
	}
	if f != format || w != uint32(c.FrameSize.X) || h != uint32(c.FrameSize.Y) {
		cam.Close()
		return nil, errors.New("was not able to set the desired format and/or frame size")
	}

	err = cam.StartStreaming()
	if err != nil {
		cam.Close()
		return nil, err
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
			return nil, err
		}
	}
	ret.fpsGuessed = float64(skipInitialFrames) / time.Since(t0).Seconds()

	return ret, nil
}

// GetFPS implements Src.
func (s *CamSrc) GetFPS() float64 {
	return s.fpsGuessed
}

// getFrame retrieves a raw frame buffer from the camera.
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

// convertFrame tries to decode a raw frame from the camera specified image format.
func (s *CamSrc) convertFrame(frame []byte) (image.Image, error) {
	switch s.c.Format {
	case FourCCMJPEG:
		b := bytes.NewBuffer(frame)
		return jpeg.Decode(b)
	case FourCCYUYV:
		rect := image.Rectangle{image.Point{}, s.c.FrameSize}
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

// GetFrame implements Src.
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

// GetFrameRaw returns a raw frame in the specified pixel format from the camera.
func (s *CamSrc) GetFrameRaw() ([]byte, FourCC, *time.Time, error) {
	frame, ts, err := s.getFrame()
	if err != nil {
		return nil, "", nil, err
	}

	return frame, s.c.Format, ts, nil
}

// Close implements Src.
func (s *CamSrc) Close() error {
	return s.cam.Close()
}

// IsLive implements Src.
func (s *CamSrc) IsLive() bool {
	return true
}

func probeCam(deviceFile string) ([]CamConfig, error) {
	cam, err := webcam.Open(deviceFile)
	if err != nil {
		return nil, fmt.Errorf("unable to open camera device '%s': %w", deviceFile, err)
	}
	defer cam.Close()

	ret := []CamConfig{}
	for f := range cam.GetSupportedFormats() {
		for _, sz := range cam.GetSupportedFrameSizes(f) {
			// Do not support variable sized frames.
			if sz.MinWidth != sz.MaxWidth || sz.MinHeight != sz.MaxHeight {
				continue
			}

			fcc, err := fourCCToStr(f)
			if err != nil {
				return nil, err
			}

			ret = append(ret, CamConfig{
				DeviceFile: deviceFile,
				Format:     fcc,
				FrameSize:  image.Pt(int(sz.MaxWidth), int(sz.MaxHeight)),
			})
		}
	}

	return ret, nil
}

// DetectCams returns a list of detected cameras and their supported pixel formats and frame sizes.
// This works even if some of the devices are currently in use.
// Cameras which list no available pixel formats are ignored.
// Only fixed frame sizes are included.
func DetectCams() ([]CamConfig, error) {
	devices, err := filepath.Glob("/dev/video*")
	if err != nil {
		return nil, err
	}

	ret := []CamConfig{}
	for _, f := range devices {
		configs, err := probeCam(f)
		if err != nil {
			return nil, fmt.Errorf("failed to probe camera '%s': %w", f, err)
		}
		ret = append(ret, configs...)
	}

	sort.Slice(ret, func(i, j int) bool {
		a, b := ret[i], ret[j]

		// Sort by device file name.
		if a.DeviceFile != b.DeviceFile {
			return a.DeviceFile < b.DeviceFile
		}

		// Prefer MJPEG.
		if a.Format != b.Format {
			return a.Format == FourCCMJPEG
		}

		return a.FrameSize.X*a.FrameSize.Y >= b.FrameSize.X*b.FrameSize.Y
	})

	return ret, nil
}
