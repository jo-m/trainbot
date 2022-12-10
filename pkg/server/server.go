package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Server struct {
	mux *http.ServeMux
}

func NewServer(embed bool) (*Server, error) {
	mux := http.NewServeMux()
	s := Server{
		mux: mux,
	}

	wwwRoot, err := getDataRoot(embed)
	if err != nil {
		return nil, err
	}

	mux.HandleFunc("/click", s.handleClick)
	mux.Handle("/", http.FileServer(wwwRoot))

	return &s, nil
}

func (s *Server) GetMux() *http.ServeMux {
	return s.mux
}

func (s *Server) handleClick(resp http.ResponseWriter, req *http.Request) {
	reqData := struct {
		X float64
		Y float64
	}{}

	err := json.NewDecoder(req.Body).Decode(&reqData)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(fmt.Sprintf("Invalid request: %s", err.Error())))
		return
	}

	fmt.Println("click", reqData)
	resp.WriteHeader(http.StatusNoContent)
}
