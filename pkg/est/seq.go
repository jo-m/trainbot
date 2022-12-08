package est

import (
	"errors"
	"fmt"
	"image"
	"image/draw"

	"github.com/jo-m/trainbot/internal/pkg/imutil"
	"github.com/rs/zerolog/log"
)

type sequence struct {
	// Those slices always must have the same length.
	// dx[i] is the assumed offset between frames[i] and frames[i+1].
	// scores[i] is the score of that assumed offset.
	// All frames must have the same size.
	dx     []int
	frames []image.Image
}

func (s sequence) reversed() sequence {
	ret := sequence{}
	for i, dx := range s.dx {
		ret.dx = append(ret.dx, -dx)
		ret.frames = append(ret.frames, s.frames[len(s.frames)-i-1])
	}
	return ret
}

// allowed to remove trailing values from dx, but not values from the beginning
func cleanupDx(dx []int) ([]int, error) {
	if len(dx) < 10 {
		return nil, errors.New("len(x) must be >= 10")
	}
	if dx[0] == 0 {
		return nil, errors.New("first dx value cannot be 0")
	}

	// remove trailing zeros
	for dx[len(dx)-1] == 0 {
		dx = dx[:len(dx)-1]
	}

	dxFit, err := fitDx(dx)
	if err != nil {
		return nil, err
	}

	return dxFit, nil
}

func isign(x int) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}

func assemble(seq sequence) (*image.RGBA, error) {
	fsz := seq.frames[0].Bounds().Size()
	for _, f := range seq.frames {
		if f.Bounds().Min.X != 0 ||
			f.Bounds().Min.Y != 0 ||
			f.Bounds().Size() != fsz {
			return nil, errors.New("frame bounds or size not consistent")
		}
	}

	// calculate base width
	sign := isign(seq.dx[0])
	w := fsz.X * sign
	h := fsz.Y
	for _, x := range seq.dx[1:] {
		if isign(x) != sign {
			return nil, errors.New("dx elements do not have consistent sign")
		}
		w += x
	}

	img := image.NewRGBA(image.Rect(0, 0, iabs(w), h))

	// forward
	if w > 0 {
		pos := 0
		for i, f := range seq.frames {
			draw.Draw(img, f.Bounds().Add(image.Pt(pos, 0)), f, f.Bounds().Min, draw.Src)
			pos += seq.dx[i]
		}
	} else {
		// backwards
		pos := -w - fsz.X
		fmt.Println(w, fsz.X)
		for i, f := range seq.frames {
			fmt.Println(pos, f.Bounds())
			draw.Draw(img, f.Bounds().Add(image.Pt(pos, 0)), f, f.Bounds().Min, draw.Src)
			pos += seq.dx[i]
		}
	}

	return img, nil
}

func processSequence(seq sequence) error {
	if len(seq.dx) != len(seq.frames) {
		log.Panic().Msg("length of frames and dx should be equal, this should not happen")
	}

	// remove trailing zeros
	for seq.dx[len(seq.dx)-1] == 0 {
		seq.dx = seq.dx[:len(seq.dx)-1]
		seq.frames = seq.frames[:len(seq.frames)-1]
	}

	// checks
	if len(seq.dx) < 10 {
		return errors.New("sequence too short")
	}
	if seq.dx[0] == 0 {
		return errors.New("seq.dx cannot start with a zero")
	}

	var err error
	seq.dx, err = fitDx(seq.dx)
	if err != nil {
		return errors.New("was not able to fit the sequence")
	}

	img, err := assemble(seq)
	if err != nil {
		return fmt.Errorf("unable to assemble image: %w", err)
	}
	imutil.Dump("imgs/_assembled.png", img) // TODO

	return nil
}
