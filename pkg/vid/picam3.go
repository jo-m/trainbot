package vid

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/rs/zerolog/log"
)

// TODO: list/probe cameras
// TODO: support MJPEG

// Hardcoded values for the Raspberry Pi Camera Module v3.
const (
	sensorH = 2592 // 2^5 × 3^4
	sensorW = 4608 // 2^9 x 3^2
	fps     = 30
)

type PiCam3Config struct {
	// ROI to extract.
	Rect image.Rectangle
	// Constant lens focus, 0=infinity, 2=approx. 0.5m.
	Focus float64
	// Rotate image by 180 degree if true.
	Rotate180 bool
	// Pixel format.
	Format FourCC
}

// PiCam3Src is a video frame source which reads frames from a Raspberry PI 3 camera module.
// It uses the `libcamera-vid` utility internally.
// Use NewPiCam3Src to open one.
type PiCam3Src struct {
	c                PiCam3Config
	proc             *exec.Cmd
	outPipe, errPipe io.ReadCloser
	buf              []byte // Raw yuv420 bytes.
}

// Compile time interface check.
var _ Src = (*PiCam3Src)(nil)

func NewPiCam3Src(c PiCam3Config) (*PiCam3Src, error) {
	if c.Rect.Max.X > sensorW || c.Rect.Max.Y > sensorH {
		return nil, errors.New("rect too large/out of bounds")
	}
	if c.Rect.Min.X < 0 || c.Rect.Min.Y < 0 {
		return nil, errors.New("rect too small/out of bounds")
	}
	if c.Rect.Dx()%2 != 0 || c.Rect.Dy()%2 != 0 {
		return nil, errors.New("rect bounds must be even")
	}
	if c.Format != FourCCYUV420 {
		return nil, errors.New("only yuv420p is supported")
	}

	sx := float64(sensorW)
	sy := float64(sensorH)
	roi := fmt.Sprintf("%f,%f,%f,%f", float64(c.Rect.Min.X)/sx, float64(c.Rect.Min.Y)/sy, float64(c.Rect.Dx())/sx, float64(c.Rect.Dy())/sy)

	args := []string{
		"--verbose=0",
		"-t", "0",
		"--inline",
		"--nopreview",
		"--codec", "yuv420",
		"--width", fmt.Sprint(c.Rect.Dx()),
		"--height", fmt.Sprint(c.Rect.Dy()),
		"--roi", roi,

		"--autofocus-mode=manual",
		fmt.Sprintf("--lens-position=%f", c.Focus),
		"--framerate", fmt.Sprint(fps),

		"-o", "-",
	}
	if c.Rotate180 {
		args = append(args, "--rotation=180")
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

	bufSz := c.Rect.Dx() * c.Rect.Dy() * 12 / 8
	ret := &PiCam3Src{
		c:       c,
		proc:    cmd,
		outPipe: outPipe,
		errPipe: errPipe,
		buf:     make([]byte, bufSz),
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

func (s *PiCam3Src) readFrame() error {
	n, err := io.ReadFull(s.outPipe, s.buf)
	if err != nil {
		return err
	}
	if n != len(s.buf) {
		return fmt.Errorf("read %d bytes for frame but should have read %d", n, len(s.buf))
	}

	return nil
}

// GetFrame retrieves the next frame.
// Note that the underlying image buffer remains owned by the video source,
// it must not be changed by the caller and might be overwritten on the next
// invocation.
// Returns io.EOF after the last frame, after which Close() should be called
// on the instance before discarding it.
func (s *PiCam3Src) GetFrame() (image.Image, *time.Time, error) {
	err := s.readFrame()
	if err != nil {
		return nil, nil, err
	}

	ts := time.Now()
	frame := NewYuv420(s.buf, s.c.Rect.Dx(), s.c.Rect.Dy())
	return frame, &ts, nil
}

// GetFrame retrieves the next frame in the raw pixel format of the source.
// Not all sources might implement this.
func (s *PiCam3Src) GetFrameRaw() ([]byte, FourCC, *time.Time, error) {
	panic("not implemented")
}

// IsLive returns if the src is a live source (e.g. camera).
func (s *PiCam3Src) IsLive() bool {
	return true
}

// GetFPS returns the current frame rate of this source.
func (s *PiCam3Src) GetFPS() float64 {
	return float64(fps)
}

// Close closes the frame source and frees resources.
func (s *PiCam3Src) Close() error {
	s.proc.Process.Signal(os.Kill)
	return s.proc.Wait()
}
