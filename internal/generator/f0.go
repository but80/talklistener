package generator

import (
	"fmt"
	"log"
	"math"
	"os"
	"reflect"

	"github.com/but80/talklistener/internal/world"
	"github.com/mjibson/go-dsp/wav"
	"github.com/pkg/errors"
)

const (
	minFreq = 100.0
	a3Freq  = 440.0
	a3Note  = 69.0
)

func loadWav(filename string) ([]float64, int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, 0, err
	}
	w, err := wav.New(file)
	if err != nil {
		return nil, 0, err
	}
	samples, err := w.ReadSamples(w.Samples)
	if err != nil {
		return nil, 0, err
	}
	result := make([]float64, w.Samples)
	switch s := samples.(type) {
	case []int16:
		for i, v := range s {
			result[i] = float64(v) / 32767.0
		}
	default:
		return nil, 0, fmt.Errorf("Unsupported sample size: %s", reflect.TypeOf(samples))
	}
	return result, int(w.SampleRate), nil
}

func wavToF0Note(infile, outfile string, framePeriod float64) ([]float64, error) {
	log.Println("基本周波数を推定中...")

	x, fs, err := loadWav(infile)
	if err != nil {
		return nil, errors.Wrap(err, "音声ファイルの読み込みに失敗しました")
	}

	file, err := os.OpenFile(outfile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, errors.Wrap(err, "一時ファイルの作成に失敗しました")
	}
	defer file.Close()

	f0 := world.Harvest(x, fs, framePeriod)
	n0 := freqToNote(interpolate(f0))
	for i, n := range n0 {
		_, err := fmt.Fprintf(file, "%.7f: %.2f\n", float64(i)*framePeriod, n)
		if err != nil {
			return nil, errors.Wrap(err, "一時ファイルの保存に失敗しました")
		}
	}
	return n0, nil
}

func interpolate(f0 []float64) []float64 {
	last := a3Freq
	for _, f := range f0 {
		if minFreq <= f {
			last = f
			break
		}
	}
	result := make([]float64, len(f0))
	for i, f := range f0 {
		if minFreq <= f {
			result[i] = f
			last = f
		} else {
			result[i] = last
		}
	}
	return result
}

func freqToNote(f0 []float64) []float64 {
	result := make([]float64, len(f0))
	for i, f := range f0 {
		result[i] = math.Log2(f/a3Freq)*12.0 + a3Note
	}
	return result
}

func resample(f0 []float64, n int) []float64 {
	result := make([]float64, (len(f0)-1)*n+1)
	for i := 0; i < len(f0)-1; i++ {
		begin := f0[i]
		end := f0[i+1]
		for j := 0; j < n; j++ {
			result[i*n+j] = (begin*float64(n-j) + end*float64(j)) / float64(n)
		}
	}
	result[(len(f0)-1)*n] = f0[len(f0)-1]
	return result
}
