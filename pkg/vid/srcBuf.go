package vid

import (
	"image"
	"io"
	"time"

	"github.com/jo-m/trainbot/internal/pkg/imutil"
	"github.com/rs/zerolog/log"
)

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

// NewSrcBuf creates a new SrcBuf.
// Will not close src, caller needs to do that after last frame is read.
func NewSrcBuf(src Src, maxFailedFrames int) *SrcBuf {
	ret := SrcBuf{
		src:             src,
		maxFailedFrames: maxFailedFrames,
		queue:           make(chan frameWithTS, 1),
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
		}

		failedFrames = 0

		// Create copy.
		frame = imutil.ToRGBA(frame)

		if live {
			select {
			case s.queue <- frameWithTS{frame, *ts}:
			default:
				log.Warn().Msg("dropped frame")
			}
		} else {
			s.queue <- frameWithTS{frame, *ts}
		}
	}
}

// GetFrame returns the next frame.
// As soon as this returns an error once, the instance needs to be discarded.
func (s *SrcBuf) GetFrame() (image.Image, *time.Time, error) {
	f, ok := <-s.queue
	if ok {
		return f.frame, &f.ts, nil
	}

	return nil, nil, <-s.err
}
