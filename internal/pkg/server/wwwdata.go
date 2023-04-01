package server

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const wwwDataDir = "wwwdata"

//go:embed wwwdata
var wwwData embed.FS

func getDataDir() (string, error) {
	_, srcFile, _, _ := runtime.Caller(0)
	srcDir := filepath.Dir(srcFile)

	try := []string{
		filepath.Join(srcDir, wwwDataDir),
		wwwDataDir,
	}

	var stat fs.FileInfo
	var err error
	for _, dir := range try {
		stat, err = os.Stat(dir)
		if err != nil {
			continue
		}
		if !stat.IsDir() {
			err = fmt.Errorf("is not a directory: '%s'", dir)
			continue
		}

		return dir, nil
	}

	return "", err
}

func getDataRoot(embed bool) (http.FileSystem, error) {
	if embed {
		data, err := fs.Sub(wwwData, wwwDataDir)
		if err != nil {
			return nil, err
		}
		return http.FS(data), nil
	}

	dir, err := getDataDir()
	if err != nil {
		return nil, err
	}
	return http.Dir(dir), nil
}
