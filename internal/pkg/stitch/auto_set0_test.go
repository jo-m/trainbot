package stitch

import (
	"image"
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

const anyLengthMagic = 0xDEADBEEF

// runTestSimple runs autostitching on a video file and checks the resulting train length.
// lengthM == 0 means that no train is expected to be detected.
func runTestSimple(t *testing.T, c Config, r image.Rectangle, video string, lengthM float64) []Train {
	// logging.MustInit(logging.LogConfig{LogLevel: "debug", LogPretty: true})
	log.Logger = zerolog.Nop()

	src, err := vid.NewFileSrc(video, false)
	require.NoError(t, err)
	defer src.Close()

	auto := NewAutoStitcher(c)

	var trains []Train
	// defer func() {
	// 	if len(trains) == 0 {
	// 		f, err := os.Create(fmt.Sprintf("%s.MISSING", filepath.Base(video)))
	// 		if err == nil {
	// 			f.Close()
	// 		}
	// 	}

	// 	for i, tr := range trains {
	// 		fname := fmt.Sprintf("%s-%02d.jpg", filepath.Base(video), i)
	// 		imutil.Dump(fname, tr.Image)
	// 	}
	// }()
	for {
		frame, ts, err := src.GetFrame()
		if err == io.EOF {
			log.Info().Msg("no more frames")
			break
		}
		require.NoError(t, err)

		frame, err = imutil.Sub(frame, r)
		require.NoError(t, err)
		tr := auto.Frame(imutil.Copy(frame), *ts)
		if tr != nil {
			trains = append(trains, *tr)
			log.Info().Msg("got train")
		}
	}

	if tr := auto.TryStitchAndReset(); tr != nil {
		trains = append(trains, *tr)
	}

	if len(trains) == 0 {
		if lengthM == 0 {
			// None expected, none found.
			return trains
		}

		assert.True(t, false, "train expected but none found: %s", video)
		return trains
	}

	if len(trains) > 1 {
		c := 0
		if lengthM != 0 {
			c = 1
		}
		assert.False(t, true, "expected %d train(s) but %d detected: %s", c, len(trains), video)
		return trains
	}

	// len(trains) == 1
	if lengthM == 0 {
		assert.False(t, true, "no train expected but one was found: %s", video)
		return trains
	}

	if lengthM != anyLengthMagic {
		assert.InDelta(t, lengthM, trains[0].LengthM(), 5, "length does not match: %s", video)
	}

	return trains
}

// runTestDetailed also checks for speed, acceleration, and direction, and compares the stitched image.
// Always expects one train to be detected.
func runTestDetailed(t *testing.T, c Config, r image.Rectangle, video string, truthImg string, lengthM, speed, accel float64, direction bool) {
	trains := runTestSimple(t, c, r, video, lengthM)
	assert.Len(t, trains, 1, "expected exactly one train")

	if len(trains) != 1 {
		return
	}

	train := trains[0]

	// Speed/accel estimation.
	assert.InDelta(t, speed, train.SpeedMpS(), 0.1, "speed does not match: %s", video)
	assert.InDelta(t, accel, train.AccelMpS2(), 0.1, "acceleration does not match: %s", video)
	assert.True(t, train.Direction() == direction, "direction does not match: %s", video)

	// Check stitched image.
	truth, err := imutil.Load(truthImg)
	require.NoError(t, err)
	testutil.AssertImagesAlmostEqual(t, truth, train.Image)
}

func Test_AutoStitcher_Set0(t *testing.T) {
	c := Config{
		PixelsPerM:  50,
		MinSpeedKPH: 10,
		MaxSpeedKPH: 160,
		MinLengthM:  10,
	}
	r := image.Rect(0, 0, 300, 300)

	runTestDetailed(t, c, r, "testdata/set0/day.mp4", "testdata/set0/day.jpg", 86, 21.53, 0.27, false)
	runTestDetailed(t, c, r, "testdata/set0/night.mp4", "testdata/set0/night.jpg", 83, 22.7, -0.5, true)
	runTestDetailed(t, c, r, "testdata/set0/rain.mp4", "testdata/set0/rain.jpg", 82, 17.9, 0, true)
	runTestDetailed(t, c, r, "testdata/set0/snow.mp4", "testdata/set0/snow.jpg", 56, 20.5, -0.75, true)
}
