package pmatch

import (
	"bytes"
	_ "embed"
	"encoding/binary"
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

func SearchRGBAVk(img, pat *image.RGBA) (maxX, maxY int, maxCos float64) {
	if pat.Bounds().Size().X > img.Bounds().Size().X ||
		pat.Bounds().Size().Y > img.Bounds().Size().Y {
		panic("patch too large")
	}

	// Search rect in img coordinates.
	searchRect := image.Rectangle{
		Min: img.Bounds().Min,
		Max: img.Bounds().Max.Sub(pat.Bounds().Size()).Add(image.Pt(1, 1)),
	}
	search := image.NewRGBA(searchRect)

	// Prepare buffers.
	h, err := vk.NewHandle(true)
	if err != nil {
		panic(err)
	}
	defer h.Destroy()

	res := results{}
	resMem := bytes.Buffer{}
	err = binary.Write(&resMem, binary.LittleEndian, res)
	if err != nil {
		panic(err)
	}
	resultsBuf, err := h.NewBuffer(resMem.Len())
	if err != nil {
		panic(err)
	}
	defer resultsBuf.Destroy(h)

	imgBuf, err := h.NewBuffer(bufsz(img))
	if err != nil {
		panic(err)
	}
	defer imgBuf.Destroy(h)
	err = imgBuf.Write(h, unsafe.Pointer(&img.Pix[0]), bufsz(img))
	if err != nil {
		panic(err)
	}

	patBuf, err := h.NewBuffer(bufsz(pat))
	if err != nil {
		panic(err)
	}
	defer patBuf.Destroy(h)
	err = patBuf.Write(h, unsafe.Pointer(&pat.Pix[0]), bufsz(pat))
	if err != nil {
		panic(err)
	}

	searchBuf, err := h.NewBuffer(bufsz(search))
	if err != nil {
		panic(err)
	}
	defer searchBuf.Destroy(h)

	// Create pipe.
	specInfo := []int{
		// Local size.
		localSizeX,
		localSizeY,
		1,
		// Dimensions constants.
		searchRect.Dx(),
		searchRect.Dy(),
		pat.Bounds().Dx(),
		pat.Bounds().Dy(),
		img.Stride,
		pat.Stride,
		search.Stride,
	}
	p, err := h.NewPipe(shaderCode, []*vk.Buffer{resultsBuf, imgBuf, patBuf, searchBuf}, specInfo)
	if err != nil {
		panic(err)
	}
	defer p.Destroy(h)

	// Run.
	err = p.Run(h, [3]uint{
		uint(searchRect.Dx()/localSizeX + 1),
		uint(searchRect.Dy()/localSizeY + 1),
		1,
	})
	if err != nil {
		panic(err)
	}

	// Read results.
	err = resultsBuf.Read(h, unsafe.Pointer(&resMem.Bytes()[0]), resMem.Len())
	if err != nil {
		panic(err)
	}
	binary.Read(&resMem, binary.LittleEndian, &res)
	return int(res.MaxX), int(res.MaxY), float64(res.Max)
}
