package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/jo-m/trainbot/pkg/vid"
	"github.com/mattn/go-mjpeg"
)

//go:embed wwwdata
var wwwData embed.FS

func handleClick(resp http.ResponseWriter, req *http.Request) {
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

func buildServer(strm *mjpeg.Stream) *http.ServeMux {
	data, err := fs.Sub(wwwData, "wwwdata")
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/stream.mjpeg", strm.ServeHTTP)
	mux.HandleFunc("/click", handleClick)
	mux.Handle("/", http.FileServer(http.FS(data)))

	return mux
}

func main() {
	configs, err := vid.DetectCams()
	if err != nil {
		panic(err)
	}
	if len(configs) == 0 {
		panic("no cameras detected")
	}

	src, err := vid.NewCamSrc(configs[0])
	if err != nil {
		panic(err)
	}

	strm := mjpeg.NewStream()
	mux := buildServer(strm)

	// Copies the stream from camera to HTTP handler.
	go func() {
		for i := uint64(0); ; i++ {
			frame, fmt, _, err := src.GetFrameRaw()
			if err != nil {
				panic(err)
			}
			if fmt != "MJPG" {
				panic("invalid stream format")
			}

			// Drop frames.
			if i%3 == 0 {
				strm.Update(frame)
			}
		}
	}()

	http.ListenAndServe(":8080", mux)
}
