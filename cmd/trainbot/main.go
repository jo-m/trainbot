// Package main (trainbot) tries to automatically stitch images of passing trains.
package main

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
	"runtime/pprof"
	"sync"

	"github.com/alexflint/go-arg"
	"github.com/jo-m/trainbot/internal/pkg/imutil"
	"github.com/jo-m/trainbot/internal/pkg/logging"
	"github.com/jo-m/trainbot/internal/pkg/stitch"
	"github.com/jo-m/trainbot/pkg/vid"
	"github.com/rs/zerolog/log"
)

type config struct {
	logging.LogConfig

	InputFile          string `arg:"--input" help:"Video4linux device file or regular video file, e.g. /dev/video0, video.mp4"`
	CameraFormatFourCC string `arg:"--camera-format-fourcc" default:"MJPG" help:"Camera pixel format FourCC string, ignored if using video file"`
	CameraW            int    `arg:"--camera-w" default:"1920" help:"Camera frame size width, ignored if using video file or picam3"`
	CameraH            int    `arg:"--camera-h" default:"1080" help:"Camera frame size height, ignored if using video file or picam3"`

	RectX uint `arg:"-X" help:"Rect to look at, x (left)"`
	RectY uint `arg:"-Y" help:"Rect to look at, y (top)"`
	RectW uint `arg:"-W" help:"Rect to look at, width"`
	RectH uint `arg:"-H" help:"Rect to look at, height"`

	PixelsPerM  float64 `arg:"--px-per-m" default:"45" help:"Pixels per meter, can be reconstructed from sleepers: they are usually 0.6m apart (in Europe)"`
	MinSpeedKPH float64 `arg:"--min-speed-kph" default:"10" help:"Assumed train min speed, km/h"`
	MaxSpeedKPH float64 `arg:"--max-speed-kph" default:"120" help:"Assumed train max speed, km/h"`

	OutputDir string `arg:"--output-path" help:"Directory to store output" default:"imgs"`

	CPUProfile  bool `arg:"--cpu-profile" help:"Write CPU profile"`
	HeapProfile bool `arg:"--heap-profile" help:"Write memory heap profiles"`
}

func (c *config) getRect() image.Rectangle {
	return image.Rect(0, 0, int(c.RectW), int(c.RectH)).Add(image.Pt(int(c.RectX), int(c.RectY)))
}

const (
	rectSizeMin = 100
	rectSizeMax = 500

	failedFramesMax = 50

	inputFilePiCam3 = "picam3"

	profCPUFile  = "prof-cpu.gz"
	profHeapFile = "prof-heap-%05d.gz"
)

func parseCheckArgs() config {
	c := config{}
	p := arg.MustParse(&c)
	logging.MustInit(c.LogConfig)

	if c.InputFile == "" {
		p.Fail("no camera device or video file passed")
	}

	r := c.getRect()
	if r.Size().X < rectSizeMin || r.Size().Y < rectSizeMin {
		p.Fail("rect is too small")
	}
	if r.Size().X > rectSizeMax || r.Size().Y > rectSizeMax {
		p.Fail("rect is too large")
	}

	return c
}

func openSrc(c config) (vid.Src, error) {
	// Pi cam.
	if c.InputFile == inputFilePiCam3 {
		return vid.NewPiCam3Src(vid.PiCam3Config{
			Rect:      c.getRect(),
			Focus:     0,
			Rotate180: true,
			Format:    vid.FourCCMJPEG,
			FPS:       30,
		})
	}

	stat, err := os.Stat(c.InputFile)
	if err != nil {
		return nil, err
	}

	if stat.Mode().IsRegular() {
		// Video file.
		return vid.NewFileSrc(c.InputFile, false)
	}

	return vid.NewCamSrc(vid.CamConfig{
		DeviceFile: c.InputFile,
		Format:     vid.FourCCFromString(c.CameraFormatFourCC),
		FrameSize:  image.Point{c.CameraW, c.CameraH},
	})
}

func detectTrainsForever(c config, trainsOut chan<- *stitch.Train) {
	rect := c.getRect()

	src, err := openSrc(c)
	if err != nil {
		log.Panic().Err(err).Str("path", c.InputFile).Msg("failed to open video source")
	}
	defer src.Close()
	srcBuf := vid.NewSrcBuf(src, failedFramesMax)

	stitcher := stitch.NewAutoStitcher(stitch.Config{
		PixelsPerM:  c.PixelsPerM,
		MinSpeedKPH: c.MinSpeedKPH,
		MaxSpeedKPH: c.MaxSpeedKPH,
	})
	defer stitcher.TryStitchAndReset()

	for i := uint64(0); ; i++ {
		frame, ts, err := srcBuf.GetFrame()
		if err != nil {
			log.Err(err).Msg("no more frames")
			break
		}

		var cropped image.Image
		if c.InputFile == inputFilePiCam3 {
			// PiCam output is already cropped.
			cropped = frame
		} else {
			cropped, err = imutil.Sub(frame, rect)
			if err != nil {
				log.Panic().Err(err).Msg("failed to crop frame")
			}
		}

		if cropped.Bounds().Size() != rect.Size() {
			log.Panic().Interface("cam", cropped.Bounds().Size()).Interface("conf", rect.Size()).Msg("rect size mismatch")
		}

		train := stitcher.Frame(cropped, *ts)
		if train != nil {
			trainsOut <- train
		}

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

func processTrains(trainsIn <-chan *stitch.Train, wg *sync.WaitGroup) {
	defer wg.Done()

	for train := range trainsIn {
		log.Info().
			Time("ts", train.StartTS).
			Float64("speedMpS", train.SpeedMpS()).
			Float64("speedKmh", train.SpeedMpS()*3.6).
			Float64("accelMpS2", train.AccelMpS2()).
			Msg("found train")

		tsString := train.StartTS.Format("20060102_150405.999_Z07:00")
		imutil.Dump(fmt.Sprintf("imgs/train_%s.jpg", tsString), train.Image)
		imutil.DumpGIF(fmt.Sprintf("imgs/train_%s.gif", tsString), train.GIF)

		func() {
			meta, err := os.Create(fmt.Sprintf("imgs/train_%s.json", tsString))
			if err != nil {
				log.Err(err).Send()
			}
			defer meta.Close()

			enc := json.NewEncoder(meta)
			enc.SetIndent("", "  ")
			err = enc.Encode(train)
			if err != nil {
				log.Err(err).Send()
			}
		}()
	}
}

func main() {
	c := parseCheckArgs()

	// Try to create output directory.
	os.MkdirAll(c.OutputDir, 0755)

	log.Info().Interface("config", c).Msg("starting")

	if c.CPUProfile {
		log.Info().Str("file", profCPUFile).Msg("writing CPU profile")
		f, err := os.Create(profCPUFile)
		if err != nil {
			log.Panic().Err(err).Msg("failed to create CPU profile file")
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	trains := make(chan *stitch.Train)
	done := sync.WaitGroup{}
	done.Add(1)
	go processTrains(trains, &done)

	detectTrainsForever(c, trains)

	close(trains)
	done.Wait()
}
