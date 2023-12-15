package pmatch

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"errors"
	"image"
	"unsafe"

	"github.com/jo-m/trainbot/pkg/vk"
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

func (r results) size() int {
	buf := bytes.Buffer{}
	err := binary.Write(&buf, binary.LittleEndian, results{})
	if err != nil {
		panic(err)
	}
	return buf.Len()
}

type SearchVk struct {
	searchRect image.Rectangle
	h          *vk.Handle
	search     *image.RGBA
	resSize    int
	resultsBuf *vk.Buffer
	imgBuf     *vk.Buffer
	patBuf     *vk.Buffer
	pipe       *vk.Pipe
}

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

func NewSearchVk(imgBounds, patBounds image.Rectangle, imgStride, patStride int) (*SearchVk, error) {
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
	s.search = image.NewRGBA(s.searchRect)

	// Create instance.
	var err error
	s.h, err = vk.NewHandle(true)
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
		imgStride,
		patStride,
		s.search.Stride,
	}
	s.pipe, err = s.h.NewPipe(shaderCode, []*vk.Buffer{s.resultsBuf, s.imgBuf, s.patBuf}, specInfo)
	if err != nil {
		if err != nil {
			s.Destroy()
			return nil, err
		}
	}

	return &s, nil
}

func (s *SearchVk) Run(img, pat *image.RGBA) (maxX, maxY int, maxCos float64, err error) {
	searchRect := image.Rectangle{
		Min: img.Bounds().Min,
		Max: img.Bounds().Max.Sub(pat.Bounds().Size()).Add(image.Pt(1, 1)),
	}
	if searchRect != s.searchRect {
		err = errors.New("img/pat size does not match state")
		return
	}

	// Write to buffers.
	err = s.imgBuf.Write(s.h, unsafe.Pointer(&img.Pix[0]), bufsz(img))
	if err != nil {
		return
	}

	err = s.patBuf.Write(s.h, unsafe.Pointer(&pat.Pix[0]), bufsz(pat))
	if err != nil {
		return
	}

	// Run.
	err = s.pipe.Run(s.h, [3]uint{
		uint(s.searchRect.Dx()/localSizeX + 1),
		uint(s.searchRect.Dy()/localSizeY + 1),
		1,
	})
	if err != nil {
		return
	}

	// Read results.
	res := results{}
	resMem := make([]byte, res.size())
	err = s.resultsBuf.Read(s.h, unsafe.Pointer(&resMem[0]), s.resSize)
	if err != nil {
		return
	}
	binary.Read(bytes.NewReader(resMem), binary.LittleEndian, &res)
	return int(res.MaxX), int(res.MaxY), float64(res.Max), nil
}

func SearchRGBAVk(img, pat *image.RGBA) (maxX, maxY int, maxCos float64) {
	h, err := NewSearchVk(img.Bounds(), pat.Bounds(), img.Stride, pat.Stride)
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
