package stitch

import (
	"image"
	"math"
	"time"

	"github.com/jo-m/trainbot/pkg/avg"
	"github.com/jo-m/trainbot/pkg/imutil"
	"github.com/jo-m/trainbot/pkg/pmatch"
	"github.com/rs/zerolog/log"
)

const (
	goodScoreNoMove   = 0.99
	goodScoreMove     = 0.925
	maxSeqLen         = 1500
	minFramePeriodS   = 0.01
	dxLowPassFactor   = 0.95
	minContrastAvg    = 0.005
	minContrastAvgDev = 0.01
)

// Config is the configuration for a AutoStitcher.
// All values must be > 0, except for MinSpeedKPH which might also be 0.
type Config struct {
	PixelsPerM  float64
	MinSpeedKPH float64
	MaxSpeedKPH float64
	MinLengthM  float64
}

func (c *Config) minPxPerFrame(framePeriodS float64) int {
	if framePeriodS == 0 {
		return 1
	}
	fps := 1 / framePeriodS
	ret := int(c.MinSpeedKPH/3.6*c.PixelsPerM/fps) - 1
	if ret < 1 {
		return 1
	}
	return ret
}

func (c *Config) maxPxPerFrame(framePeriodS float64) int {
	if framePeriodS == 0 {
		return 1
	}
	fps := 1 / framePeriodS
	ret := int(c.MaxSpeedKPH/3.6*c.PixelsPerM/fps) + 1
	if ret < 1 {
		return 1
	}
	return ret
}

func (c *Config) minSpeedPxPS() float64 {
	return c.MinSpeedKPH / 3.6 * c.PixelsPerM
}

func (c *Config) minLengthPx() float64 {
	return c.MinLengthM * c.PixelsPerM
}

type sequence struct {
	// Timestamp of the frame before the first frame in the sequence (frames[-1]).
	// Must be a pointer, we cannot depend on the zero value to determine if this has not yet been set
	// (for example, a video file without metadata/known start time might have time.Time{} as first timestamp).
	startTS *time.Time

	// The slices must always have the same length.

	// frame[i] contains the i-th frame.
	// All frames must have the same image size.
	frames []image.Image
	// dx[x] is the pixel offset between frames[i-1] and frames[i].
	// Speed of a frame, in pixels/s is calculated as dx[i]/(ts[i] - ts[i-1]).
	// dx[0] must never be 0.
	dx []int
	// ts[i] is the timestamp of the i-th frame.
	ts []time.Time
}

// AutoStitcher is an automatic train detector and stitcher.
// Use NewAutoStitcher() to create an instance.
type AutoStitcher struct {
	c Config

	prevFrameIx uint64
	// Those are all together zero/nil or not.
	prevFrameTS    time.Time
	prevFrameColor image.Image
	prevFrameRGBA  *image.RGBA

	seq          sequence
	dxAbsLowPass float64
}

// NewAutoStitcher creates a new AutoStitcher.
func NewAutoStitcher(c Config) *AutoStitcher {
	return &AutoStitcher{
		c: c,
	}
}

func findOffset(prev, curr *image.RGBA, maxDx int) (dx int, score float64) {
	t0 := time.Now()
	defer func() {
		log.Trace().Dur("dur", time.Since(t0)).Msg("findOffset() duration")
	}()

	if prev.Rect.Size() != curr.Rect.Size() {
		log.Panic().Msg("inconsistent size, this should not happen")
	}

	// Centered crop from prev frame,
	// width is 3x max pixels per frame given by max velocity
	w := maxDx * 3
	if prev.Rect.Dx() < w {
		panic("frame width is too small")
	}
	// and height 1/2 of frame.
	h := int(float64(prev.Rect.Dy())*1/2 + 1)
	subRect := image.Rect(0, 0, w, h).
		Add(curr.Rect.Min).
		Add(
			curr.Rect.Size().
				Sub(image.Pt(int(w), h)).
				Div(2),
		)
	sub, err := imutil.Sub(prev, subRect)
	if err != nil {
		log.Panic().Err(err).Msg("this should not happen")
	}

	// Centered slice crop from next frame,
	// width is 1x max pixels per frame given by max velocity and same height as above.
	w = maxDx
	sliceRect := image.Rect(0, 0, w, h).
		Add(curr.Rect.Min).
		Add(
			curr.Rect.Size().
				Sub(image.Pt(w, h)).
				Div(2),
		)

	slice, err := imutil.Sub(curr, sliceRect)
	if err != nil {
		log.Panic().Err(err).Msg("this should not happen")
	}

	// We expect this x value to be found by the search if the frame has not moved.
	xZero := sliceRect.Min.Sub(subRect.Min).X

	x, _, score := pmatch.SearchRGBAC(sub.(*image.RGBA), slice.(*image.RGBA))
	return x - xZero, score
}

func (r *AutoStitcher) reset() {
	log.Trace().Msg("resetting sequence")

	r.seq = sequence{}
	r.dxAbsLowPass = 0
}

func (r *AutoStitcher) record(prevTS time.Time, frame image.Image, dx int, ts time.Time) {
	log.Trace().Time("prevTS", prevTS).Time("ts", ts).Int("dx", dx).Msg("record")
	if r.seq.startTS == nil {
		r.seq.startTS = &prevTS
	}

	r.seq.frames = append(r.seq.frames, frame)
	r.seq.dx = append(r.seq.dx, dx)
	r.seq.ts = append(r.seq.ts, ts)
}

func iabs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

// TryStitchAndReset tries to stitch any remaining frames and resets the sequence.
func (r *AutoStitcher) TryStitchAndReset() *Train {
	defer r.reset()

	if len(r.seq.dx) == 0 {
		log.Info().Msg("nothing to assemble")
		return nil
	}

	log.Info().Msg("end of sequence, trying to stitch")
	train, err := fitAndStitch(r.seq, r.c)
	if err != nil {
		log.Err(err).Time("startTs", r.seq.ts[0]).Msg("unable to fit and stitch sequence")
	}

	return train
}

func sum3(v [3]float64) float64 {
	return v[0] + v[1] + v[2]
}

// Frame adds a frame to the AutoStitcher.
// Takes ownership of the image data buffer, so be sure to make a copy before passing it.
func (r *AutoStitcher) Frame(frameColor image.Image, ts time.Time) *Train {
	t0 := time.Now()
	defer func() {
		log.Trace().Dur("dur", time.Since(t0)).Msg("Frame() duration")
	}()

	log.Trace().Time("ts", ts).Uint64("frameIx", r.prevFrameIx).Msg("Frame()")

	// Convert to gray.
	frameRGBA := imutil.ToRGBA(frameColor)
	// Make sure we always save the previous frame.
	defer func() {
		r.prevFrameIx++
		r.prevFrameTS = ts
		r.prevFrameColor = frameColor
		r.prevFrameRGBA = frameRGBA
	}()

	if r.prevFrameColor == nil {
		// First frame, we skip as we need a previous one to do any processing.
		return nil
	}

	// Compute fps and min/max allowed pixel difference.
	framePeriodS := ts.Sub(r.prevFrameTS).Seconds()
	if framePeriodS < minFramePeriodS {
		log.Warn().Float64("framePeriodS", framePeriodS).Msg("frame period too small")
		return nil
	}
	minDx := r.c.minPxPerFrame(framePeriodS)
	maxDx := r.c.maxPxPerFrame(framePeriodS)

	// Sanity check.
	if frameRGBA.Rect.Dx() < maxDx*3 {
		log.Error().Int("dx", frameRGBA.Rect.Dx()).Int("maxDx*3", maxDx*3).Float64("framePeriodS", framePeriodS).Msg("image is not wide enough to resolve the given max speed")
		return nil
	}

	// Check for minimal contrast and brightness. TODO: Maybe do directly on RGBA image.
	avg, avgDev := avg.RGBA(frameRGBA)
	if sum3(avg)/3 < minContrastAvg || sum3(avgDev)/3 < minContrastAvgDev {
		log.Trace().Interface("avgDev", avgDev).Interface("avg", avg).Msg("contrast too low, discarding")
		return nil
	}

	dx, score := findOffset(r.prevFrameRGBA, frameRGBA, maxDx)
	log.Debug().Uint64("prevFrameIx", r.prevFrameIx).Int("dx", dx).Float64("score", score).Msg("received frame")

	isActive := len(r.seq.dx) > 0
	if isActive {
		r.dxAbsLowPass = r.dxAbsLowPass*(dxLowPassFactor) + math.Abs(float64(dx))*(1-dxLowPassFactor)

		// Bail out before we use too much memory.
		if len(r.seq.dx) > maxSeqLen {
			log.Debug().Msg("len(r.seq.dx) > maxSeqLen")
			return r.TryStitchAndReset()
		}

		// We have reached the end of a sequence.
		if r.dxAbsLowPass < float64(minDx) {
			log.Debug().Float64("dxAbsLowPass", r.dxAbsLowPass).Msg("r.dxAbsLowPass < float64(minDx)")
			return r.TryStitchAndReset()
		}

		r.record(r.prevFrameTS, frameColor, dx, ts)
		return nil
	}

	if score >= goodScoreNoMove && iabs(dx) < minDx {
		log.Debug().Msg("not moving")
		return nil
	}

	if score >= goodScoreMove && iabs(dx) >= minDx && iabs(dx) <= maxDx {
		log.Info().Msg("start of new sequence")
		r.record(r.prevFrameTS, frameColor, dx, ts)
		r.dxAbsLowPass = math.Abs(float64(dx))
		return nil
	}

	log.Debug().
		Float64("score", score).
		Float64("goodScoreMove", goodScoreMove).
		Interface("avgDev", avgDev).
		Interface("avg", avg).
		Int("dx", dx).
		Int("minDx", minDx).
		Int("maxDx", maxDx).
		Msg("inconclusive frame")
	return nil
}
