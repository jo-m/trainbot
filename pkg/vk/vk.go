//go:build vk
// +build vk

package vk

// See pkg/pmatch/c.go for docs on CC flags.

// #cgo CFLAGS: -Wall -Werror -Wextra -pedantic -std=c99
// #cgo CFLAGS: -O2
// #cgo LDFLAGS: -lvulkan
//
// #cgo amd64 CFLAGS: -march=x86-64 -mtune=generic
//
// #cgo arm64 CFLAGS: -mcpu=cortex-a72 -mtune=cortex-a72
//
// #include "vk.h"
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

// Package errors.
var (
	ErrNotLittleEndian      = errors.New("only little endian supported for now")
	ErrNeedAtLeastOneBuffer = errors.New("need at least one buffer")
)

// Result is a wrapper around VkResult from <vulkan.h>, implementing the Go Error interface.
type Result struct {
	vkResult C.VkResult
}

// Error implements Error.
func (r *Result) Error() string {
	return fmt.Sprintf("VkResult: %d", r.vkResult)
}

// Num retrieves the embedded VkResult as int.
func (r *Result) Num() int {
	return int(r.vkResult)
}

// Handle is a Vulkan instance and device.
// Use NewHandle() to create an instance.
type Handle struct {
	valid bool
	h     C.vk_handle
}

func (h *Handle) ptr() *C.vk_handle {
	if !h.valid {
		panic("invalid instance")
	}

	return (*C.vk_handle)(&h.h)
}

// NewHandle creates a new handle.
// Must call Destroy() on the instance after usage.
func NewHandle(enableValidation bool) (*Handle, error) {
	if C.check_big_endian() != C.VK_SUCCESS {
		return nil, ErrNotLittleEndian
	}

	ret := Handle{valid: true}
	result := C.create_vk_handle(
		ret.ptr(),
		(C.bool)(enableValidation),
	)

	if result != C.VK_SUCCESS {
		return nil, fmt.Errorf("creating instance failed: %w", &Result{result})
	}

	return &ret, nil
}

// Destroy destroys a handle.
func (h *Handle) Destroy() {
	if !h.valid {
		panic("invalid instance")
	}

	C.vk_handle_destroy(h.ptr())
	h.valid = false
}

// GetDeviceString returns a string describing the underlying device.
func (h *Handle) GetDeviceString() string {
	if !h.valid {
		panic("invalid instance")
	}

	str := make([]C.char, 512)
	l := C.vk_handle_get_device_string(h.ptr(), (*C.char)(&str[0]), C.size_t(len(str)))

	return C.GoStringN((*C.char)(&str[0]), l)
}

// Buffer is a VkBuffer with associated memory.
// Use NewBuffer() to create an instance.
type Buffer struct {
	valid bool
	b     C.vk_buffer
}

func (b *Buffer) ptr() *C.vk_buffer {
	if !b.valid {
		panic("invalid instance")
	}

	return (*C.vk_buffer)(&b.b)
}

// NewBuffer creates a new buffer and allocates size bytes of memory for it.
// Must call Destroy() on the instance after usage.
func (h *Handle) NewBuffer(size int) (*Buffer, error) {
	if !h.valid {
		panic("invalid instance")
	}

	ret := Buffer{valid: true}
	result := C.create_vk_buffer(
		h.ptr(),
		ret.ptr(),
		C.size_t(size),
	)

	if result != C.VK_SUCCESS {
		return nil, fmt.Errorf("creating buffer failed: %w", &Result{result})
	}

	return &ret, nil
}

// Destroy destroys a buffer.
func (b *Buffer) Destroy(h *Handle) {
	if !b.valid {
		panic("invalid instance")
	}
	if !h.valid {
		panic("invalid instance")
	}

	C.vk_buffer_destroy(h.ptr(), b.ptr())
	b.valid = false
}

// Write writes data from a memory location to this buffer.
func (b *Buffer) Write(h *Handle, src unsafe.Pointer, size int) error {
	if !b.valid {
		panic("invalid instance")
	}
	if !h.valid {
		panic("invalid instance")
	}

	result := C.vk_buffer_write(h.ptr(), b.ptr(), src, C.size_t(size))

	if result != C.VK_SUCCESS {
		return fmt.Errorf("writing to buffer failed: %w", &Result{result})
	}
	return nil
}

// Read reads data from this buffer to a memory location.
func (b *Buffer) Read(h *Handle, dst unsafe.Pointer, size int) error {
	if !b.valid {
		panic("invalid instance")
	}
	if !h.valid {
		panic("invalid instance")
	}

	result := C.vk_buffer_read(h.ptr(), dst, b.ptr(), C.size_t(size))

	if result != C.VK_SUCCESS {
		return fmt.Errorf("reading from buffer failed: %w", &Result{result})
	}
	return nil
}

// BufSz returns the size in bytes for a numeric slice.
func BufSz[T float32 | int32 | uint32](buf []T) int {
	return int(unsafe.Sizeof(buf[0])) * len(buf)
}

// BufWrite is a typed wrapper around Buffer.Write().
func BufWrite[T float32 | int32 | uint32](h *Handle, dst *Buffer, src []T) error {
	return dst.Write(h, unsafe.Pointer(&src[0]), int(unsafe.Sizeof(src[0]))*len(src))
}

// BufRead is a typed wrapper around Buffer.Read().
func BufRead[T float32 | int32 | uint32](h *Handle, dst []T, src *Buffer) error {
	return src.Read(h, unsafe.Pointer(&dst[0]), int(unsafe.Sizeof(dst[0]))*len(dst))
}

// Pipe is a compute pipeline.
// Use NewPipe() to create an instance.
type Pipe struct {
	valid bool
	p     C.vk_pipe
}

func (p *Pipe) ptr() *C.vk_pipe {
	if !p.valid {
		panic("invalid instance")
	}

	return (*C.vk_pipe)(&p.p)
}

// NewPipe creates a new pipe.
// Must call Destroy() on the instance after usage.
// The shader code must be valid SPIR-V bytecode, and its size a multiple of 4.
// Buffers bufs will be bound to the shader program as VK_DESCRIPTOR_TYPE_STORAGE_BUFFERs in slice order.
// Specialization constants can be only int for simplicity.
// A single push constants buffer can be optionally specified (pushConstantRangeSz > 0).
func (h *Handle) NewPipe(shader []byte, bufs []*Buffer, specConstants []int, pushConstantRangeSz int) (*Pipe, error) {
	if !h.valid {
		panic("invalid instance")
	}

	if len(bufs) == 0 {
		return nil, ErrNeedAtLeastOneBuffer
	}

	// Buffers and descriptor types.
	cBuffers := make([]C.vk_buffer, len(bufs))
	descTypes := make([]C.VkDescriptorType, len(bufs))
	for i, buf := range bufs {
		cBuffers[i] = buf.b
		descTypes[i] = C.VK_DESCRIPTOR_TYPE_STORAGE_BUFFER
	}

	// Specialization constants.
	var specData []C.int32_t
	var specInfo *C.VkSpecializationInfo
	if len(specConstants) > 0 {
		specData = make([]C.int32_t, len(specConstants))
		for i, val := range specConstants {
			specData[i] = C.int32_t(val)
		}

		specInfo = C.alloc_int32_spec_info((*C.int32_t)(&specData[0]), C.uint32_t(len(specData)))
		defer C.free(unsafe.Pointer(specInfo))
	} else {
		specInfo = &C.VkSpecializationInfo{}
	}

	// Push constants.
	pushConstants := C.VkPushConstantRange{
		stageFlags: C.VK_SHADER_STAGE_COMPUTE_BIT,
		offset:     0,
		size:       C.uint32_t(pushConstantRangeSz),
	}

	// Create pipeline.
	ret := Pipe{valid: true}
	result := C.create_vk_pipe(
		h.ptr(),
		ret.ptr(),
		(*C.uint8_t)(&shader[0]),
		(C.size_t)(len(shader)),
		(*C.vk_buffer)(&cBuffers[0]),
		(*C.VkDescriptorType)(&descTypes[0]),
		C.uint32_t(len(descTypes)),
		*specInfo,
		pushConstants,
	)

	if result != C.VK_SUCCESS {
		return nil, fmt.Errorf("creating pipeline failed: %w", &Result{result})
	}

	return &ret, nil
}

// Destroy destroys a pipe.
func (p *Pipe) Destroy(h *Handle) {
	if !p.valid {
		panic("invalid instance")
	}
	if !h.valid {
		panic("invalid instance")
	}

	C.vk_pipe_destroy(h.ptr(), p.ptr())
	p.valid = false
}

// Run runs a compute pipeline with the given workgroup size.
func (p *Pipe) Run(h *Handle, workgroupSize [3]uint, pushConstantBuf []byte) error {
	if !p.valid {
		panic("invalid instance")
	}
	if !h.valid {
		panic("invalid instance")
	}

	dims := C.dim3{
		C.uint32_t(workgroupSize[0]),
		C.uint32_t(workgroupSize[1]),
		C.uint32_t(workgroupSize[2]),
	}

	var pushConstantPtr *byte
	if len(pushConstantBuf) > 0 {
		pushConstantPtr = &pushConstantBuf[0]
	}

	result := C.vk_pipe_run(
		h.ptr(),
		p.ptr(),
		dims,
		(*C.uint8_t)(pushConstantPtr),
		C.size_t(len(pushConstantBuf)))

	if result != C.VK_SUCCESS {
		return fmt.Errorf("running pipeline failed: %w", &Result{result})
	}
	return nil
}
