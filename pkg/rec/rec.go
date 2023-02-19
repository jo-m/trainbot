package rec

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path"
	"time"

	"github.com/jo-m/trainbot/internal/pkg/imutil"
	"github.com/jo-m/trainbot/pkg/pmatch"
	"github.com/rs/zerolog/log"
)

const (
	scoreThreshold = 0.99
	metaFileName   = "meta.json"
)

type frameMeta struct {
	Number   int       `json:"nr"`
	TimeUTC  time.Time `json:"tsUTC"`
	FileName string    `json:"fileName"`
}

type AutoRec struct {
	basePath string

	prevCount int
	prevTs    time.Time
	prevFrame *image.RGBA
	avgScore  float64

	currentMeta []frameMeta
	currentPath string
}

func NewAutoRec(basePath string) *AutoRec {
	return &AutoRec{
		basePath: basePath,
	}
}

func (r *AutoRec) initialize(ts time.Time) error {
	r.currentPath = path.Join(r.basePath, ts.Format("20060102_150405.999_Z07:00"))

	log.Info().Str("path", r.currentPath).Time("ts", ts).Msg("initializing recording")

	return os.MkdirAll(r.currentPath, 0755)
}

func (r *AutoRec) finalize(ts time.Time) error {
	log.Info().Str("path", r.currentPath).Int("nFrames", len(r.currentMeta)).Msg("finalizing recording")

	f, err := os.Create(path.Join(r.currentPath, metaFileName))
	if err != nil {
		log.Err(err).Send()
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(r.currentMeta)
	if err != nil {
		log.Err(err).Send()
		return err
	}

	r.currentMeta = nil
	r.currentPath = ""

	return nil
}

func (r *AutoRec) record(prevFrame image.Image, prevTs time.Time) error {
	meta := frameMeta{
		Number:   r.prevCount,
		TimeUTC:  prevTs,
		FileName: fmt.Sprintf("frame_%06d.qoi", r.prevCount),
	}
	r.currentMeta = append(r.currentMeta, meta)

	log.Debug().Str("fileName", meta.FileName).Msg("dumping frame")
	err := imutil.Dump(path.Join(r.currentPath, meta.FileName), prevFrame)
	if err != nil {
		log.Err(err).Send()
		return err
	}
	return nil
}

// will make a copy of the image
func (r *AutoRec) Frame(frame image.Image, ts time.Time) error {
	// create copy
	frameCopy := imutil.ToRGBA(frame)
	defer func() {
		r.prevTs = ts
		r.prevFrame = frameCopy
		r.prevCount++
	}()

	if r.prevFrame == nil {
		// first time
		return nil
	}

	// match similarity of past and current frame
	score := pmatch.CosSimRGBA(frameCopy, r.prevFrame)
	if r.avgScore == 0 || score < scoreThreshold {
		// initialize, and/or make sure that we don't miss it if something changes
		r.avgScore = score
		log.Debug().Msg("initialize score")
	} else {
		// just update
		r.avgScore = r.avgScore*0.9 + score*0.1
	}
	log.Debug().Float64("score", score).Float64("avgScore", r.avgScore).Send()

	shouldRecord := r.avgScore < scoreThreshold
	isRecording := len(r.currentMeta) > 0
	if shouldRecord {
		if !isRecording {
			err := r.initialize(ts)
			if err != nil {
				return err
			}
		}

		err := r.record(r.prevFrame, r.prevTs)
		if err != nil {
			return err
		}
	} else {
		if isRecording {
			err := r.finalize(ts)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
