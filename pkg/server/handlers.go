package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jo-m/trainbot/pkg/vid"
	"github.com/rs/zerolog/log"
)

var jsonMimeType string = "application/json"
var contentTypeHeader string = http.CanonicalHeaderKey("Content-Type")

// http localhost:8080/cameras
func (s *Server) handleCameras(resp http.ResponseWriter, req *http.Request) {
	cams, err := vid.DetectCams()
	if err != nil {
		log.Err(err).Send()

		resp.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(resp, err.Error())
		return
	}

	resp.Header().Add("content-type", "application/json")
	resp.WriteHeader(http.StatusOK)
	err = json.NewEncoder(resp).Encode(cams)
	if err != nil {
		log.Err(err).Send()
	}
}

// http POST localhost:8080/click X:=1 Y:=2
func (s *Server) handleClick(resp http.ResponseWriter, req *http.Request) {
	reqData := struct {
		X float64
		Y float64
	}{}

	if req.Header.Get(contentTypeHeader) != jsonMimeType {
		log.Warn().Msg("bad request (missing header)")

		resp.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(resp, "Missing content-type header")
		return
	}

	err := json.NewDecoder(req.Body).Decode(&reqData)
	if err != nil {
		log.Err(err).Msg("bad request")

		resp.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(resp, "Invalid request: %s", err.Error())
		return
	}

	log.Info().Interface("data", reqData).Msg("click")

	resp.WriteHeader(http.StatusNoContent)
}
