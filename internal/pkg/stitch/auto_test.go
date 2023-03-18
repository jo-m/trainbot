package stitch

import (
	"io"
	"testing"

	"github.com/jo-m/trainbot/internal/pkg/imutil"
	"github.com/jo-m/trainbot/internal/pkg/testutil"
	"github.com/jo-m/trainbot/pkg/vid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	testutil.AssertImagesAlmostEqual(t, truth, train.Image)
}

func Test_AutoStitcher_1(t *testing.T) {
	runTest(t, "testdata/test1.mp4", "testdata/test1.jpg", 21.53, -0.6, false)
	runTest(t, "testdata/test2.mp4", "testdata/test2.jpg", 22.6, -0.5, true)
}