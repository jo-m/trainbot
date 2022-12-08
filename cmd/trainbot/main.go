package main

import (
	"image"
	"io"
	"os"
	"runtime/pprof"

	"github.com/alexflint/go-arg"
	"github.com/jo-m/trainbot/internal/pkg/imutil"
	"github.com/jo-m/trainbot/internal/pkg/logging"
	"github.com/jo-m/trainbot/pkg/pmatch"
	"github.com/jo-m/trainbot/pkg/rec"
	"github.com/jo-m/trainbot/pkg/vid"
	"github.com/rs/zerolog/log"
)

type config struct {
	logging.LogConfig

	VideoFile  string `arg:"--video-file" help:"Video file"`
	CPUProfile string `arg:"--cpu-profile" help:"Write CPU profile to this file"`

	RectX uint `arg:"-X" help:"Rect to look at, x (left)"`
	RectY uint `arg:"-Y" help:"Rect to look at, y (top)"`
	RectW uint `arg:"-W" help:"Rect to look at, width"`
	RectH uint `arg:"-H" help:"Rect to look at, height"`

	RecBasePath string `arg:"--rec-base-path" help:"Base path to store recordings" default:"imgs"`

	estimatorConfig
}

// TODO: move into estimator
func findOffset(prev, curr *image.Gray, maxDx int) (goodEstimate bool, dx int) {
	if prev.Rect.Size() != curr.Rect.Size() {
		panic("inconsistent size")
	}

	// centered crop from prev frame,
	// width is 3x max pixels per frame given by max velocity
	w := maxDx * 3
	// and 3/4 of frame height
	h := int(float64(prev.Rect.Dy())*3/4 + 1)
	subRect := image.Rect(0, 0, w, h).
		Add(curr.Rect.Min).
		Add(
			curr.Rect.Size().
				Sub(image.Pt(int(w), h)).
				Div(2),
		)
	sub, err := imutil.Sub(prev, subRect)
	if err != nil {
		log.Panic().Err(err).Send()
	}

	// centered slice crop from next frame,
	// width is 1x max pixels per frame given by max velocity
	// and 3/4 of frame height
	sliceRect := image.Rect(0, 0, maxDx, h).
		Add(curr.Rect.Min).
		Add(
			curr.Rect.Size().
				Sub(image.Pt(w, h)).
				Div(2),
		)

	slice, err := imutil.Sub(curr, sliceRect)
	if err != nil {
		log.Panic().Err(err).Send()
	}

	// we expect this x value found by search
	// if nothing has moved
	xZero := sliceRect.Min.Sub(subRect.Min).X

	x, _, score := pmatch.SearchGrayC(sub.(*image.Gray), slice.(*image.Gray))
	return score > 0.95, x - xZero
}

func main() {
	c := config{}
	p := arg.MustParse(&c)
	logging.MustInit(c.LogConfig)

	r := image.Rect(0, 0, int(c.RectW), int(c.RectH)).Add(image.Pt(int(c.RectX), int(c.RectY)))
	// sz := r.Size().X * r.Size().Y
	if r.Size().X < 100 || r.Size().Y < 100 {
		p.Fail("rect is too small")
	}
	if r.Size().X > 300 || r.Size().Y > 300 {
		p.Fail("rect is too large")
	}

	log.Info().Msg("starting")

	if c.CPUProfile != "" {
		log.Info().Str("file", c.CPUProfile).Msg("writing profile")
		f, err := os.Create(c.CPUProfile)
		if err != nil {
			log.Panic().Err(err).Send()
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	src, err := vid.NewSrc(c.VideoFile)
	if err != nil {
		log.Panic().Err(err).Send()
	}

	rec := rec.NewAutoRec(c.RecBasePath)
	e := newEstimator(c.estimatorConfig)
	var prevGray *image.Gray
	for i := 0; ; i++ {
		frame, ts, err := src.GetFrame()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Panic().Err(err).Send()
		}

		cropped := frame.SubImage(r)
		err = rec.Frame(cropped, *ts)
		if err != nil {
			log.Panic().Err(err).Send()
		}

		gray := imutil.ToGray(cropped)
		if prevGray != nil {
			good, dx := findOffset(prevGray, gray, c.maxPxPerFrame())
			e.Frame(cropped, good, dx)
		}

		prevGray = gray
	}
}
