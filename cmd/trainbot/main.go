package main

import (
	"fmt"
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

	VideoFile string `arg:"--video-file" help:"Video file, e.g. video.mp4"`

	CameraDevice       string `arg:"--camera-device" help:"Video4linux device file, e.g. /dev/video0"`
	CameraFormatFourCC string `arg:"--format-fourcc" default:"MJPG" help:"Camera pixel format FourCC string, ignored if using video file"`
	CameraFrameSizeW   int    `arg:"--framesz-w" default:"1920" help:"Camera frame size width, ignored if using video file"`
	CameraFrameSizeH   int    `arg:"--framesz-h" default:"1080" help:"Camera frame size height, ignored if using video file"`

	RectX uint `arg:"-X" help:"Rect to look at, x (left)"`
	RectY uint `arg:"-Y" help:"Rect to look at, y (top)"`
	RectW uint `arg:"-W" help:"Rect to look at, width"`
	RectH uint `arg:"-H" help:"Rect to look at, height"`

	PixelsPerM  float64 `arg:"--px-per-m" default:"50" help:"Pixels per meter, can be reconstructed from sleepers: they are usually 0.6m apart (in Europe)"`
	MinSpeedKPH float64 `arg:"--min-speed-kph" default:"10" help:"Assumed train min speed, km/h"`
	MaxSpeedKPH float64 `arg:"--max-speed-kph" default:"120" help:"Assumed train max speed, km/h"`

	RecBasePath string `arg:"--rec-base-path" help:"Base path to store recordings" default:"imgs"`

	CPUProfile  bool `arg:"--cpu-profile" help:"Write CPU profile"`
	HeapProfile bool `arg:"--heap-profile" help:"Write memory heap profiles"`
}

const (
	rectSizeMin = 100
	rectSizeMax = 400

	failedFramesMax = 50

	profCPUFile  = "prof-cpu.gz"
	profHeapFile = "prof-heap-%05d.gz"
)

func main() {
	c := config{}
	p := arg.MustParse(&c)
	logging.MustInit(c.LogConfig)

	if c.CameraDevice == "" && c.VideoFile == "" {
		p.Fail("no camera device and no video file passed")
	}
	if c.CameraDevice != "" && c.VideoFile != "" {
		p.Fail("cannot pass both camera device and video file")
	}

	r := image.Rect(0, 0, int(c.RectW), int(c.RectH)).Add(image.Pt(int(c.RectX), int(c.RectY)))
	if r.Size().X < rectSizeMin || r.Size().Y < rectSizeMin {
		p.Fail("rect is too small")
	}
	if r.Size().X > rectSizeMax || r.Size().Y > rectSizeMax {
		p.Fail("rect is too large")
	}

	log.Info().Msg("starting")

	if c.CPUProfile {
		log.Info().Str("file", profCPUFile).Msg("writing CPU profile")
		f, err := os.Create(profCPUFile)
		if err != nil {
			log.Panic().Err(err).Msg("failed to create CPU profile file")
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	var src vid.Src
	var err error
	if c.CameraDevice != "" {
		src, err = vid.NewCamSrc(vid.CamConfig{
			DeviceFile: c.CameraDevice,
			Format:     vid.FourCC(c.CameraFormatFourCC),
			FrameSize:  image.Point{c.CameraFrameSizeW, c.CameraFrameSizeH},
		})
	} else {
		src, err = vid.NewFileSrc(c.VideoFile, false)
	}
	if err != nil {
		log.Panic().Err(err).Str("path", c.CameraDevice+c.VideoFile).Msg("failed to open video source")
	}

	rec := rec.NewAutoRec(c.RecBasePath)
	est := est.NewEstimator(est.Config{
		PixelsPerM:  c.PixelsPerM,
		MinSpeedKPH: c.MinSpeedKPH,
		MaxSpeedKPH: c.MinSpeedKPH,

		VideoFPS: src.GetFPS(),
	})

	failedFrames := 0
	for i := 0; ; i++ {
		frame, ts, err := src.GetFrame()
		if err == io.EOF {
			log.Info().Msg("no more frames")
			break
		}
		if err != nil {
			failedFrames++
			log.Warn().Err(err).Send()
			if failedFrames >= failedFramesMax {
				log.Panic().Msg("retrieving frames failed too many times, exiting")
			}
			continue
		} else {
			failedFrames = 0
		}

		cropped, err := imutil.Sub(frame, r)
		if err != nil {
			log.Panic().Err(err).Msg("failed to crop frame")
		}

		err = rec.Frame(cropped, *ts)
		if err != nil {
			log.Panic().Err(err).Msg("failed to record frame")
		}

		est.Frame(cropped, *ts)

		if c.HeapProfile && i%1000 == 0 {
			fname := fmt.Sprintf(profHeapFile, i)
			f, err := os.Create(fname)
			if err != nil {
				log.Err(err).Str("file", fname).Msg("failed to open heap profile file")
				continue
			}
			log.Info().Str("file", fname).Msg("writing heap profile")
			err = pprof.WriteHeapProfile(f)
			if err != nil {
				log.Err(err).Str("file", fname).Msg("failed to write heap profile")
			}
			f.Close()
		}
	}
}
