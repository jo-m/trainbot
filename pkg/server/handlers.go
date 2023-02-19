package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jo-m/trainbot/pkg/vid"
	"github.com/rs/zerolog/log"
)

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
