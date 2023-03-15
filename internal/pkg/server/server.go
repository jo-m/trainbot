package server

import (
	"bytes"
	"image"
	"image/jpeg"
	"net/http"

	"github.com/mattn/go-mjpeg"
)

type Server struct {
	mux    *http.ServeMux
	stream *mjpeg.Stream
}

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

func (s *Server) GetMux() *http.ServeMux {
	return s.mux
}

func (s *Server) SetFrame(frame image.Image) error {
	buf := bytes.Buffer{}
	err := jpeg.Encode(&buf, frame, nil)
	if err != nil {
		return err
	}
	return s.stream.Update(buf.Bytes())
}

func (s *Server) SetFrameRawJPEG(frame []byte) error {
	return s.stream.Update(frame)
}
