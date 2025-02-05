// Package main (trainbot) tries to automatically stitch images of passing trains.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"io/fs"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/jmoiron/sqlx"
	"github.com/jo-m/trainbot/internal/pkg/db"
	"github.com/jo-m/trainbot/internal/pkg/logging"
	"github.com/jo-m/trainbot/internal/pkg/prometheus"
	"github.com/jo-m/trainbot/internal/pkg/stitch"
	"github.com/jo-m/trainbot/internal/pkg/upload"
	"github.com/jo-m/trainbot/pkg/imutil"
	"github.com/jo-m/trainbot/pkg/vid"
	"github.com/nfnt/resize"
	"github.com/rs/zerolog/log"
)

type config struct {
	logging.LogConfig

	InputFile          string `arg:"--input,env:INPUT" help:"Video4linux device file or regular video file, e.g. /dev/video0, video.mp4, or 'picam3'" placeholder:"FILE"`
	CameraFormatFourCC string `arg:"--camera-format-fourcc,env:CAMERA_FORMAT_FOURCC" default:"MJPG" help:"Camera pixel format FourCC string, ignored if using video file" placeholder:"CODE"`
	CameraW            int    `arg:"--camera-w,env:CAMERA_W" default:"1920" help:"Camera frame size width, ignored if using video file or picam3" placeholder:"X"`
	CameraH            int    `arg:"--camera-h,env:CAMERA_H" default:"1080" help:"Camera frame size height, ignored if using video file or picam3" placeholder:"Y"`

	RectX uint `arg:"-X,--rect-x,env:RECT_X" help:"Rect to look at, x (left)" placeholder:"N"`
	RectY uint `arg:"-Y,--rect-y,env:RECT_Y" help:"Rect to look at, y (top)" placeholder:"N"`
	RectW uint `arg:"-W,--rect-w,env:RECT_W" help:"Rect to look at, width" placeholder:"N"`
	RectH uint `arg:"-H,--rect-h,env:RECT_H" help:"Rect to look at, height" placeholder:"N"`

	Rotate180 bool `arg:"--rotate-180,env:ROTATE_180" help:"Rotate camera picture 180 degrees (only picam3)"`

	PixelsPerM          float64 `arg:"--px-per-m,env:PX_PER_M" default:"45" help:"Pixels per meter, can be reconstructed from sleepers: they are usually 0.6m apart (in Europe)" placeholder:"K"`
	MinSpeedKPH         float64 `arg:"--min-speed-kph,env:MIN_SPEED_KPH" default:"25" help:"Assumed train min speed, km/h" placeholder:"K"`
	MaxSpeedKPH         float64 `arg:"--max-speed-kph,env:MAX_SPEED_KPH" default:"160" help:"Assumed train max speed, km/h" placeholder:"K"`
	MinLengthM          float64 `arg:"--min-len-m,env:MIN_LEN_M" default:"5" help:"Minimum length of trains" placeholder:"K"`
	MaxFrameCountPerSeq int     `arg:"--max-frame-count-per-seq,env:MAX_FRAME_COUNT_PER_SEQ" default:"1500" help:"How many frames to accept max. before force-ending a train sequence. If you have high fps videos/long trains, you can increase it from the default, but the program will use more memory." placeholder:"N"`

	CPUProfile  bool `arg:"--cpu-profile,env:CPU_PROFILE" help:"Write CPU profile"`
	HeapProfile bool `arg:"--heap-profile,env:HEAP_PROFILE" help:"Write memory heap profiles"`

	EnableUpload bool `arg:"--enable-upload,env:ENABLE_UPLOAD" help:"Enable uploading of data."`

	upload.FTPConfig
	upload.DataStore

	Prometheus       bool   `arg:"--prometheus,env:PROMETHEUS" default:"false" help:"Expose Prometheus-compatible metrics endpoint."`
	PrometheusListen string `arg:"--prometheus-listen,env:PROMETHEUS_LISTEN" default:":18963" help:"Which host and port to bind prometheus endpoint to."`
}

func (c *config) getRect() image.Rectangle {
	return image.Rect(0, 0, int(c.RectW), int(c.RectH)).Add(image.Pt(int(c.RectX), int(c.RectY)))
}

func (c *config) mustOpenDB() *sqlx.DB {
	dbx, err := db.Open(c.GetDBPath())
	if err != nil {
		log.Panic().Err(err).Msg("could not create/open database")
	}

	return dbx
}

const (
	rectSizeMin = 100
	rectSizeMax = 500

	failedFramesMax = 50

	inputFilePiCam3 = "picam3"

	profCPUFile  = "prof-cpu.gz"
	profHeapFile = "prof-heap-%05d.gz"
)

func parseCheckArgs() config {
	c := config{}
	p := arg.MustParse(&c)
	logging.MustInit(c.LogConfig)
	if c.Prometheus {
		prometheus.Init(c.PrometheusListen)
	}

	if c.InputFile == "" {
		p.Fail("no camera device or video file passed")
	}

	r := c.getRect()
	if r.Size().X == 0 && r.Size().Y == 0 {
		p.Fail("no rect set (use --rect-.. parameters to set crop region)")
	}
	if r.Size().X < rectSizeMin || r.Size().Y < rectSizeMin {
		p.Fail(fmt.Sprintf("rect is too small (minimum width and height is %d px)", rectSizeMin))
	}
	if r.Size().X > rectSizeMax || r.Size().Y > rectSizeMax {
		p.Fail(fmt.Sprintf("rect is too large (maximum width and height is %d px)", rectSizeMax))
	}

	return c
}

func openSrc(c config) (vid.Src, error) {
	// Pi cam.
	if c.InputFile == inputFilePiCam3 {
		return vid.NewPiCam3Src(vid.PiCam3Config{
			Rect:      c.getRect(),
			Focus:     0,
			Rotate180: c.Rotate180,
			Format:    vid.FourCCFromString(c.CameraFormatFourCC),
			FPS:       30,
		})
	}

	stat, err := os.Stat(c.InputFile)
	if err != nil {
		return nil, err
	}

	if stat.Mode().IsRegular() {
		// Video file.
		return vid.NewFileSrc(c.InputFile, false)
	}

	return vid.NewCamSrc(vid.CamConfig{
		DeviceFile: c.InputFile,
		Format:     vid.FourCCFromString(c.CameraFormatFourCC),
		FrameSize:  image.Point{c.CameraW, c.CameraH},
	})
}

func detectTrainsForever(c config, trainsOut chan<- *stitch.Train) {
	rect := c.getRect()

	src, err := openSrc(c)
	if err != nil {
		log.Panic().Err(err).Str("path", c.InputFile).Msg("failed to open video source")
	}
	defer src.Close()
	srcBuf := vid.NewSrcBuf(src, failedFramesMax)

	stitcher := stitch.NewAutoStitcher(stitch.Config{
		PixelsPerM:          c.PixelsPerM,
		MinSpeedKPH:         c.MinSpeedKPH,
		MaxSpeedKPH:         c.MaxSpeedKPH,
		MinLengthM:          c.MinLengthM,
		MaxFrameCountPerSeq: c.MaxFrameCountPerSeq,
	})
	defer func() {
		train := stitcher.TryStitchAndReset()
		if train != nil {
			trainsOut <- train
		}
	}()

	for i := uint64(0); ; i++ {
		frame, ts, err := srcBuf.GetFrame()
		if err != nil {
			log.Err(err).Msg("no more frames")
			break
		}

		var cropped image.Image
		if c.InputFile == inputFilePiCam3 {
			// PiCam output is already cropped.
			cropped = frame
		} else {
			cropped, err = imutil.Sub(frame, rect)
			if err != nil {
				log.Panic().Err(err).Msg("failed to crop frame")
			}

			// Create a new image with only the cropped pixels,
			// so we can gc the potentially large area from the
			// original image.
			cropped = imutil.Copy(cropped)
		}

		if cropped.Bounds().Size() != rect.Size() {
			log.Panic().Interface("cam", cropped.Bounds().Size()).Interface("conf", rect.Size()).Msg("rect size mismatch")
		}

		train := stitcher.Frame(cropped, *ts)
		if train != nil {
			trainsOut <- train
		}

		if c.HeapProfile && i%1000 == 0 {
			fname := fmt.Sprintf(profHeapFile, i)
			// #nosec 304
			f, err := os.Create(fname)
			if err != nil {
				log.Err(err).Str("file", fname).Msg("failed to open heap profile file")
				continue
			}

			log.Info().Str("file", fname).Msg("writing heap profile")
			err = pprof.WriteHeapProfile(f)
			if err != nil {
				log.Err(err).Str("file", fname).Msg("failed to write heap profile")
			}

			err = f.Close()
			if err != nil {
				log.Err(err).Str("file", fname).Msg("failed to close heap profile file")
			}
		}
	}
}

func processTrains(store upload.DataStore, dbx *sqlx.DB, trainsIn <-chan *stitch.Train, wg *sync.WaitGroup) {
	defer wg.Done()

	for train := range trainsIn {
		log.Info().
			Time("ts", train.StartTS).
			Float64("speedMpS", train.SpeedMpS()).
			Float64("speedKmh", train.SpeedMpS()*3.6).
			Float64("accelMpS2", train.AccelMpS2()).
			Str("direction", train.DirectionS()).
			Msg("found train")

		// Dump stitched image.
		dbTrain := db.Train{StartTS: train.StartTS}
		err := imutil.Dump(store.GetBlobPath(dbTrain.ImgFileName()), train.Image)
		if err != nil {
			log.Err(err).Send()
			continue
		}
		log.Debug().Str("imgFileName", dbTrain.ImgFileName()).Msg("wrote JPEG")

		// Dump thumbnail.
		thumb := resize.Thumbnail(math.MaxUint, 64, train.Image, resize.Bilinear)
		err = imutil.DumpJPEG(store.GetBlobThumbPath(dbTrain.ImgFileName()), thumb, 75)
		if err != nil {
			log.Err(err).Send()
			continue
		}

		// Dump GIF.
		err = imutil.DumpGIF(store.GetBlobPath(dbTrain.GIFFileName()), train.GIF)
		if err != nil {
			log.Err(err).Send()
			continue
		}
		log.Debug().Str("gifFileName", dbTrain.GIFFileName()).Msg("wrote GIF")

		id, err := db.InsertTrain(dbx, *train)
		if err != nil {
			log.Err(err).Send()
		}
		log.Info().Int64("id", id).Msg("added train to db")
	}
}

func uploadOnce(store upload.DataStore, dbx *sqlx.DB, c upload.FTPConfig) {
	ctx := context.Background()
	uploader, err := upload.NewFTP(ctx, c)
	if err != nil {
		log.Err(err).Msg("could not create uploader")
		return
	}
	defer uploader.Close()

	n, err := upload.All(ctx, store, dbx, uploader)
	if err != nil {
		log.Err(err).Msg("uploading all failed")
		return
	}

	log.Debug().Int("n", n).Msg("uploaded files")
}

func uploadForever(store upload.DataStore, dbx *sqlx.DB, c upload.FTPConfig) {
	for {
		uploadOnce(store, dbx, c)
		time.Sleep(time.Second * 5)
	}
}

func cleanupOrphanedRemoteBlobsOnce(dbx *sqlx.DB, c upload.FTPConfig) {
	ctx := context.Background()
	uploader, err := upload.NewFTP(ctx, c)
	if err != nil {
		log.Err(err).Msg("could not create uploader")
		return
	}
	defer uploader.Close()

	n, err := upload.CleanupOrphanedRemoteBlobs(ctx, dbx, uploader)
	if err != nil {
		log.Err(err).Msg("cleaning up orphaned remote blobs failed")
		return
	}

	log.Info().Int("n", n).Msg("cleaned up orphaned remote blobs")
}

func cleanupOrphanedRemoteBlobsForever(dbx *sqlx.DB, c upload.FTPConfig) {
	for {
		cleanupOrphanedRemoteBlobsOnce(dbx, c)

		// Sleep for a long time because we don't want to annoy the server sysadmins with FTP LIST commands
		// returning large listings all the time.
		// #nosec G404
		sleepS := 3600 + rand.Intn(3600)
		log.Info().Int("sleepS", sleepS).Msg("sleeping until next cleanup")
		time.Sleep(time.Duration(sleepS) * time.Second)
	}
}

func deleteOldLocalBlobsOnce(store upload.DataStore, dbx *sqlx.DB) error {
	for {
		toCleanup, err := db.GetNextCleanup(dbx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Debug().Msg("no more files to clean up")
				return nil
			}

			log.Err(err).Send()
			return err
		}

		err = os.Remove(store.GetBlobPath(toCleanup.ImgFileName()))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				log.Debug().Str("path", toCleanup.ImgFileName()).Msg("tried removing but file does not exist")
			} else {
				log.Err(err).Send()
				return err
			}
		}

		err = os.Remove(store.GetBlobThumbPath(toCleanup.ImgFileName()))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				log.Debug().Str("path", toCleanup.ImgFileName()).Msg("tried removing but file does not exist")
			} else {
				log.Err(err).Send()
				return err
			}
		}

		err = os.Remove(store.GetBlobPath(toCleanup.GIFFileName()))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				log.Debug().Str("path", toCleanup.GIFFileName()).Msg("tried removing but file does not exist")
			} else {
				log.Err(err).Send()
				return err
			}
		}

		err = db.SetCleanedUp(dbx, toCleanup.ID)
		if err != nil {
			log.Err(err).Send()
			return err
		}
	}
}

func deleteOldLocalBlobsForever(store upload.DataStore, dbx *sqlx.DB) {
	for {
		err := deleteOldLocalBlobsOnce(store, dbx)
		if err != nil {
			log.Err(err).Msg("failed up clean up")
		}
		time.Sleep(time.Second * 5)
	}
}

func main() {
	c := parseCheckArgs()

	// Try to create output directory.
	err := os.MkdirAll(c.DataDir, 0750)
	if err != nil {
		log.Panic().Err(err).Msg("could not create output directory")
	}

	log.Info().Interface("config", c).Msg("starting")

	if c.CPUProfile {
		log.Info().Str("file", profCPUFile).Msg("writing CPU profile")
		f, err := os.Create(profCPUFile)
		if err != nil {
			log.Panic().Err(err).Msg("failed to create CPU profile file")
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			log.Panic().Err(err).Msg("failed to start CPU profile")
		}
		defer pprof.StopCPUProfile()
	}

	blobsDir := filepath.Join(c.DataDir, "blobs")
	err = os.MkdirAll(blobsDir, 0750)
	if err != nil {
		log.Panic().Err(err).Msg("could not create data and blobs directory")
	}

	dbx := c.mustOpenDB()
	defer dbx.Close()

	trains := make(chan *stitch.Train)
	done := sync.WaitGroup{}
	done.Add(1)
	go processTrains(c.DataStore, c.mustOpenDB(), trains, &done)
	if c.EnableUpload {
		go uploadForever(c.DataStore, c.mustOpenDB(), c.FTPConfig)
		go deleteOldLocalBlobsForever(c.DataStore, c.mustOpenDB())
		go cleanupOrphanedRemoteBlobsForever(c.mustOpenDB(), c.FTPConfig)
	}

	detectTrainsForever(c, trains)

	close(trains)
	done.Wait()
}
