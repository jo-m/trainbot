package vid

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// Copied from the stdlib jpeg package.
const (
	rst0Marker = 0xd0 // ReSTart (0).
	rst7Marker = 0xd7 // ReSTart (7).
	soiMarker  = 0xd8 // Start Of Image.
	eoiMarker  = 0xd9 // End Of Image.
	sosMarker  = 0xda // Start Of Scan.
)

// JPEGScanner works similar to bufio.Scanner but on JPEG images.
// Useful to pull frame after frame out of a MJPEG stream.
type JPEGScanner struct {
	r   *bufio.Reader
	buf []byte
}

// NewJPEGScanner creates a new JPEGScanner.
// Responsibility to close reader remains with caller.
func NewJPEGScanner(r io.Reader) *JPEGScanner {
	return &JPEGScanner{
		r:   bufio.NewReader(r),
		buf: make([]byte, 0),
	}
}

// resetBuf resets the underlying buffer to length 0 while preserving its capacity.
func (s *JPEGScanner) resetBuf() {
	s.buf = s.buf[:0]
}

// readBytes reads n bytes into the buffer, growing capacity if needed, and returns a slice of the bytes read.
func (s *JPEGScanner) readBytes(n int) ([]byte, error) {
	// Grow capacity if necessary, and grow length.
	oldLen := len(s.buf)
	newLen := oldLen + n
	if oldLen+n > cap(s.buf) {
		s.buf = append(s.buf, make([]byte, n)...)
	} else {
		s.buf = s.buf[:newLen]
	}

	// Read until buffer is full.
	pos := oldLen
	for pos < newLen {
		read, err := s.r.Read(s.buf[pos:newLen])
		if err != nil {
			return nil, err
		}
		pos += read
	}

	return s.buf[oldLen:newLen], nil
}

func (s *JPEGScanner) scanImageData() error {
	for {
		marker, err := s.r.Peek(2)
		if err != nil && len(marker) == 0 {
			return err
		}

		if len(marker) == 1 {
			if marker[0] == 0xFF {
				return errors.New("invalid data")
			}

			// Advance 1.
			_, err := s.readBytes(1)
			if err != nil {
				return err
			}
			continue
		}

		if marker[0] == 0xFF {
			if marker[1] == 0 {
				// Advance 2.
				_, err := s.readBytes(2)
				if err != nil {
					return err
				}
				continue
			}

			if marker[1] >= rst0Marker && marker[1] <= rst7Marker {
				// Advance 2.
				_, err := s.readBytes(2)
				if err != nil {
					return err
				}
				continue
			}

			// Hand back control.
			return nil
		}

		// Advance 1.
		_, err = s.readBytes(1)
		if err != nil {
			return err
		}
		continue
	}
}

// Scan reads until it has read an entire JPEG image, and returns the buffer containing its data.
// https://github.com/corkami/formats/blob/master/image/jpeg.md.
func (s *JPEGScanner) Scan() ([]byte, error) {
	s.resetBuf()

	soi, err := s.readBytes(2)
	if err != nil {
		return nil, fmt.Errorf("could not read soi: %w", err)
	}
	if soi[0] != 0xFF || soi[1] != soiMarker {
		return nil, fmt.Errorf("invalid soi found: 0x%s", hex.EncodeToString(soi[0:2]))
	}

	for {
		marker, err := s.readBytes(2)
		if err != nil {
			return nil, fmt.Errorf("could not read segment marker: %w", err)
		}
		if marker[0] != 0xFF {
			return nil, fmt.Errorf("invalid segment marker 0x%x", marker[1])
		}

		// Handle the segment types without length field.
		if marker[1] == eoiMarker {
			break
		}
		if marker[1] >= rst0Marker && marker[1] <= rst7Marker {
			return nil, errors.New("restart marker at invalid position")
		}

		segLenB, err := s.readBytes(2)
		if err != nil {
			return nil, fmt.Errorf("could not read segment length: %w", err)
		}
		segLen := uint16(segLenB[0])<<8 + uint16(segLenB[1]) // Includes includes length field itself but not the marker.

		// Read rest of the segment, we do not parse it.
		_, err = s.readBytes(int(segLen) - 2) // The length we have already read.
		if err != nil {
			return nil, fmt.Errorf("could not read segment data: %w", err)
		}

		if marker[1] == sosMarker {
			err := s.scanImageData()
			if err != nil {
				return nil, fmt.Errorf("failed to read image data: %w", err)
			}
		}
	}

	return s.buf, nil
}
