package stitch

import (
	"io"
	"testing"

	"github.com/jo-m/trainbot/internal/pkg/testutil"
	"github.com/jo-m/trainbot/pkg/imutil"
	"github.com/jo-m/trainbot/pkg/vid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runTestDetailed(t *testing.T, video string, truthImg string, speed, accel float64, direction bool) {
	// logging.MustInit(logging.LogConfig{LogLevel: "debug", LogPretty: true})
	log.Logger = zerolog.Nop()

	var src vid.Src
	src, err := vid.NewFileSrc(video, false)
	require.NoError(t, err)
	defer src.Close()
	src = vid.NewSrcBuf(src, 0)

	c := Config{
		PixelsPerM:  50,
		MinSpeedKPH: 10,
		MaxSpeedKPH: 160,
		MinLengthM:  10,
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

	if t := auto.TryStitchAndReset(); t != nil {
		train = t
	}

	require.NotNil(t, train)

	// Speed/accel estimation.
	assert.InDelta(t, speed, train.SpeedMpS(), 0.1, video+": speed does not match")
	assert.InDelta(t, accel, train.AccelMpS2(), 0.1, video+": acceleration does not match")
	assert.True(t, train.Direction() == direction, video+": direction does not match")

	// imutil.Dump("truth-"+path.Base(truthImg), train.Image)

	// Check stitched image.
	truth, err := imutil.Load(truthImg)
	require.NoError(t, err)
	testutil.AssertImagesAlmostEqual(t, truth, train.Image)
}

func Test_AutoStitcher_Set0(t *testing.T) {
	runTestDetailed(t, "testdata/set0/day.mp4", "testdata/set0/day.jpg", 21.53, 0.27, false)
	runTestDetailed(t, "testdata/set0/night.mp4", "testdata/set0/night.jpg", 22.7, -0.5, true)
	runTestDetailed(t, "testdata/set0/rain.mp4", "testdata/set0/rain.jpg", 17.9, 0, true)
	runTestDetailed(t, "testdata/set0/snow.mp4", "testdata/set0/snow.jpg", 20.5, -0.75, true)
}
