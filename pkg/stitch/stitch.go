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

const maxMemoryMB = 1024 * 1024 * 50

type sequence struct {
	// Those slices always must have the same length.
	// dx[i] is the assumed offset between frames[i] and frames[i+1].
	// ts[i] is the timestamp of that frame.
	// All frames must have the same size.
	dx     []int
	ts     []time.Time
	frames []image.Image
}

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

func stitch(seq sequence) (*image.RGBA, error) {
	t0 := time.Now()
	defer log.Trace().Dur("dur", time.Since(t0)).Msg("stitch() duration")

	fsz := seq.frames[0].Bounds().Size()
	for _, f := range seq.frames {
		if f.Bounds().Min.X != 0 ||
			f.Bounds().Min.Y != 0 ||
			f.Bounds().Size() != fsz {
			return nil, errors.New("frame bounds or size not consistent")
		}
	}

	// Calculate base width.
	sign := isign(seq.dx[0])
	w := fsz.X * sign
	h := fsz.Y
	for _, x := range seq.dx[1:] {
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
		for i, f := range seq.frames {
			draw.Draw(img, f.Bounds().Add(image.Pt(pos, 0)), f, f.Bounds().Min, draw.Src)
			pos += seq.dx[i]
		}
	} else {
		// Backwards.
		pos := -w - fsz.X
		for i, f := range seq.frames {
			draw.Draw(img, f.Bounds().Add(image.Pt(pos, 0)), f, f.Bounds().Min, draw.Src)
			pos += seq.dx[i]
		}
	}

	return img, nil
}

type Train struct {
	StartTS time.Time
	EndTS   time.Time
	Image   *image.RGBA
	Speed   float64
	Accel   float64
	Conf    Config
}

// absolute
func (t *Train) SpeedMpS() float64 {
	return math.Abs(t.Speed) / t.Conf.PixelsPerM
}

// corrected for speed direction
func (t *Train) AccelMpS2() float64 {
	// TODO: test
	return t.Accel / t.Conf.PixelsPerM * sign(t.Speed)
}

func fitAndStitch(seq sequence, c Config) (*Train, error) {
	start := time.Now()
	defer log.Trace().Dur("dur", time.Since(start)).Msg("fitAndStitch() duration")

	if len(seq.dx) != len(seq.frames) {
		log.Panic().Msg("length of frames and dx should be equal, this should not happen")
	}

	// Remove trailing zeros.
	for seq.dx[len(seq.dx)-1] == 0 {
		seq.dx = seq.dx[:len(seq.dx)-1]
		seq.ts = seq.ts[:len(seq.ts)-1]
		seq.frames = seq.frames[:len(seq.frames)-1]
	}

	// Various sanity checks.
	if len(seq.dx) < 10 {
		return nil, errors.New("sequence too short")
	}
	if seq.dx[0] == 0 {
		return nil, errors.New("seq.dx cannot start with a zero")
	}

	var err error
	var v0, a float64
	seq.dx, v0, a, err = fitDx(seq.ts, seq.dx)
	if err != nil {
		return nil, fmt.Errorf("was not able to fit the sequence: %w", err)
	}

	img, err := stitch(seq)
	if err != nil {
		return nil, fmt.Errorf("unable to assemble image: %w", err)
	}

	// estimate speed at halftime
	t0 := seq.ts[0]
	tMid := seq.ts[len(seq.ts)/2]
	speed := v0 + a*tMid.Sub(t0).Seconds()

	tEnd := seq.ts[len(seq.ts)-1]

	return &Train{
		t0,
		tEnd,
		img,
		speed,
		a,
		c,
	}, nil
}
