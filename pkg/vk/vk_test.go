package vk

import (
	_ "embed"
	"strings"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:generate glslangValidator -V -o testfiles/minimal.spv testfiles/minimal.comp

//go:embed testfiles/minimal.spv
var shaderMinimal []byte

func Test_CreateDestroyHandle(t *testing.T) {
	h, err := NewHandle(true)
	assert.NoError(t, err)
	h.Destroy()

	h, err = NewHandle(false)
	assert.NoError(t, err)
	h.Destroy()
}

func Test_GetDeviceString(t *testing.T) {
	h, err := NewHandle(true)
	require.NoError(t, err)
	defer h.Destroy()

	s := h.GetDeviceString()
	assert.True(t, strings.HasPrefix(s, "name="))
}

func Test_CreateDestroyBuffer(t *testing.T) {
	h, err := NewHandle(true)
	require.NoError(t, err)
	defer h.Destroy()

	b, err := h.NewBuffer(1024)
	assert.NoError(t, err)
	b.Destroy(h)
}

func Test_CreateDestroyPipe(t *testing.T) {
	h, err := NewHandle(true)
	require.NoError(t, err)
	defer h.Destroy()

	bufs := make([]*Buffer, 3)
	for i := range bufs {
		var err error
		bufs[i], err = h.NewBuffer(1024)
		require.NoError(t, err)

		defer bufs[i].Destroy(h)
	}

	specInfo := []int{1, 2}

	p, err := h.NewPipe(shaderMinimal, bufs, specInfo)
	assert.NoError(t, err)
	p.Destroy(h)
}

func Test_CreateDestroyPipe_NoSpecInfo(t *testing.T) {
	h, err := NewHandle(true)
	require.NoError(t, err)
	defer h.Destroy()

	bufs := make([]*Buffer, 3)
	for i := range bufs {
		var err error
		bufs[i], err = h.NewBuffer(1024)
		require.NoError(t, err)

		defer bufs[i].Destroy(h)
	}

	specInfo := []int{}

	p, err := h.NewPipe(shaderMinimal, bufs, specInfo)
	assert.NoError(t, err)
	p.Destroy(h)
}

func Test_CreateDestroyPipe_NoBuffers(t *testing.T) {
	h, err := NewHandle(true)
	require.NoError(t, err)
	defer h.Destroy()

	specInfo := []int{1, 2}

	p, err := h.NewPipe(shaderMinimal, []*Buffer{}, specInfo)
	assert.Nil(t, p)
	assert.Error(t, err, ErrNeedAtLeastOneBuffer)
}

func Test_ReadWriteBuffer(t *testing.T) {
	h, err := NewHandle(true)
	require.NoError(t, err)
	defer h.Destroy()

	const size = 128

	buf, err := h.NewBuffer(size)
	require.NoError(t, err)
	defer buf.Destroy(h)

	src := make([]byte, size)
	dst := make([]byte, size)
	for i := range src {
		src[i] = 0xAB
		dst[i] = 0
	}

	err = buf.Write(h, unsafe.Pointer(&src[0]), size)
	assert.NoError(t, err)

	err = buf.Read(h, unsafe.Pointer(&dst[0]), size)
	assert.NoError(t, err)

	for _, val := range src {
		assert.Equal(t, byte(0xAB), val)
	}
}

func Test_ReadWriteBufferT(t *testing.T) {
	h, err := NewHandle(true)
	require.NoError(t, err)
	defer h.Destroy()

	const (
		l = 128
	)

	src := make([]int32, l)
	dst := make([]int32, l)
	for i := range src {
		src[i] = 123456
		dst[i] = 0
	}

	buf, err := h.NewBuffer(BufSz(src))
	require.NoError(t, err)
	defer buf.Destroy(h)

	err = BufWrite(h, buf, src)
	assert.NoError(t, err)

	err = BufRead(h, src, buf)
	assert.NoError(t, err)

	for _, val := range src {
		assert.Equal(t, int32(123456), val)
	}
}

func Test_RunPipe(t *testing.T) {
	h, err := NewHandle(true)
	require.NoError(t, err)
	defer h.Destroy()

	const (
		sizeX       = 200
		sizeY       = 534
		bufVal      = 100
		constVal    = 10
		localSizeXY = 16
	)

	// Prepare buffers.
	buf0Data := make([]int32, sizeX*sizeY)
	buf1Data := make([]int32, sizeX*sizeY)
	buf2Data := make([]int32, sizeX*sizeY)
	for i := range buf0Data {
		buf0Data[i] = int32(i)
		buf1Data[i] = bufVal
		buf2Data[i] = 0
	}

	buf0, err := h.NewBuffer(BufSz(buf0Data))
	require.NoError(t, err)
	defer buf0.Destroy(h)
	err = BufWrite(h, buf0, buf0Data)
	require.NoError(t, err)

	buf1, err := h.NewBuffer(BufSz(buf1Data))
	require.NoError(t, err)
	defer buf1.Destroy(h)
	err = BufWrite(h, buf1, buf1Data)
	require.NoError(t, err)

	buf2, err := h.NewBuffer(BufSz(buf2Data))
	require.NoError(t, err)
	defer buf2.Destroy(h)
	err = BufWrite(h, buf2, buf2Data)
	require.NoError(t, err)

	// Create pipe.
	specInfo := []int{
		localSizeXY,
		localSizeXY,
		1,
		sizeX,
		sizeY,
		constVal,
	}
	p, err := h.NewPipe(shaderMinimal, []*Buffer{buf0, buf1, buf2}, specInfo)
	assert.NoError(t, err)
	defer p.Destroy(h)

	// Run.
	err = p.Run(h, [3]uint{sizeX/localSizeXY + 1, sizeY/localSizeXY + 1, 1})
	assert.NoError(t, err)

	err = BufRead(h, buf2Data, buf2)
	require.NoError(t, err)

	// Check.
	for i := range buf2Data {
		assert.Equal(t, buf0Data[i]+buf1Data[i]+constVal, buf2Data[i])
	}
}
