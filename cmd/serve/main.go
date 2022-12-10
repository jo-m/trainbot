package main

import (
	"fmt"
	"net/http"

	"github.com/alexflint/go-arg"
	"github.com/jo-m/trainbot/internal/pkg/logging"
	"github.com/jo-m/trainbot/pkg/server"
	"github.com/rs/zerolog/log"
)

type config struct {
	logging.LogConfig

	LiveReload bool   `arg:"--live-reload,env:LIVE_RELOAD" default:"false" help:"Live reload WWW static files"`
	ListenAddr string `arg:"--listen-addr,env:LISTEN_ADDR" default:"localhost:8080" help:"Address and port to listen on"`
}

func main() {
	c := config{}
	arg.MustParse(&c)
	logging.MustInit(c.LogConfig)

	log.Info().Interface("config", c).Msg("starting")

	srv, err := server.NewServer(!c.LiveReload)
	if err != nil {
		log.Panic().Err(err).Msg("unable to initialize server")
	}

	log.Info().Str("url", fmt.Sprintf("http://%s", c.ListenAddr)).Msg("serving")
	http.ListenAndServe(c.ListenAddr, srv.GetMux())

	http.ListenAndServe(c.ListenAddr, srv.GetMux())
}
