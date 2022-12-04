package vid

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"sync"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func Probe(path string) (fileProbe *FFProbeJSON, vidProbe *FFStream, err error) {
	data, err := ffmpeg.Probe(path)
	if err != nil {
		return nil, nil, err
	}

	fileProbe = &FFProbeJSON{}
	err = json.Unmarshal([]byte(data), fileProbe)
	if err != nil {
		return nil, nil, err
	}

	c := 0
	var stream FFStream
	for _, s := range fileProbe.Streams {
		if s.CodecType == "video" {
			c++
			stream = s
		}
	}
	if c == 0 {
		return nil, nil, errors.New("no video stream found in file")
	}
	if c > 1 {
		return nil, nil, errors.New("more than one video stream found in file")
	}

	return fileProbe, &stream, nil
}

func ProbeSize(path string) (w, h int, err error) {
	_, vidProbe, err := Probe(path)
	if err != nil {
		return 0, 0, err
	}

	return vidProbe.Width, vidProbe.Height, nil
}

type Source struct {
	reader   *io.PipeReader
	writer   *io.PipeWriter
	w, h     int
	pxFormat string
	buf      []byte

	ffmpegErr  error
	ffmpegLock sync.Mutex
}

func NewSource(path string) (*Source, error) {
	_, vidProbe, err := Probe(path)
	if err != nil {
		return nil, err
	}

	pxFormat := vidProbe.PixFmt
	if pxFormat != "yuvj420p" {
		return nil, fmt.Errorf("unsupported pix_fmt %s", pxFormat)
	}

	reader, writer := io.Pipe()

	sz := vidProbe.Width * vidProbe.Height * 4
	buf := make([]byte, sz)

	s := Source{
		reader:   reader,
		writer:   writer,
		w:        vidProbe.Width,
		h:        vidProbe.Height,
		pxFormat: pxFormat,
		buf:      buf,
	}

	go s.run(path)

	return &s, nil
}

func (s *Source) run(path string) {
	defer s.writer.Close()

	err := ffmpeg.Input(path).
		Output("pipe:",
			ffmpeg.KwArgs{
				// TODO: what about pixel format?
				"format": "rawvideo", "pix_fmt": "rgba",
			}).
		WithOutput(s.writer).
		ErrorToStdOut(). // TODO: remove
		Run()

	if err != nil {
		s.ffmpegLock.Lock()
		s.ffmpegErr = err
		s.ffmpegLock.Unlock()
	}
}

// GetFrame retrieves a frame from the video.
// Note that the underlying image buffer is owned by the video source,
// it must not be changed by the caller and will be overwritten on the next
// invocation.
func (s *Source) GetFrame() (*image.RGBA, error) {
	s.ffmpegLock.Lock()
	err := s.ffmpegErr
	s.ffmpegLock.Unlock()

	if err != nil {
		return nil, err
	}

	n, err := io.ReadFull(s.reader, s.buf)
	if n == 0 || err == io.EOF {
		return nil, io.EOF
	}

	rect := image.Rectangle{Max: image.Point{X: s.w, Y: s.h}}
	return &image.RGBA{
		Pix:    s.buf,
		Stride: 4 * rect.Dx(),
		Rect:   rect,
	}, nil
}

func (s *Source) Close() error {
	return s.writer.Close()
}
