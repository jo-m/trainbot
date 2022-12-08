package vid

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type FileSrc struct {
	reader  *io.PipeReader
	writer  *io.PipeWriter
	w, h    int
	buf     []byte
	startTs time.Time
	fps     float64
	count   uint64

	verbose bool

	ffmpegErr  error
	ffmpegLock sync.Mutex
}

// compile time interface check
var _ Src = (*FileSrc)(nil)

func parseFPS(fps string) (float64, error) {
	s := strings.SplitN(fps, "/", 2)
	if len(s) != 2 {
		return 0, errors.New("invalid FPS string")
	}

	a, err := strconv.ParseFloat(s[0], 64)
	if err != nil {
		return 0, err
	}

	b, err := strconv.ParseFloat(s[1], 64)
	if err != nil {
		return 0, err
	}

	return a / b, nil
}

func NewFileSrc(path string, verbose bool) (src *FileSrc, err error) {
	_, vidProbe, err := Probe(path)
	if err != nil {
		return nil, err
	}

	fps, err := parseFPS(vidProbe.RFrameRate)
	if err != nil {
		return nil, fmt.Errorf("unable to parse fps '%s': %w", vidProbe.RFrameRate, err)
	}

	reader, writer := io.Pipe()

	sz := vidProbe.Width * vidProbe.Height * 4 // TODO: this (4) depends on pixel format
	buf := make([]byte, sz)

	s := FileSrc{
		reader:  reader,
		writer:  writer,
		w:       vidProbe.Width,
		h:       vidProbe.Height,
		buf:     buf,
		startTs: vidProbe.Tags.CreationTime,
		fps:     fps,
		count:   0,

		verbose: verbose,
	}

	go s.run(path)

	return &s, nil
}

func (s *FileSrc) run(path string) {
	defer s.writer.Close()

	input := ffmpeg.Input(path).
		Output("pipe:",
			ffmpeg.KwArgs{
				// TODO: what about pixel format?
				"format": "rawvideo", "pix_fmt": "rgba",
			}).
		WithOutput(s.writer)
	if s.verbose {
		logReader, logWriter := io.Pipe()
		input.WithErrorOutput(logWriter)

		go func() {
			defer logReader.Close()
			defer logWriter.Close()

			input := bufio.NewReaderSize(logReader, 1024)
			for {
				line, _, err := input.ReadLine()
				if err != nil {
					log.Info().Err(err).Msg("ffmpeg stderr reader terminated")
				}

				log.Info().Str("line", string(line)).Msg("ffmpeg output")
			}
		}()
	}

	err := input.Run()
	if err != nil {
		s.ffmpegLock.Lock()
		s.ffmpegErr = err
		s.ffmpegLock.Unlock()
	}
}

func (s *FileSrc) GetFrame() (image.Image, *time.Time, error) {
	s.ffmpegLock.Lock()
	err := s.ffmpegErr
	s.ffmpegLock.Unlock()

	if err != nil {
		return nil, nil, err
	}

	n, err := io.ReadFull(s.reader, s.buf)
	if n == 0 || err == io.EOF {
		return nil, nil, io.EOF
	}

	ts := s.startTs.Add(time.Second * time.Duration(s.count) / time.Duration(s.fps))
	s.count++

	rect := image.Rectangle{Max: image.Point{X: s.w, Y: s.h}}
	return &image.RGBA{
		Pix:    s.buf,
		Stride: 4 * rect.Dx(),
		Rect:   rect,
	}, &ts, nil
}

// GetFPS implements Src.
func (s *FileSrc) GetFPS() float64 {
	return s.fps
}

func (s *FileSrc) Close() error {
	return s.writer.Close()
}
