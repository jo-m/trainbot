package main

import (
	"image"
	"io"
	"os"
	"runtime/pprof"

	"github.com/alexflint/go-arg"
	"github.com/jo-m/trainbot/internal/pkg/imutil"
	"github.com/jo-m/trainbot/internal/pkg/logging"
	"github.com/jo-m/trainbot/pkg/est"
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

	PixelsPerM  float64 `arg:"--px-per-m" default:"50" help:"Pixels per meter, can be reconstructed from sleepers: they are usually 0.6m apart (in Europe)"`
	MinSpeedKPH float64 `arg:"--min-speed-kph" default:"10" help:"Assumed train min speed, km/h"`
	MaxSpeedKPH float64 `arg:"--max-speed-kph" default:"120" help:"Assumed train max speed, km/h"`
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

	src, fps, err := vid.NewSrc(c.VideoFile)
	if err != nil {
		log.Panic().Err(err).Send()
	}

	rec := rec.NewAutoRec(c.RecBasePath)
	est := est.NewEstimator(est.Config{
		PixelsPerM:  c.PixelsPerM,
		MinSpeedKPH: c.MinSpeedKPH,
		MaxSpeedKPH: c.MinSpeedKPH,

		VideoFPS: fps,
	})

	for i := 0; ; i++ {
		frame, ts, err := src.GetFrame()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Panic().Err(err).Send()
		}

		cropped, err := imutil.Sub(frame, r)
		if err != nil {
			log.Panic().Err(err).Send()
		}

		err = rec.Frame(cropped, *ts)
		if err != nil {
			log.Panic().Err(err).Send()
		}

		est.Frame(imutil.ToGray(cropped), *ts)
	}
}
