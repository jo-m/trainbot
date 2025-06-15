package stitch

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"math"
	"time"

	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
	"github.com/go-gst/go-gst/gst/video"
	"github.com/mccutchen/palettor"
	"github.com/nfnt/resize"
	"github.com/rs/zerolog/log"
	"jo-m.ch/go/trainbot/internal/pkg/prometheus"
)

const (
	maxMemoryMB = 1024 * 1024 * 50
)

func isign(x int) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}

func sign(x float64) float64 {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}

func stitch(frames []image.Image, dx []int) (*image.RGBA, error) {
	t0 := time.Now()
	defer func() {
		log.Trace().Dur("dur", time.Since(t0)).Msg("stitch() duration")
	}()

	log.Info().Ints("dx", dx).Int("len(frames)", len(frames)).Msg("stitch()")

	// Sanity checks.
	if len(dx) < 2 {
		return nil, errors.New("sequence too short to stitch")
	}
	if len(frames) != len(dx) {
		log.Panic().Msg("frames and dx do not have the same length, this should not happen")
	}
	fb := frames[0].Bounds()
	for _, f := range frames {
		if f.Bounds() != fb {
			log.Panic().Msg("frame bounds or size not consistent, this should not happen")
		}
	}

	// Calculate base width.
	sign := isign(dx[0])
	w := fb.Dx() * sign
	h := fb.Dy()
	for _, x := range dx[1:] {
		if isign(x) != sign {
			return nil, errors.New("dx elements do not have consistent sign")
		}
		w += x
	}

	// Memory alloc sanity check.
	rect := image.Rect(0, 0, iabs(w), h)
	if rect.Size().X*rect.Size().Y*4 > maxMemoryMB {
		return nil, fmt.Errorf("would allocate too much memory: size %dx%d", rect.Size().X, rect.Size().Y)
	}
	img := image.NewRGBA(rect)

	// Forward?
	if w > 0 {
		pos := 0
		for i, f := range frames {
			draw.Draw(img, img.Bounds().Add(image.Pt(pos, 0)), f, f.Bounds().Min, draw.Src)
			pos += dx[i]
		}
	} else {
		// Backwards.
		pos := -w - fb.Dx()
		for i, f := range frames {
			draw.Draw(img, img.Bounds().Add(image.Pt(pos, 0)), f, f.Bounds().Min, draw.Src)
			pos += dx[i]
		}
	}

	return img, nil
}

// Train represents a detected train.
type Train struct {
	StartTS time.Time

	// Always positive.
	NFrames int

	// Always positive (absolute value).
	LengthPx float64
	// Positive sign means movement to the right, negative to the left.
	SpeedPxS float64
	// Positive sign means increasing speed for trains going to the right, breaking for trains going to the left.
	AccelPxS2 float64

	Conf Config

	Image *image.RGBA `json:"-"`
	GIF   *gif.GIF    `json:"-"`
}

// LengthM returns the absolute length in m.
func (t *Train) LengthM() float64 {
	return math.Abs(t.LengthPx) / t.Conf.PixelsPerM
}

// SpeedMpS returns the absolute speed in m/s.
func (t *Train) SpeedMpS() float64 {
	return math.Abs(t.SpeedPxS) / t.Conf.PixelsPerM
}

// AccelMpS2 returns the acceleration in m/2^2, corrected for speed direction:
// Positive means accelerating, negative means breaking.
func (t *Train) AccelMpS2() float64 {
	return t.AccelPxS2 / t.Conf.PixelsPerM * sign(t.SpeedPxS)
}

// Direction returns the train direction. Right = true, left = false.
func (t *Train) Direction() bool {
	return t.SpeedPxS > 0
}

// DirectionS returns the train direction as string "left" or "right".
func (t *Train) DirectionS() string {
	if t.SpeedPxS > 0 {
		return "right"
	}

	return "left"
}

func createGIF(seq sequence, stitched image.Image) (*gif.GIF, error) {
	// Extract palette.
	thumb := resize.Thumbnail(300, 300, stitched, resize.Lanczos3)
	const (
		paletteSize = 20
		nIter       = 100
	)
	pal, err := palettor.Extract(paletteSize, nIter, thumb)
	if err != nil {
		return nil, err
	}

	g := gif.GIF{}

	prevTS := *seq.startTS
	rect := seq.frames[0].Bounds().Sub(seq.frames[0].Bounds().Min)
	for i, ts := range seq.ts {
		dt := ts.Sub(prevTS)

		// Skip every other frame.
		if i%2 == 1 {
			continue
		}

		paletted := image.NewPaletted(rect, pal.Colors())
		draw.Draw(paletted, paletted.Bounds(), seq.frames[i], seq.frames[i].Bounds().Min, draw.Src)

		g.Image = append(g.Image, paletted)
		g.Delay = append(g.Delay, int(dt.Seconds()*100))

		prevTS = ts
	}

	return &g, nil
}

func createH264(seq sequence, stitched image.Image) (*gif.GIF, error) {
	// https://github.com/go-gst/go-gst/blob/v1.4.0/examples/appsrc/main.go

	// SW: x264enc
	// HW on RPi: v4l2h264enc
	// HW on PC AMD: va264enc

	// appsrc ! x264enc ! mp4mux ! filesync location=/tmp/test.mp4

	gst.Init(nil)

	pipeline, err := gst.NewPipeline("")
	if err != nil {
		return nil, err
	}

	encoder := "x264enc"
	elems, err := gst.NewElementMany("appsrc", "videoconvert", encoder, "h264parse", "mp4mux" /*"autovideosink"*/, "filesink")
	//elems, err := gst.NewElementMany("appsrc", "videoconvert", "autovideosink")
	if err != nil {
		return nil, err
	}

	pipeline.AddMany(elems...)
	gst.ElementLinkMany(elems...)

	src := app.SrcFromElement(elems[0])
	elems[5].SetArg("location", "/tmp/test.mp4")
	//elems[4].SetArg("sync", "false")

	// Specify the format we want to provide as application into the pipeline
	// by creating a video info with the given format and creating caps from it for the appsrc element.
	videoInfo := video.NewInfo().
		WithFormat(video.FormatRGBA, 300, 300). /*uint(seq.frames[0].Bounds().Dx()), uint(seq.frames[0].Bounds().Dy()))*/
		WithFPS(gst.Fraction(2, 1))             // FIXME

	src.SetCaps(videoInfo.ToCaps())
	src.SetProperty("format", gst.FormatTime)

	// Initialize a frame counter
	var i int
	//palette := video.FormatRGB8P.Palette()

	// Since our appsrc element operates in pull mode (it asks us to provide data),
	// we add a handler for the need-data callback and provide new data from there.
	// In our case, we told gstreamer that we do 2 frames per second. While the
	// buffers of all elements of the pipeline are still empty, this will be called
	// a couple of times until all of them are filled. After this initial period,
	// this handler will be called (on average) twice per second.
	src.SetCallbacks(&app.SourceCallbacks{
		NeedDataFunc: func(self *app.Source, _ uint) {

			// If we've reached the end of the palette, end the stream.
			if i == len(seq.frames) {
				src.EndStream()
				return
			}

			log.Debug().Int("frame", i).Msg("Producing frame")

			// Create a buffer that can hold exactly one video RGBA frame.
			buffer := gst.NewBufferWithSize(videoInfo.Size())

			// For each frame we produce, we set the timestamp when it should be displayed
			// The autovideosink will use this information to display the frame at the right time.
			buffer.SetPresentationTimestamp(gst.ClockTime(time.Duration(i) * 33 * time.Millisecond)) // FIXME

			// Produce an image frame for this iteration.
			pixels := seq.frames[i].(*image.RGBA).Pix
			//pixels := produceImageFrame(palette[i])

			// At this point, buffer is only a reference to an existing memory region somewhere.
			// When we want to access its content, we have to map it while requesting the required
			// mode of access (read, read/write).
			// See: https://gstreamer.freedesktop.org/documentation/plugin-development/advanced/allocation.html
			//
			// There are convenience wrappers for building buffers directly from byte sequences as
			// well.
			buffer.Map(gst.MapWrite).WriteData(pixels)
			buffer.Unmap()

			//buffer := gst.NewBufferFromBytes(pixels)

			// Push the buffer onto the pipeline.
			self.PushBuffer(buffer)

			log.Debug().Msg("buffer pushed")

			i++
		},
	})

	mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)

	pipeline.Ref()
	defer pipeline.Unref()

	pipeline.SetState(gst.StatePlaying)

	// Retrieve the bus from the pipeline and add a watch function
	pipeline.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
		log.Warn().Str("msg", msg.String()).Msg("gstreamer message")
		switch msg.Type() {
		case gst.MessageEOS:
			//time.Sleep(5 * time.Second)
			//pipeline.SetState(gst.StateNull)
			//time.Sleep(5 * time.Second)
			//mainLoop.Quit()
			//time.Sleep(5 * time.Second)
			//return false
		}
		//if err := handleMessage(msg); err != nil {
		//	fmt.Println(err)
		//	loop.Quit()
		//	return false
		//}
		return true
	})

	mainLoop.Run()
	//mainLoop.RunError()

	return nil, errors.New("look at /tmp/test.mp4 :)")
}
func produceImageFrame(c color.Color) []uint8 {
	width := 300
	height := 300
	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}
	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, c)
		}
	}

	return img.Pix
}

// fitAndStitch tries to stitch an image from a sequence.
// Will first try to fit a constant acceleration speed model for smoothing.
// Might modify seq (drops leading frames with no movement).
func fitAndStitch(seq sequence, c Config) (*Train, error) {
	start := time.Now()
	defer func() {
		log.Trace().Dur("dur", time.Since(start)).Msg("fitAndStitch() duration")
	}()

	log.Info().Ints("dx", seq.dx).Int("len(frames)", len(seq.frames)).Msg("fitAndStitch()")

	// Sanity checks.
	if len(seq.frames) != len(seq.dx) || len(seq.frames) != len(seq.ts) {
		log.Panic().Msg("length of frames, dx, ts are not equal, this should not happen")
	}
	if seq.startTS == nil {
		log.Panic().Msg("startTS is nil, this should not happen")
	}
	if len(seq.dx) == 0 || seq.dx[0] == 0 {
		log.Panic().Int("len", len(seq.dx)).Msg("sequence is empty or first value is 0")
	}

	// Remove trailing zeros.
	for len(seq.dx) > 0 && seq.dx[len(seq.dx)-1] == 0 {
		seq.dx = seq.dx[:len(seq.dx)-1]
		seq.ts = seq.ts[:len(seq.ts)-1]
		seq.frames = seq.frames[:len(seq.frames)-1]
	}
	prometheus.RecordSequenceLength(len(seq.frames))

	dxFit, ds, v0, a, err := fitDx(seq, float64(c.maxPxPerFrame(1)))
	if err != nil {
		prometheus.RecordFitAndStitchResult("unable_to_fit")
		return nil, fmt.Errorf("was not able to fit the sequence: %w", err)
	}

	if math.Abs(ds) < c.minLengthPx() {
		prometheus.RecordFitAndStitchResult("too_short")
		return nil, fmt.Errorf("discarded because too short, %f < %f", ds, c.minLengthPx())
	}

	// Estimate speed at halftime.
	t0 := seq.ts[0]
	tMid := seq.ts[len(seq.ts)/2]
	speed := v0 + a*tMid.Sub(t0).Seconds()

	if math.Abs(speed) < c.minSpeedPxPS() {
		prometheus.RecordFitAndStitchResult("too_slow")
		return nil, fmt.Errorf("discarded because too slow, %f < %f", speed, c.minSpeedPxPS())
	}

	img, err := stitch(seq.frames, dxFit)
	if err != nil {
		prometheus.RecordFitAndStitchResult("unable_to_assemble_image")
		return nil, fmt.Errorf("unable to assemble image: %w", err)
	}

	//gif, err := createGIF(seq, img)
	gif, err := createH264(seq, img)
	if err != nil {
		panic(err)
	}

	prometheus.RecordFitAndStitchResult("success")
	return &Train{
		t0,
		len(seq.frames),
		ds,
		-speed, // Negate because when things move to the left we get positive dx values.
		-a,
		c,
		img,
		gif,
	}, nil
}
