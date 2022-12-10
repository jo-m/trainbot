package main

import (
	"fmt"
	"image"
	"io"
	"net/http"

	"github.com/alexflint/go-arg"
	"github.com/jo-m/trainbot/internal/pkg/logging"
	"github.com/jo-m/trainbot/pkg/server"
	"github.com/jo-m/trainbot/pkg/vid"
	"github.com/rs/zerolog/log"
)

const (
	failedFramesMax = 50
)

type config struct {
	logging.LogConfig

	LiveReload bool   `arg:"--live-reload,env:LIVE_RELOAD" default:"false" help:"Live reload WWW static files"`
	ListenAddr string `arg:"--listen-addr,env:LISTEN_ADDR" default:"localhost:8080" help:"Address and port to listen on"`

	CameraDevice       string `arg:"--camera-device" help:"Video4linux device file, e.g. /dev/video0"`
	CameraFormatFourCC string `arg:"--format-fourcc" default:"MJPG" help:"Camera pixel format FourCC string, ignored if using video file"`
	CameraFrameSizeW   int    `arg:"--framesz-w" default:"1920" help:"Camera frame size width, ignored if using video file"`
	CameraFrameSizeH   int    `arg:"--framesz-h" default:"1080" help:"Camera frame size height, ignored if using video file"`
}

func parseCheckArgs() config {
	c := config{}
	p := arg.MustParse(&c)
	logging.MustInit(c.LogConfig)

	if c.CameraDevice == "" {
		p.Fail("no camera device passed")
	}

	return c
}

func main() {
	c := parseCheckArgs()

	log.Info().Interface("config", c).Msg("starting")

	srv, err := server.NewServer(!c.LiveReload)
	if err != nil {
		log.Panic().Err(err).Msg("unable to initialize server")
	}

	src, err := vid.NewCamSrc(vid.CamConfig{
		DeviceFile: c.CameraDevice,
		Format:     vid.FourCC(c.CameraFormatFourCC),
		FrameSize:  image.Point{c.CameraFrameSizeW, c.CameraFrameSizeH},
	})
	if err != nil {
		log.Panic().Err(err).Str("path", c.CameraDevice).Msg("failed to open video source")
	}

	go func() {
		log.Info().Str("url", fmt.Sprintf("http://%s", c.ListenAddr)).Msg("serving")
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
			log.Panic().Msg("unsupported image format")
		}
		if err != nil {
			failedFrames++
			log.Warn().Err(err).Int("failedFrames", failedFrames).Msg("failed to retrieve frame")
			if failedFrames >= failedFramesMax {
				log.Error().Msg("retrieving frames failed too many times, exiting")
				return
			}
			continue
		} else {
			failedFrames = 0
		}

		if i%3 == 0 {
			srv.SetFrameRawJPEG(frameRaw)
		}
	}
}
