package main

import (
	"image"
	"image/draw"
	"math"

	"github.com/jo-m/trainbot/internal/pkg/imutil"
	"github.com/rs/zerolog/log"
)

const magicBad = 0xDEADBEEF

type estimatorConfig struct {
	PixelsPerM  float32 `arg:"--px-per-m" default:"50" help:"Pixels per meter, can be reconstructed from sleepers"`
	MinSpeedKPH float32 `arg:"--min-speed-kph" default:"10" help:"Assumed train min speed, km/h"`
	MaxSpeedKPH float32 `arg:"--max-speed-kph" default:"120" help:"Assumed train max speed, km/h"`

	// TODO: calculate automatically
	VideoFPS float32 `arg:"--video-fps" default:"30" help:"Assumed video FPS"`
}

func (e *estimatorConfig) minPxPerFrame() int {
	return int(e.MinSpeedKPH*e.PixelsPerM/e.VideoFPS) - 1
}

func (e *estimatorConfig) maxPxPerFrame() int {
	return int(e.MaxSpeedKPH*e.PixelsPerM/e.VideoFPS) - 1
}

// TODO: rename?
type estimator struct {
	c estimatorConfig

	count uint64

	dx       []int
	frames   []image.Image
	startC   uint64
	avgSpeed float64
}

func newEstimator(c estimatorConfig) *estimator {
	return &estimator{
		c: c,
	}
}

func cleanup(dx []int) []int {
	// remove trailing zeros
	zeros := 0
	for i := len(dx) - 1; i >= 0; i-- {
		if dx[i] == 0 { // TODO: handle other small values
			zeros++
		} else {
			break
		}
	}
	dx = dx[:len(dx)-zeros]

	for i, val := range dx {
		if val == 0 || val == magicBad { // TODO: handle other small values
			dx[i] = dx[i-1] // TODO: zeros at start??
		}
	}

	// TODO: remove last frames
	return dx[:len(dx)-7]
}

// TODO: handle other direction
func assembleFrames(dx []int, frames []image.Image) {
	w := frames[0].Bounds().Dx()
	for _, x := range dx {
		w += x
	}
	h := frames[0].Bounds().Dy()

	for _, f := range frames {
		if f.Bounds().Min.X != 0 ||
			f.Bounds().Min.Y != 0 ||
			f.Bounds().Size() != frames[0].Bounds().Size() {
			panic("invalid frames") // TODO
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// TODO
	sum := 0
	for i, f := range frames {
		draw.Draw(img, f.Bounds().Add(image.Pt(sum, 0)), f, f.Bounds().Min, draw.Src)
		sum += dx[i]
	}

	imutil.Dump("imgs/_assembled.png", img)
}

// will take a copy of the frame
func (e *estimator) Frame(frame image.Image, good bool, dx int) {
	defer func() { e.count++ }()

	dxAbs := dx
	if dxAbs < 0 {
		dxAbs = -dx
	}

	if len(e.dx) == 0 && (dxAbs < e.c.minPxPerFrame() || dxAbs > e.c.maxPxPerFrame()) {
		return
	}

	if len(e.dx) == 0 {
		log.Info().Uint64("count", e.count).Msg("new train")
		e.startC = e.count
		e.avgSpeed = float64(dx)
	}

	if good && dxAbs <= e.c.maxPxPerFrame() {
		e.dx = append(e.dx, dx)
		e.avgSpeed = float64(dx)*0.15 + e.avgSpeed*0.85
	} else {
		e.dx = append(e.dx, magicBad)
	}

	e.frames = append(e.frames, imutil.ToRGBA(frame))

	if math.Abs(e.avgSpeed) < float64(e.c.minPxPerFrame()) {
		// TODO: add logging everywhere
		clean := cleanup(e.dx)
		frames := e.frames[:len(clean)]
		assembleFrames(clean, frames)

		e.dx = nil
		e.frames = nil
		e.startC = 0
		e.avgSpeed = 0
	}
}

func (e *estimator) Done() bool {
	return true
}
