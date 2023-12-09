package vid

import (
	"image"
	"io"
	"time"

	"github.com/jo-m/trainbot/internal/pkg/prometheus"
	"github.com/jo-m/trainbot/pkg/imutil"
	"github.com/rs/zerolog/log"
)

const queueSize = 200

type frameWithTS struct {
	frame image.Image
	ts    time.Time
}

// SrcBuf buffers a video source.
// Use NewSrcBuf to create an instance.
type SrcBuf struct {
	src             Src
	maxFailedFrames int
	queue           chan frameWithTS
	err             chan error
}

// Compile time interface check.
var _ Src = (*SrcBuf)(nil)

// NewSrcBuf creates a new SrcBuf.
// Will not close src, caller needs to do that after last frame is read.
func NewSrcBuf(src Src, maxFailedFrames int) *SrcBuf {
	ret := SrcBuf{
		src:             src,
		maxFailedFrames: maxFailedFrames,
		queue:           make(chan frameWithTS, queueSize),
		err:             make(chan error),
	}

	go ret.run()

	return &ret
}

func (s *SrcBuf) cleanup(err error) {
	close(s.queue)
	s.err <- err
	close(s.err)
}

func (s *SrcBuf) run() {
	live := s.src.IsLive()
	failedFrames := 0

	for {
		frame, ts, err := s.src.GetFrame()
		if err != nil {
			failedFrames++
			log.Warn().Err(err).Int("failedFrames", failedFrames).Msg("failed to retrieve frame")

			if err == io.EOF {
				s.cleanup(err)
				return
			}

			if failedFrames >= s.maxFailedFrames {
				log.Error().Msg("retrieving frames failed too many times, exiting")
				s.cleanup(err)
				return
			}

			continue
		}

		failedFrames = 0

		// Create copy.
		frame = imutil.Copy(frame)

		if live {
			select {
			case s.queue <- frameWithTS{frame, *ts}:
			default:
				log.Warn().Msg("dropped frame")
				prometheus.RecordFrameDisposition("dropped")
			}
		} else {
			s.queue <- frameWithTS{frame, *ts}
		}
	}
}

// GetFrame returns the next frame.
// As soon as this returns an error once, the instance needs to be discarded.
// The underlying image buffer will be owned by the caller, src will not reuse or modify it.
func (s *SrcBuf) GetFrame() (image.Image, *time.Time, error) {
	f, ok := <-s.queue
	if ok {
		return f.frame, &f.ts, nil
	}

	return nil, nil, <-s.err
}

// GetFPS implements Src.
func (s *SrcBuf) GetFPS() float64 {
	return s.src.GetFPS()
}

// IsLive implements Src.
func (s *SrcBuf) IsLive() bool {
	return s.src.IsLive()
}

// Close implements Src.
func (s *SrcBuf) Close() error {
	panic("do not call this, instead close the underlying source yourself")
}

// GetFrameRaw implements Src.
func (s *SrcBuf) GetFrameRaw() ([]byte, FourCC, *time.Time, error) {
	panic("not implemented")
}
