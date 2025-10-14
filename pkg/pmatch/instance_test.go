package pmatch

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"jo-m.ch/go/trainbot/pkg/imutil"
)

func Test_NewInstance(t *testing.T) {
	img := imutil.ToRGBA(LoadTestImg())
	pat, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	require.NoError(t, err)

	inst := NewInstance()
	defer inst.Destroy()

	x, y, score := inst.SearchRGBA(img, pat.(*image.RGBA))
	assert.InDelta(t, 1., score, delta)
	assert.Equal(t, x0, x)
	assert.Equal(t, y0, y)

	// Also resets pat bounds origin to (0,0).
	patCopy := imutil.ToRGBA(pat.(*image.RGBA))

	x, y, score = inst.SearchRGBA(img, patCopy)
	assert.InDelta(t, 1., score, delta)
	assert.Equal(t, x0, x)
	assert.Equal(t, y0, y)

	t.Log("Instance Kind:", inst.Kind())
}
