package vid

import (
	"image"
	"strings"
	"time"
)

type Src interface {
	// GetFrame retrieves the next frame.
	// Note that the underlying image buffer remains owned by the video source,
	// it must not be changed by the caller and will be overwritten on the next
	// invocation.
	// Returns io.EOF after the last frame, after which Close() should be called
	// on the instance before discarding it.
	GetFrame() (*image.RGBA, *time.Time, error)
	Close() error
}

func NewSrc(path string) (Src, error) {
	if strings.HasPrefix(path, "/dev/video") {
		panic("not implemented") // TODO
	}

	return NewFileSrc(path, false)
}
