package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"

	"github.com/jo-m/trainbot/pkg/vid"
	"github.com/mattn/go-mjpeg"
	"github.com/rs/zerolog/log"
)

// Server is a simple MJPEG HTTP streaming server.
// Use NewServer to initiate an instance.
type Server struct {
	mux    *http.ServeMux
	stream *mjpeg.Stream
}

// NewServer creates a new server.
func NewServer(embed bool) (*Server, error) {
	mux := http.NewServeMux()
	s := Server{
		mux: mux,

		stream: mjpeg.NewStream(),
	}

	wwwRoot, err := getDataRoot(embed)
	if err != nil {
		return nil, err
	}

	mux.HandleFunc("/cameras", s.handleCameras)
	mux.HandleFunc("/stream.mjpeg", s.stream.ServeHTTP)

	mux.Handle("/", http.FileServer(wwwRoot))

	return &s, nil
}

// handleCameras detects v4l cameras and returns them as JSON.
// Test via
//
//	http localhost:8080/cameras
func (s *Server) handleCameras(resp http.ResponseWriter, _ *http.Request) {
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

// GetMux returns the router.
func (s *Server) GetMux() *http.ServeMux {
	return s.mux
}

// SetFrame encodes an image as JPEG and streams it.
func (s *Server) SetFrame(frame image.Image) error {
	buf := bytes.Buffer{}
	err := jpeg.Encode(&buf, frame, nil)
	if err != nil {
		return err
	}
	return s.stream.Update(buf.Bytes())
}

// SetFrameRawJPEG streams a raw encoded JPEG frame.
func (s *Server) SetFrameRawJPEG(frame []byte) error {
	cp := append([]byte(nil), frame...)
	return s.stream.Update(cp)
}
