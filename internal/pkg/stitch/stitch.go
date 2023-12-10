package stitch

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"math"
	"time"

	"github.com/jo-m/trainbot/internal/pkg/prometheus"
	"github.com/mccutchen/palettor"
	"github.com/nfnt/resize"
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
	defer func() {
		log.Trace().Dur("dur", time.Since(t0)).Msg("stitch() duration")
	}()

	log.Info().Ints("dx", dx).Int("len(frames)", len(frames)).Msg("stitch()")

	// Sanity checks.
	if len(dx) < 2 {
		return nil, errors.New("sequence too short to stitch")
	}
	if len(frames) != len(dx) {
		log.Panic().Msg("frames and dx do not have the same length, this should not happen")
	}
	fb := frames[0].Bounds()
	for _, f := range frames {
		if f.Bounds() != fb {
			log.Panic().Msg("frame bounds or size not consistent, this should not happen")
		}
	}

	// Calculate base width.
	sign := isign(dx[0])
	w := fb.Dx() * sign
	h := fb.Dy()
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
			draw.Draw(img, img.Bounds().Add(image.Pt(pos, 0)), f, f.Bounds().Min, draw.Src)
			pos += dx[i]
		}
	} else {
		// Backwards.
		pos := -w - fb.Dx()
		for i, f := range frames {
			draw.Draw(img, img.Bounds().Add(image.Pt(pos, 0)), f, f.Bounds().Min, draw.Src)
			pos += dx[i]
		}
	}

	return img, nil
}

// Train represents a detected train.
type Train struct {
	StartTS time.Time
	EndTS   time.Time

	// Always positive.
	NFrames int

	// Always positive (absolute value).
	LengthPx float64
	// Positive sign means movement to the right, negative to the left.
	SpeedPxS float64
	// Positive sign means increasing speed for trains going to the right, breaking for trains going to the left.
	AccelPxS2 float64

	Conf Config

	Image *image.RGBA `json:"-"`
	GIF   *gif.GIF    `json:"-"`
}

// LengthM returns the absolute length in m.
func (t *Train) LengthM() float64 {
	return math.Abs(t.LengthPx) / t.Conf.PixelsPerM
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

// DirectionS returns the train direction as string "left" or "right".
func (t *Train) DirectionS() string {
	if t.SpeedPxS > 0 {
		return "right"
	}

	return "left"
}

func createGIF(seq sequence, stitched image.Image) (*gif.GIF, error) {
	// Extract palette.
	thumb := resize.Thumbnail(300, 300, stitched, resize.Lanczos3)
	const (
		paletteSize = 20
		nIter       = 100
	)
	pal, err := palettor.Extract(paletteSize, nIter, thumb)
	if err != nil {
		return nil, err
	}

	g := gif.GIF{}

	prevTS := *seq.startTS
	rect := seq.frames[0].Bounds().Sub(seq.frames[0].Bounds().Min)
	for i, ts := range seq.ts {
		dt := ts.Sub(prevTS)

		// Skip every other frame.
		if i%2 == 1 {
			continue
		}

		paletted := image.NewPaletted(rect, pal.Colors())
		draw.Draw(paletted, paletted.Bounds(), seq.frames[i], seq.frames[i].Bounds().Min, draw.Src)

		g.Image = append(g.Image, paletted)
		g.Delay = append(g.Delay, int(dt.Seconds()*100))

		prevTS = ts
	}

	return &g, nil
}

// fitAndStitch tries to stitch an image from a sequence.
// Will first try to fit a constant acceleration speed model for smoothing.
// Might modify seq (drops leading frames with no movement).
func fitAndStitch(seq sequence, c Config) (*Train, error) {
	start := time.Now()
	defer func() {
		log.Trace().Dur("dur", time.Since(start)).Msg("fitAndStitch() duration")
	}()

	log.Info().Ints("dx", seq.dx).Int("len(frames)", len(seq.frames)).Msg("fitAndStitch()")

	// Sanity checks.
	if len(seq.frames) != len(seq.dx) || len(seq.frames) != len(seq.ts) {
		log.Panic().Msg("length of frames, dx, ts are not equal, this should not happen")
	}
	if seq.startTS == nil {
		log.Panic().Msg("startTS is nil, this should not happen")
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
	prometheus.RecordSequenceLength(len(seq.frames))

	dxFit, ds, v0, a, err := fitDx(seq, float64(c.maxPxPerFrame(1)))
	if err != nil {
		return nil, fmt.Errorf("was not able to fit the sequence: %w", err)
	}

	if math.Abs(ds) < c.minLengthPx() {
		return nil, fmt.Errorf("discarded because too short, %f < %f", ds, c.minLengthPx())
	}

	// Estimate speed at halftime.
	t0 := seq.ts[0]
	tMid := seq.ts[len(seq.ts)/2]
	speed := v0 + a*tMid.Sub(t0).Seconds()

	if math.Abs(speed) < c.minSpeedPxPS() {
		return nil, fmt.Errorf("discarded because too slow, %f < %f", speed, c.minSpeedPxPS())
	}

	img, err := stitch(seq.frames, dxFit)
	if err != nil {
		return nil, fmt.Errorf("unable to assemble image: %w", err)
	}

	gif, err := createGIF(seq, img)
	if err != nil {
		panic(err)
	}

	return &Train{
		t0,
		seq.ts[len(seq.ts)-1],
		len(seq.frames),
		ds,
		-speed, // Negate because when things move to the left we get positive dx values.
		-a,
		c,
		img,
		gif,
	}, nil
}
