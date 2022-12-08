package rec

import (
	"encoding/json"
	"fmt"
	"image"
	"io"
	"os"
	"path"
	"time"

	"github.com/jo-m/trainbot/internal/pkg/imutil"
	"github.com/jo-m/trainbot/pkg/vid"
)

type Reader struct {
	dirPath string
	meta    []frameMeta
	count   int
}

// compile time interface check
var _ vid.Src = (*Reader)(nil)

func NewReader(dirPath string) (*Reader, error) {
	metaPath := path.Join(dirPath, metaFileName)
	metaF, err := os.Open(metaPath)
	if err != nil {
		return nil, fmt.Errorf("error opening '%s', is this a valid recording dir?: %w", metaPath, err)
	}
	defer metaF.Close()

	ret := Reader{
		dirPath: dirPath,
	}
	err = json.NewDecoder(metaF).Decode(&ret.meta)
	if err != nil {
		return nil, fmt.Errorf("unable to parse metadata file: %w", err)
	}

	return &ret, nil
}

func (r *Reader) GetFrame() (image.Image, *time.Time, error) {
	if r.count >= len(r.meta) {
		return nil, nil, io.EOF
	}

	defer func() { r.count++ }()

	meta := r.meta[r.count]
	path := path.Join(r.dirPath, meta.FileName)
	img, err := imutil.Load(path)
	if err != nil {
		return nil, nil, err
	}

	return img, &meta.TimeUTC, nil
}

// GetFPS implements Src.
func (r *Reader) GetFPS() float64 {
	n := len(r.meta)
	return float64(n) / r.meta[n-1].TimeUTC.Sub(r.meta[0].TimeUTC).Seconds()
}

func (r *Reader) Close() error { return nil }
