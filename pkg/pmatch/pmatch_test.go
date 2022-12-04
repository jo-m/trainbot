package pmatch

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"testing"

	"github.com/jo-m/trainbot/internal/pkg/imutil"
	"github.com/jo-m/trainbot/internal/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testImgPath = "testdata/bird.jpg"

const (
	x0, y0, w, h = 65, 35, 30, 20
)

type scoreGrayFn func(img, pat *image.Gray, offset image.Point) float64

func testScore(t *testing.T, img, pat *image.Gray, perfectScore float64, scoreFn scoreGrayFn) {
	// score at patch origin
	offset0 := image.Pt(x0, y0)
	score0 := scoreFn(img, pat, offset0)
	assert.Equal(t, perfectScore, score0)

	// score at offset
	offset1 := image.Pt(x0+1, y0+0)
	score1 := scoreFn(img, pat, offset1)
	assert.Less(t, score1, score0)

	offset2 := image.Pt(x0+0, y0+10)
	score2 := scoreFn(img, pat, offset2)
	assert.Less(t, score2, score0)

	offset3 := image.Pt(x0+1, y0+1)
	score3 := scoreFn(img, pat, offset3)
	assert.Less(t, score3, score0)

	// score at larger offset
	offset4 := image.Pt(x0+3, y0+3)
	score4 := scoreFn(img, pat, offset4)
	assert.Less(t, score4, score3)
}

func Test_ScoreGrayCos_Simple(t *testing.T) {
	img := imutil.ToGray(testutil.LoadImg(t, testImgPath))
	pat, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	require.NoError(t, err)

	ScoreGrayCos(img, pat.(*image.Gray), image.Pt(0, 0))

	testScore(t, img, pat.(*image.Gray), 1, ScoreGrayCos)

	// reset pat bounds origin to (0,0)
	pat = imutil.GrayReset0(pat.(*image.Gray))
	testScore(t, img, pat.(*image.Gray), 1, ScoreGrayCos)
}

func Test_ScoreGrayCos_Panics(t *testing.T) {
	img := imutil.ToGray(testutil.LoadImg(t, testImgPath))
	pat, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	require.NoError(t, err)

	assert.Panics(t, func() {
		ScoreGrayCos(img, pat.(*image.Gray), image.Pt(0, -1))
	})
	assert.Panics(t, func() {
		ScoreGrayCos(img, pat.(*image.Gray), image.Pt(-1, 0))
	})
	assert.Panics(t, func() {
		ScoreGrayCos(img, pat.(*image.Gray), image.Pt(-1, -1))
	})

	assert.Panics(t, func() {
		ScoreGrayCos(img, pat.(*image.Gray), image.Pt(0, 200))
	})
	assert.Panics(t, func() {
		ScoreGrayCos(img, pat.(*image.Gray), image.Pt(200, 0))
	})
	assert.Panics(t, func() {
		ScoreGrayCos(img, pat.(*image.Gray), image.Pt(200, 200))
	})
}

func Benchmark_ScoreGrayCos(b *testing.B) {
	imgRGB, err := imutil.Load(testImgPath)
	if err != nil {
		b.Fail()
	}
	img := imutil.ToGray(imgRGB)
	pat, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	if err != nil {
		b.Fail()
	}

	for i := 0; i < b.N; i++ {
		ScoreGrayCos(img, pat.(*image.Gray), image.Pt(x0, y0))
	}
}

func Test_Search_ScoreGrayCos(t *testing.T) {
	img := imutil.ToGray(testutil.LoadImg(t, testImgPath))
	pat, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	require.NoError(t, err)

	x, y, score := Search(img, pat.(*image.Gray), ScoreGrayCos)
	assert.Equal(t, 1., score)
	assert.Equal(t, x0, x)
	assert.Equal(t, y0, y)

	// reset pat bounds origin to (0,0)
	pat = imutil.GrayReset0(pat.(*image.Gray))

	x, y, score = Search(img, pat.(*image.Gray), ScoreGrayCos)
	assert.Equal(t, 1., score)
	assert.Equal(t, x0, x)
	assert.Equal(t, y0, y)
}

func Benchmark_Search_ScoreGrayCos(b *testing.B) {
	imgRGB, err := imutil.Load(testImgPath)
	if err != nil {
		b.Fail()
	}
	img := imutil.ToGray(imgRGB)
	pat, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	if err != nil {
		b.Fail()
	}

	for i := 0; i < b.N; i++ {
		Search(img, pat.(*image.Gray), ScoreGrayCos)
	}
}
