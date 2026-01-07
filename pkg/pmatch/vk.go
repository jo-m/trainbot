//go:build vk
// +build vk

package pmatch

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"errors"
	"image"
	"unsafe"

	"jo-m.ch/go/trainbot/pkg/vk"
)

//go:generate glslangValidator -V -o vk.spv vk.comp

//go:embed vk.spv
var shaderCode []byte

const (
	localSizeX = 4
	localSizeY = 4
	localSizeZ = 1
)

func bufsz(img *image.RGBA) int {
	return img.Bounds().Dy() * img.Stride
}

type results struct {
	MaxUInt uint32
	Max     float32
	MaxX    uint32
	MaxY    uint32
}

type pushConstants struct {
	imgStride uint32
	patStride uint32
}

func (r results) size() int {
	buf := bytes.Buffer{}
	err := binary.Write(&buf, binary.LittleEndian, results{})
	if err != nil {
		panic(err)
	}
	return buf.Len()
}

// SearchVk holds the state needed for patch matching search via Vulkan.
// Use NewSearchVk create an instance.
type SearchVk struct {
	searchRect    image.Rectangle
	h             *vk.Handle
	resSize       int
	pushConstants bytes.Buffer
	resultsBuf    *vk.Buffer
	imgBuf        *vk.Buffer
	patBuf        *vk.Buffer
	pipe          *vk.Pipe
}

// Destroy destroys a SearchVk and frees resources.
func (s *SearchVk) Destroy() {
	if s.pipe != nil {
		s.pipe.Destroy(s.h)
		s.pipe = nil
	}
	if s.patBuf != nil {
		s.patBuf.Destroy(s.h)
		s.patBuf = nil
	}
	if s.imgBuf != nil {
		s.imgBuf.Destroy(s.h)
		s.imgBuf = nil
	}
	if s.resultsBuf != nil {
		s.resultsBuf.Destroy(s.h)
		s.resultsBuf = nil
	}
	if s.h != nil {
		s.h.Destroy()
		s.h = nil
	}
}

// NewSearchVk creates a new instance of SearchVk.
// Destroy must be called to clean up.
func NewSearchVk(imgBounds, patBounds image.Rectangle, imgStride, patStride int, validate bool) (*SearchVk, error) {
	if patBounds.Size().X > imgBounds.Size().X ||
		patBounds.Size().Y > imgBounds.Size().Y {
		panic("patch too large")
	}

	s := SearchVk{}

	// Search rect in img coordinates.
	s.searchRect = image.Rectangle{
		Min: imgBounds.Min,
		Max: imgBounds.Max.Sub(patBounds.Size()).Add(image.Pt(1, 1)),
	}

	// Create instance.
	var err error
	s.h, err = vk.NewHandle(validate)
	if err != nil {
		return nil, err
	}

	// Prepare buffers.
	s.resSize = results{}.size()
	s.resultsBuf, err = s.h.NewBuffer(s.resSize)
	if err != nil {
		s.Destroy()
		return nil, err
	}

	s.imgBuf, err = s.h.NewBuffer(imgBounds.Dy() * imgStride)
	if err != nil {
		s.Destroy()
		return nil, err
	}

	s.patBuf, err = s.h.NewBuffer(patBounds.Dy() * patStride)
	if err != nil {
		s.Destroy()
		return nil, err
	}

	// Create pipe.
	binary.Write(&s.pushConstants, binary.LittleEndian, pushConstants{}) // #nosec G104: bytes.Buffer.Write{} always returns err = nil
	specInfo := []int{
		// Local size.
		localSizeX,
		localSizeY,
		1,
		// Dimensions constants.
		s.searchRect.Dx(),
		s.searchRect.Dy(),
		patBounds.Dx(),
		patBounds.Dy(),
	}
	s.pipe, err = s.h.NewPipe(shaderCode, []*vk.Buffer{s.resultsBuf, s.imgBuf, s.patBuf}, specInfo, s.pushConstants.Len())
	if err != nil {
		if err != nil {
			s.Destroy()
			return nil, err
		}
	}

	return &s, nil
}

// Run runs a search.
func (s *SearchVk) Run(img, pat *image.RGBA) (maxX, maxY int, maxCos float64, err error) {
	searchRect := image.Rectangle{
		Min: img.Bounds().Min,
		Max: img.Bounds().Max.Sub(pat.Bounds().Size()).Add(image.Pt(1, 1)),
	}
	if searchRect != s.searchRect {
		err = errors.New("img/pat size does not match state")
		return
	}

	// Reset results buffer to zero.
	err = s.resultsBuf.Zero(s.h, s.resSize)
	if err != nil {
		return
	}

	// Write to buffers.
	err = s.imgBuf.Write(s.h, unsafe.Pointer(&img.Pix[0]), bufsz(img)) // #nosec G103
	if err != nil {
		return
	}

	err = s.patBuf.Write(s.h, unsafe.Pointer(&pat.Pix[0]), bufsz(pat)) // #nosec G103
	if err != nil {
		return
	}

	// Prepare push constants.
	s.pushConstants.Reset()
	err = binary.Write(
		&s.pushConstants,
		binary.LittleEndian,
		pushConstants{uint32(img.Stride), uint32(pat.Stride)})
	if err != nil {
		return
	}

	// Run.
	err = s.pipe.Run(
		s.h,
		[3]uint{
			uint(s.searchRect.Dx()/localSizeX + 1),
			uint(s.searchRect.Dy()/localSizeY + 1),
			1,
		},
		s.pushConstants.Bytes())
	if err != nil {
		return
	}

	// Read results.
	res := results{}
	resMem := make([]byte, res.size())
	err = s.resultsBuf.Read(s.h, unsafe.Pointer(&resMem[0]), s.resSize) // #nosec G103
	if err != nil {
		return
	}
	err = binary.Read(bytes.NewReader(resMem), binary.LittleEndian, &res)
	if err != nil {
		return
	}
	return int(res.MaxX), int(res.MaxY), float64(res.Max), nil
}

// SearchRGBAVk is a wrapper which allocates a SearchVk, runs a search, and then destroys the instance again.
// Should only be used for testing, in real applications the instance should be reused.
// Deprecated: do not use, only for testing.
func SearchRGBAVk(img, pat *image.RGBA) (maxX, maxY int, maxCos float64) {
	h, err := NewSearchVk(img.Bounds(), pat.Bounds(), img.Stride, pat.Stride, false)
	if err != nil {
		panic(err)
	}
	defer h.Destroy()

	maxX, maxY, maxCos, err = h.Run(img, pat)
	if err != nil {
		panic(err)
	}
	return
}
