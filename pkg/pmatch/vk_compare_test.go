//go:build vk
// +build vk

package pmatch

import (
	"image"
	"image/color"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"jo-m.ch/go/trainbot/pkg/imutil"
)

// Test_SearchRGBA_VkVsC compares Vulkan and C implementations on identical inputs.
// This test exposes bugs where the two implementations produce different results.
func Test_SearchRGBA_VkVsC(t *testing.T) {
	// Test with the standard bird image
	img := imutil.ToRGBA(LoadTestImg())
	pat, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	if err != nil {
		t.Fatal(err)
	}
	patCopy := imutil.ToRGBA(pat.(*image.RGBA))

	// Compare exact match
	xC, yC, scoreC := SearchRGBAC(img, patCopy)
	xVk, yVk, scoreVk := SearchRGBAVk(img, patCopy)

	t.Logf("Exact match - C: (%d, %d) score=%.10f, Vk: (%d, %d) score=%.10f",
		xC, yC, scoreC, xVk, yVk, scoreVk)

	assert.Equal(t, xC, xVk, "x position should match")
	assert.Equal(t, yC, yVk, "y position should match")
	assert.InDelta(t, scoreC, scoreVk, 1e-5, "scores should be close")
}

// Test_SearchRGBA_VkVsC_RandomPattern tests with a random pattern that doesn't exist in the image.
// This is more likely to expose boundary and scoring bugs.
func Test_SearchRGBA_VkVsC_RandomPattern(t *testing.T) {
	img := imutil.ToRGBA(LoadTestImg())

	// Create a random pattern
	rng := rand.New(rand.NewSource(42))
	patW, patH := 25, 20
	pat := image.NewRGBA(image.Rect(0, 0, patW, patH))
	for y := 0; y < patH; y++ {
		for x := 0; x < patW; x++ {
			pat.SetRGBA(x, y, color.RGBA{
				R: uint8(rng.Intn(256)),
				G: uint8(rng.Intn(256)),
				B: uint8(rng.Intn(256)),
				A: 255,
			})
		}
	}

	xC, yC, scoreC := SearchRGBAC(img, pat)
	xVk, yVk, scoreVk := SearchRGBAVk(img, pat)

	t.Logf("Random pattern - C: (%d, %d) score=%.10f, Vk: (%d, %d) score=%.10f",
		xC, yC, scoreC, xVk, yVk, scoreVk)

	assert.Equal(t, xC, xVk, "x position should match for random pattern")
	assert.Equal(t, yC, yVk, "y position should match for random pattern")
	assert.InDelta(t, scoreC, scoreVk, 1e-5, "scores should be close for random pattern")
}

// Test_SearchRGBA_VkVsC_SmallImage tests with a small image to stress boundary conditions.
func Test_SearchRGBA_VkVsC_SmallImage(t *testing.T) {
	// Create a small image and pattern
	imgW, imgH := 16, 16
	patW, patH := 4, 4

	rng := rand.New(rand.NewSource(123))

	img := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
	for y := 0; y < imgH; y++ {
		for x := 0; x < imgW; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(rng.Intn(256)),
				G: uint8(rng.Intn(256)),
				B: uint8(rng.Intn(256)),
				A: 255,
			})
		}
	}

	pat := image.NewRGBA(image.Rect(0, 0, patW, patH))
	for y := 0; y < patH; y++ {
		for x := 0; x < patW; x++ {
			pat.SetRGBA(x, y, color.RGBA{
				R: uint8(rng.Intn(256)),
				G: uint8(rng.Intn(256)),
				B: uint8(rng.Intn(256)),
				A: 255,
			})
		}
	}

	xC, yC, scoreC := SearchRGBAC(img, pat)
	xVk, yVk, scoreVk := SearchRGBAVk(img, pat)

	t.Logf("Small image - C: (%d, %d) score=%.10f, Vk: (%d, %d) score=%.10f",
		xC, yC, scoreC, xVk, yVk, scoreVk)

	// Check search rectangle bounds
	// searchRect.Dx() = imgW - patW + 1 = 16 - 4 + 1 = 13
	// Valid x: 0..12, Valid y: 0..12
	maxValidX := imgW - patW
	maxValidY := imgH - patH

	t.Logf("Valid range: x in [0, %d], y in [0, %d]", maxValidX, maxValidY)

	assert.LessOrEqual(t, xC, maxValidX, "C: x should be within bounds")
	assert.LessOrEqual(t, yC, maxValidY, "C: y should be within bounds")
	assert.LessOrEqual(t, xVk, maxValidX, "Vk: x should be within bounds")
	assert.LessOrEqual(t, yVk, maxValidY, "Vk: y should be within bounds")

	assert.Equal(t, xC, xVk, "x position should match")
	assert.Equal(t, yC, yVk, "y position should match")
	assert.InDelta(t, scoreC, scoreVk, 1e-5, "scores should be close")
}

// Test_SearchRGBA_VkVsC_EdgePattern tests with a pattern placed at the edge of the search area.
func Test_SearchRGBA_VkVsC_EdgePattern(t *testing.T) {
	imgW, imgH := 20, 20
	patW, patH := 5, 5

	// Place a distinctive pattern at the far edge of the search area
	// Search area is [0, imgW-patW] x [0, imgH-patH] = [0, 15] x [0, 15]
	edgeX := imgW - patW // 15 - the last valid x position
	edgeY := imgH - patH // 15 - the last valid y position

	// Create image with gradient
	img := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
	for y := 0; y < imgH; y++ {
		for x := 0; x < imgW; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(x * 10),
				G: uint8(y * 10),
				B: 100,
				A: 255,
			})
		}
	}

	// Place a bright white block at the edge (this will be our target)
	for y := edgeY; y < imgH; y++ {
		for x := edgeX; x < imgW; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	// Pattern is the white block
	pat := image.NewRGBA(image.Rect(0, 0, patW, patH))
	for y := 0; y < patH; y++ {
		for x := 0; x < patW; x++ {
			pat.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	xC, yC, scoreC := SearchRGBAC(img, pat)
	xVk, yVk, scoreVk := SearchRGBAVk(img, pat)

	t.Logf("Edge pattern - C: (%d, %d) score=%.10f, Vk: (%d, %d) score=%.10f",
		xC, yC, scoreC, xVk, yVk, scoreVk)
	t.Logf("Expected position: (%d, %d)", edgeX, edgeY)

	assert.Equal(t, edgeX, xC, "C: should find pattern at edge x")
	assert.Equal(t, edgeY, yC, "C: should find pattern at edge y")
	assert.Equal(t, edgeX, xVk, "Vk: should find pattern at edge x")
	assert.Equal(t, edgeY, yVk, "Vk: should find pattern at edge y")
	assert.InDelta(t, 1.0, scoreC, 1e-5, "C: score should be 1.0 for exact match")
	assert.InDelta(t, 1.0, scoreVk, 1e-5, "Vk: score should be 1.0 for exact match")
}

// Test_SearchRGBA_VkVsC_MultipleRuns tests consistency across multiple runs.
func Test_SearchRGBA_VkVsC_MultipleRuns(t *testing.T) {
	img := imutil.ToRGBA(LoadTestImg())

	rng := rand.New(rand.NewSource(999))
	patW, patH := 30, 25
	pat := image.NewRGBA(image.Rect(0, 0, patW, patH))
	for y := 0; y < patH; y++ {
		for x := 0; x < patW; x++ {
			pat.SetRGBA(x, y, color.RGBA{
				R: uint8(rng.Intn(256)),
				G: uint8(rng.Intn(256)),
				B: uint8(rng.Intn(256)),
				A: 255,
			})
		}
	}

	// Run multiple times to check for race conditions
	for i := 0; i < 10; i++ {
		xC, yC, scoreC := SearchRGBAC(img, pat)
		xVk, yVk, scoreVk := SearchRGBAVk(img, pat)

		assert.Equal(t, xC, xVk, "run %d: x position should match", i)
		assert.Equal(t, yC, yVk, "run %d: y position should match", i)
		// Float32 (Vulkan) vs float64 (C) precision difference allows ~1e-4 delta
		assert.InDelta(t, scoreC, scoreVk, 1e-4, "run %d: scores should be close", i)
	}
}
