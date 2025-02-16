// Package main (confighelper) is a simple web server which helps to find the correct command line arguments for running trainbot.
package main

import (
	"fmt"
	"image"
	"io"
	"math"
	"net/http"

	"github.com/alexflint/go-arg"
	"github.com/jo-m/trainbot/internal/pkg/logging"
	"github.com/jo-m/trainbot/internal/pkg/server"
	"github.com/jo-m/trainbot/pkg/vid"
	"github.com/rs/zerolog/log"
)

const (
	failedFramesMax = 50
	inputFilePiCam3 = "picam3"
)

type config struct {
	logging.LogConfig

	LiveReload bool   `arg:"--live-reload" default:"false" help:"Do not bake in WWW static files (browser window reload is still needed)"`
	ListenAddr string `arg:"--listen-addr" default:"localhost:8080" help:"Address and port to listen on"`

	InputFile string `arg:"--input,required" help:"Video4linux device file, e.g. /dev/video0, or 'picam3'"`
	CameraW   int    `arg:"--camera-w" default:"1920" help:"Camera frame size width, ignored for picam3"`
	CameraH   int    `arg:"--camera-h" default:"1080" help:"Camera frame size height, ignored for picam3"`

	Rotate180 bool `arg:"--rotate-180,env:ROTATE_180" help:"Rotate camera picture 180 degrees (only picam3)"`

	ProbeOnly bool `arg:"--probe-only" help:"Only print v4l camera probe output and exit"`
}

func parseCheckArgs() config {
	c := config{}
	c.LogPretty = true
	p := arg.MustParse(&c)
	logging.MustInit(c.LogConfig)

	if c.InputFile == "" && !c.ProbeOnly {
		p.Fail("no camera device passed")
	}

	return c
}

func main() {
	c := parseCheckArgs()

	log.Info().Interface("config", c).Msg("starting")

	if c.ProbeOnly {
		cams, err := vid.DetectCams()
		if err != nil {
			log.Panic().Err(err).Msg("unable to probe")
		}

		if len(cams) == 0 {
			log.Panic().Err(err).Msg("no cameras detected")
		}

		for _, cam := range cams {
			fmt.Println(cam)
		}

		return
	}

	srv, err := server.NewServer(!c.LiveReload)
	if err != nil {
		log.Panic().Err(err).Msg("unable to initialize server")
	}

	var src vid.Src
	if c.InputFile == inputFilePiCam3 {
		src, err = vid.NewPiCam3Src(vid.PiCam3Config{
			Focus:     0,
			Rotate180: c.Rotate180,
			Format:    vid.FourCCMJPEG,
			FPS:       5,
		})
	} else {
		src, err = vid.NewCamSrc(vid.CamConfig{
			DeviceFile: c.InputFile,
			Format:     vid.FourCCMJPEG,
			FrameSize:  image.Point{c.CameraW, c.CameraH},
		})
	}
	if err != nil {
		log.Panic().Err(err).Str("path", c.InputFile).Msg("failed to open video source")
	}

	go func() {
		log.Info().Str("url", fmt.Sprintf("http://%s", c.ListenAddr)).Msg("serving")
		// #nosec G114 This should not be exposed to the internet and only lives temporarily.
		err := http.ListenAndServe(c.ListenAddr, srv.GetMux())
		log.Panic().Err(err).Send()
	}()

	failedFrames := 0
	for i := 0; ; i++ {
		frameRaw, fourcc, _, err := src.GetFrameRaw()
		if err == io.EOF {
			log.Info().Msg("no more frames")
			break
		}
		if fourcc != vid.FourCCMJPEG {
			err = fmt.Errorf("unsupported image format: %d", fourcc)
		}

		if err != nil {
			failedFrames++
			log.Warn().Err(err).Int("failedFrames", failedFrames).Msg("failed to retrieve frame")
			if failedFrames >= failedFramesMax {
				log.Error().Msg("retrieving frames failed too many times, exiting")
				return
			}
			continue
		}
		failedFrames = 0

		// Stream, at ca. 5fps.
		everyNth := int(math.Max(src.GetFPS()/5, 1))
		if i%everyNth == 0 {
			err := srv.SetFrameRawJPEG(frameRaw)
			if err != nil {
				log.Panic().Err(err).Send()
			}
		}
	}
}
