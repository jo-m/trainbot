package stitch

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"math"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	maxMemoryMB = 1024 * 1024 * 50
)

func isign(x int) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}

func sign(x float64) float64 {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}

func stitch(frames []image.Image, dx []int) (*image.RGBA, error) {
	t0 := time.Now()
	defer log.Trace().Dur("dur", time.Since(t0)).Msg("stitch() duration")

	// Sanity checks.
	if len(dx) < 2 {
		return nil, errors.New("sequence too short to stitch")
	}
	if len(frames) != len(dx) {
		log.Panic().Msg("frames and dx do not have the same length, this should not happen")
	}
	fsz := frames[0].Bounds().Size()
	for _, f := range frames {
		if f.Bounds().Min.X != 0 ||
			f.Bounds().Min.Y != 0 ||
			f.Bounds().Size() != fsz {
			log.Panic().Msg("frame bounds or size not consistent, this should not happen")
		}
	}

	// Calculate base width.
	sign := isign(dx[0])
	w := fsz.X * sign
	h := fsz.Y
	for _, x := range dx[1:] {
		if isign(x) != sign {
			return nil, errors.New("dx elements do not have consistent sign")
		}
		w += x
	}

	// Memory alloc sanity check.
	rect := image.Rect(0, 0, iabs(w), h)
	if rect.Size().X*rect.Size().Y*4 > maxMemoryMB {
		return nil, fmt.Errorf("would allocate too much memory: size %dx%d", rect.Size().X, rect.Size().Y)
	}
	img := image.NewRGBA(rect)

	// Forward?
	if w > 0 {
		pos := 0
		for i, f := range frames {
			draw.Draw(img, f.Bounds().Add(image.Pt(pos, 0)), f, f.Bounds().Min, draw.Src)
			pos += dx[i]
		}
	} else {
		// Backwards.
		pos := -w - fsz.X
		for i, f := range frames {
			draw.Draw(img, f.Bounds().Add(image.Pt(pos, 0)), f, f.Bounds().Min, draw.Src)
			pos += dx[i]
		}
	}

	return img, nil
}

type Train struct {
	StartTS time.Time
	EndTS   time.Time

	Image *image.RGBA `json:"-"`

	SpeedPxS  float64
	AccelPxS2 float64
	Conf      Config
}

// SpeedMpS returns the absolute speed in m/s.
func (t *Train) SpeedMpS() float64 {
	return math.Abs(t.SpeedPxS) / t.Conf.PixelsPerM
}

// AccelMpS2 returns the acceleration in m/2^2, corrected for speed direction:
// Positive means accelerating, negative means breaking.
func (t *Train) AccelMpS2() float64 {
	return t.AccelPxS2 / t.Conf.PixelsPerM * sign(t.SpeedPxS)
}

// Direction returns the train direction. Right = true, left = false.
func (t *Train) Direction() bool {
	return t.SpeedPxS > 0
}

// fitAndStitch tries to stitch an image from a sequence.
// Will first try to fit a constant acceleration speed model for smoothing.
// Might modify seq (drops leading frames with no movement).
func fitAndStitch(seq sequence, c Config) (*Train, error) {
	start := time.Now()
	defer log.Trace().Dur("dur", time.Since(start)).Msg("fitAndStitch() duration")

	// Sanity checks.
	if len(seq.frames) != len(seq.dx) || len(seq.frames) != len(seq.ts) {
		log.Panic().Msg("length of frames, dx, ts are not equal, this should not happen")
	}
	if seq.startTS.IsZero() {
		log.Panic().Msg("startTS is zero, this should not happen")
	}
	if len(seq.dx) == 0 || seq.dx[0] == 0 {
		log.Panic().Int("len", len(seq.dx)).Msg("sequence is empty or first value is 0")
	}

	// Remove trailing zeros.
	for len(seq.dx) > 0 && seq.dx[len(seq.dx)-1] == 0 {
		seq.dx = seq.dx[:len(seq.dx)-1]
		seq.ts = seq.ts[:len(seq.ts)-1]
		seq.frames = seq.frames[:len(seq.frames)-1]
	}

	dxFit, v0, a, err := fitDx(seq)
	if err != nil {
		return nil, fmt.Errorf("was not able to fit the sequence: %w", err)
	}

	img, err := stitch(seq.frames, dxFit)
	if err != nil {
		return nil, fmt.Errorf("unable to assemble image: %w", err)
	}

	// Estimate speed at halftime.
	t0 := seq.ts[0]
	tMid := seq.ts[len(seq.ts)/2]
	speed := v0 + a*tMid.Sub(t0).Seconds()

	tEnd := seq.ts[len(seq.ts)-1]

	return &Train{
		t0,
		tEnd,
		img,
		-speed, // Negate because when things move to the left we get positive dx values.
		-a,
		c,
	}, nil
}
