package vid

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/rs/zerolog/log"
)

// Hardcoded values for the Raspberry Pi Camera Module v3.
// TODO: select mode depending on rect size.
const (
	sensorH = 2592 // 2^5 × 3^4
	sensorW = 4608 // 2^9 x 3^2
)

// PiCam3Config is the configuration for a PiCam3Src.
type PiCam3Config struct {
	// ROI to extract. Defaults to full image if empty.
	Rect image.Rectangle
	// Constant lens focus, 0=infinity, 2=approx. 0.5m.
	Focus float64
	// Rotate image by 180 degree if true.
	Rotate180 bool
	// Pixel format.
	Format FourCC
	// Frames per second.
	FPS int
}

// PiCam3Src is a video frame source which reads frames from a Raspberry PI 3 camera module.
// It uses the `libcamera-vid` utility internally.
// Use NewPiCam3Src() to open one.
type PiCam3Src struct {
	c                PiCam3Config
	proc             *exec.Cmd
	outPipe, errPipe io.ReadCloser

	yuvBuf      []byte       // Raw yuv420 bytes, only used in yuv420p mode.
	jpegScanner *JPEGScanner // JPEG buf, only used in MJPEG mode.
}

// Compile time interface check.
var _ Src = (*PiCam3Src)(nil)

// NewPiCam3Src creates a new PiCam3Src.
func NewPiCam3Src(c PiCam3Config) (*PiCam3Src, error) {
	if c.Rect == image.Rect(0, 0, 0, 0) {
		c.Rect = image.Rect(0, 0, sensorW, sensorH)
	}

	if c.Rect.Max.X > sensorW || c.Rect.Max.Y > sensorH {
		return nil, errors.New("rect too large/out of bounds")
	}
	if c.Rect.Min.X < 0 || c.Rect.Min.Y < 0 {
		return nil, errors.New("rect too small/out of bounds")
	}
	if c.Rect.Min.X%2 != 0 || c.Rect.Min.Y%2 != 0 {
		return nil, errors.New("rect position must be even")
	}
	if c.Rect.Dx()%2 != 0 || c.Rect.Dy()%2 != 0 {
		return nil, errors.New("rect bounds must be even")
	}

	sx := float64(sensorW)
	sy := float64(sensorH)
	roi := fmt.Sprintf("%f,%f,%f,%f", float64(c.Rect.Min.X)/sx, float64(c.Rect.Min.Y)/sy, float64(c.Rect.Dx())/sx, float64(c.Rect.Dy())/sy)

	args := []string{
		"--verbose=1",
		"--timeout=0",
		"--inline",
		"--nopreview",
		"--width", fmt.Sprint(c.Rect.Dx()),
		"--height", fmt.Sprint(c.Rect.Dy()),
		"--roi", roi,
		"--mode=4056:3040:12:P",

		"--autofocus-mode=manual",
		fmt.Sprintf("--lens-position=%f", c.Focus),
		"--framerate", fmt.Sprint(c.FPS),

		"--output", "-",
	}
	if c.Rotate180 {
		args = append(args, "--rotation=180")
	}

	var bufSz int
	switch c.Format {
	case FourCCYUV420:
		args = append(args, "--codec=yuv420")
		bufSz = c.Rect.Dx() * c.Rect.Dy() * 12 / 8
	case FourCCMJPEG:
		args = append(args, "--codec=mjpeg")
		args = append(args, "--quality=90")
		bufSz = 0
	default:
		return nil, fmt.Errorf("unsupported image format '%s'", c.Format.String())
	}

	cmd := exec.Command("libcamera-vid", args...)

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	ret := &PiCam3Src{
		c:       c,
		proc:    cmd,
		outPipe: outPipe,
		errPipe: errPipe,

		yuvBuf:      make([]byte, bufSz),
		jpegScanner: NewJPEGScanner(outPipe),
	}

	go ret.processErr()

	return ret, nil
}

// processErr forwards stderr from libcamera-vid to the logging system.
func (s *PiCam3Src) processErr() {
	scanner := bufio.NewScanner(s.errPipe)
	for scanner.Scan() {
		line := scanner.Text()
		log.Info().Str("src", "stderr").Msg(line)
	}
}

func (s *PiCam3Src) readFrame() ([]byte, error) {
	switch s.c.Format {
	case FourCCYUV420:
		n, err := io.ReadFull(s.outPipe, s.yuvBuf)
		if err != nil {
			return nil, err
		}
		if n != len(s.yuvBuf) {
			return nil, fmt.Errorf("read %d bytes for frame but should have read %d", n, len(s.yuvBuf))
		}
		return s.yuvBuf, nil
	case FourCCMJPEG:
		return s.jpegScanner.Scan()
	default:
		panic("unsupported image format")
	}
}

// GetFrame implements Src.
func (s *PiCam3Src) GetFrame() (image.Image, *time.Time, error) {
	buf, err := s.readFrame()
	if err != nil {
		return nil, nil, err
	}

	ts := time.Now()
	switch s.c.Format {
	case FourCCYUV420:
		return NewYuv420(buf, s.c.Rect.Dx(), s.c.Rect.Dy()), &ts, nil
	case FourCCMJPEG:
		im, err := jpeg.Decode(bytes.NewBuffer(buf))
		if err != nil {
			return nil, nil, err
		}
		return im, &ts, nil
	default:
		panic("unsupported image format")
	}
}

// GetFrameRaw implements Src.
func (s *PiCam3Src) GetFrameRaw() ([]byte, FourCC, *time.Time, error) {
	buf, err := s.readFrame()
	if err != nil {
		return nil, 0, nil, err
	}
	ts := time.Now()

	return buf, s.c.Format, &ts, nil
}

// IsLive implements Src.
func (s *PiCam3Src) IsLive() bool {
	return true
}

// GetFPS implements Src.
func (s *PiCam3Src) GetFPS() float64 {
	return float64(s.c.FPS)
}

// Close implements Src.
func (s *PiCam3Src) Close() error {
	s.proc.Process.Signal(os.Kill)
	return s.proc.Wait()
}
