package vid

import (
	"image"
	"time"
)

// Src describes a frame source.
type Src interface {
	// GetFrame retrieves the next frame.
	// Note that the underlying image buffer remains owned by the video source,
	// it must not be changed by the caller and might be overwritten on the next
	// invocation.
	// Returns io.EOF after the last frame, after which Close() should be called
	// on the instance before discarding it.
	GetFrame() (image.Image, *time.Time, error)

	// GetFrame retrieves the next frame in the raw pixel format of the source.
	// Not all sources might implement this.
	GetFrameRaw() ([]byte, FourCC, *time.Time, error)

	// IsLive returns if the src is a live source (e.g. camera).
	IsLive() bool

	// GetFPS returns the current frame rate of this source.
	GetFPS() float64

	// Close closes the frame source and frees resources.
	Close() error
}
