package stitch

import (
	"image"
	"io"
	"testing"

	"github.com/jo-m/trainbot/internal/pkg/imutil"
	"github.com/jo-m/trainbot/pkg/vid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertImagesAlmostEqual(t *testing.T, truth image.Image, test image.Image) {
	require.Equal(t, truth.Bounds().Size(), test.Bounds().Size())

	var diff uint64
	for y := 0; y < truth.Bounds().Dy(); y++ {
		for x := 0; x < truth.Bounds().Dx(); x++ {
			r0, g0, b0, _ := truth.At(x, y).RGBA()
			r1, g1, b1, _ := test.At(x, y).RGBA()

			diff += uint64(iabs(int(r0)-int(r1)) / 255)
			diff += uint64(iabs(int(g0)-int(g1)) / 255)
			diff += uint64(iabs(int(b0)-int(b1)) / 255)
		}
	}

	diffPerPx := float64(diff) / float64(truth.Bounds().Dx()) / float64(truth.Bounds().Dy()) / 3
	assert.Less(t, diffPerPx, 1.)
}

func runTest(t *testing.T, video string, truthImg string, speed, accel float64, direction bool) {
	// logging.MustInit(logging.LogConfig{LogLevel: "debug", LogPretty: true})
	log.Logger = zerolog.Nop()

	src, err := vid.NewFileSrc(video, false)
	require.NoError(t, err)
	defer src.Close()

	c := Config{
		PixelsPerM:  50,
		MinSpeedKPH: 10,
		MaxSpeedKPH: 160,
	}
	auto := NewAutoStitcher(c)

	var train *Train
	for {
		frame, ts, err := src.GetFrame()
		if err == io.EOF {
			log.Info().Msg("no more frames")
			break
		}
		require.NoError(t, err)

		t := auto.Frame(frame, *ts)
		if t != nil {
			train = t
			log.Info().Msg("got train")
		}
	}

	require.NotNil(t, train)

	// Speed/accel estimation.
	assert.InDelta(t, speed, train.SpeedMpS(), 0.1)
	assert.InDelta(t, accel, train.AccelMpS2(), 0.1)
	assert.True(t, train.Direction() == direction)

	// Check stitched image.
	truth, err := imutil.Load(truthImg)
	require.NoError(t, err)
	assertImagesAlmostEqual(t, truth, train.Image)
}

func Test_AutoStitcher_1(t *testing.T) {
	runTest(t, "testdata/test1.mp4", "testdata/test1.jpg", 21.53, -0.6, false)
	runTest(t, "testdata/test2.mp4", "testdata/test2.jpg", 22.6, -0.5, true)
}
