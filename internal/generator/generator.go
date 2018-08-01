package generator

import (
	"fmt"
	"io/ioutil"
	"math"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/but80/talklistener/internal/julius"
	"github.com/but80/talklistener/internal/vsqx"
	"github.com/pkg/errors"
)

const (
	f0FramePeriod    = .005
	resampleRate     = 5
	notesFramePeriod = f0FramePeriod / float64(resampleRate)
	resolution       = 480
	bpm              = 125.00
	tickTime         = 60.0 / bpm / float64(resolution) // = 0.001
	splitConsonant   = true
	useLPF           = 1
)

var vowels = map[string]int{
	"a": 0,
	"i": 1,
	"u": 2,
	"e": 3,
	"o": 4,
}

type ConsonantDef struct {
	Phoneme string
	Kana    []string
}

var consonantDefs = map[string]*ConsonantDef{
	"":   {Phoneme: "", Kana: []string{"あ", "い", "う", "え", "お"}},
	"k":  {Phoneme: "k", Kana: []string{"か", "き", "く", "け", "こ"}},
	"s":  {Phoneme: "s", Kana: []string{"さ", "すぃ", "す", "せ", "そ"}},
	"t":  {Phoneme: "t", Kana: []string{"た", "てぃ", "とぅ", "て", "と"}},
	"n":  {Phoneme: "n", Kana: []string{"な", "に", "ぬ", "ね", "の"}},
	"h":  {Phoneme: "h", Kana: []string{"は", "ひ", "ふ", "へ", "ほ"}},
	"f":  {Phoneme: `p\`, Kana: []string{"ふぁ", "ふぃ", "ふ", "ふぇ", "ふぉ"}},
	"m":  {Phoneme: "m", Kana: []string{"ま", "み", "む", "め", "も"}},
	"y":  {Phoneme: "j", Kana: []string{"や", "い", "ゆ", "いぇ", "よ"}},
	"r":  {Phoneme: "4", Kana: []string{"ら", "り", "る", "れ", "ろ"}},
	"w":  {Phoneme: "w", Kana: []string{"わ", "うぃ", "う", "うぇ", "うぉ"}},
	"g":  {Phoneme: "g", Kana: []string{"が", "ぎ", "ぐ", "げ", "ご"}},
	"z":  {Phoneme: "z", Kana: []string{"ざ", "ずぃ", "ず", "ぜ", "ぞ"}},
	"d":  {Phoneme: "d", Kana: []string{"だ", "でぃ", "どぅ", "で", "ど"}},
	"b":  {Phoneme: "b", Kana: []string{"ば", "び", "ぶ", "べ", "ぼ"}},
	"ky": {Phoneme: "k'", Kana: []string{"きゃ", "き", "きゅ", "きぇ", "きょ"}},
	"sh": {Phoneme: "S", Kana: []string{"しゃ", "し", "しゅ", "しぇ", "しょ"}},
	"ty": {Phoneme: "t'", Kana: []string{"てゃ", "てぃ", "てゅ", "て", "てょ"}},
	"ch": {Phoneme: "tS", Kana: []string{"ちゃ", "ち", "ちゅ", "ちぇ", "ちょ"}},
	"ny": {Phoneme: "n'", Kana: []string{"にゃ", "に", "にゅ", "にぇ", "にょ"}},
	"hy": {Phoneme: "C", Kana: []string{"ひゃ", "ひ", "ひゅ", "ひぇ", "ひょ"}},
	"my": {Phoneme: "m'", Kana: []string{"みゃ", "み", "みゅ", "みぇ", "みょ"}},
	"ry": {Phoneme: "4'", Kana: []string{"りゃ", "り", "りゅ", "りぇ", "りょ"}},
	"gy": {Phoneme: "g'", Kana: []string{"ぎゃ", "ぎ", "ぎゅ", "ぎぇ", "ぎょ"}},
	"j":  {Phoneme: "dZ", Kana: []string{"じゃ", "じ", "じゅ", "じぇ", "じょ"}},
	"dy": {Phoneme: "d'", Kana: []string{"でゃ", "でぃ", "でゅ", "で", "でょ"}},
	"by": {Phoneme: "b'", Kana: []string{"びゃ", "び", "びゅ", "びぇ", "びょ"}},
	"p":  {Phoneme: "p", Kana: []string{"ぱ", "ぴ", "ぷ", "ぺ", "ぽ"}},
	"py": {Phoneme: "p'", Kana: []string{"ぴゃ", "ぴ", "ぴゅ", "ぴぇ", "ぴょ"}},
	"ts": {Phoneme: "ts", Kana: []string{"つぁ", "つぃ", "つ", "つぇ", "つぉ"}},
	"zy": {Phoneme: "z'", Kana: []string{"ずゃ", "ずぃ", "ず", "ずぇ", "ずぉ"}},
}

var specials = map[string]string{
	"q":    "",
	"sp":   "",
	"silB": "",
	"silE": "",
	"N":    "ん",
}

func timeToTick(time float64) int {
	return int(math.Round(time / tickTime))
}

func durationToVelocity(dur float64) int {
	velocity := 64 - int(math.Round(dur*.0)) // TODO: 子音の長さをベロシティに反映
	if velocity < 1 {
		velocity = 1
	} else if 127 < velocity {
		velocity = 127
	}
	return velocity
}

type generator struct {
	noteCenter int
	vsqx       *vsqx.VSQ3

	consonant          string
	consonantBeginTime float64
	consonantEndTime   float64
	vowel              string
	vowelBeginTime     float64
	vowelEndTime       float64
}

func (gen *generator) reset() {
	gen.consonant = ""
	gen.consonantBeginTime = -1.0
	gen.consonantEndTime = -1.0
	gen.vowel = ""
	gen.vowelBeginTime = -1.0
	gen.vowelEndTime = -1.0
}

func (gen *generator) setConsonant(begin, end float64, unit string) {
	gen.consonant = unit
	gen.consonantBeginTime = begin
	gen.consonantEndTime = end
}

func (gen *generator) setVowel(begin, end float64, unit string) {
	gen.vowelBeginTime = begin
	gen.vowelEndTime = end
	gen.vowel = unit
}

func (gen *generator) flush() error {
	var ok bool
	vowelIndex := -1
	var cdef *ConsonantDef

	if gen.vowel != "" {
		if vowelIndex, ok = vowels[gen.vowel]; !ok {
			return fmt.Errorf("[%s] [%s] の発音を特定できません", gen.consonant, gen.vowel)
		}
	}
	if gen.consonant != "" {
		if cdef, ok = consonantDefs[gen.consonant]; !ok || len(cdef.Kana) <= vowelIndex {
			return fmt.Errorf("[%s] [%s] の発音を特定できません", gen.consonant, gen.vowel)
		}
	}

	if gen.consonant == "" || gen.consonantBeginTime < .0 {
		if gen.vowel == "" || gen.vowelBeginTime < .0 {
			// 何もない
		} else {
			// 母音のみ
			gen.vsqx.AddNote(
				64,
				timeToTick(gen.vowelBeginTime),
				timeToTick(gen.vowelEndTime),
				gen.noteCenter,
				consonantDefs[""].Kana[vowelIndex],
				"",
			)
		}
	} else {
		if gen.vowel == "" || gen.vowelBeginTime < .0 {
			// 子音のみ
			gen.vsqx.AddNote(
				durationToVelocity(gen.consonantEndTime-gen.consonantBeginTime),
				timeToTick(gen.consonantBeginTime),
				timeToTick(gen.consonantEndTime),
				gen.noteCenter,
				gen.consonant,
				cdef.Phoneme,
			)
		} else {
			// 子音＋母音
			gen.vsqx.AddNote(
				durationToVelocity(gen.vowelBeginTime-gen.consonantBeginTime),
				timeToTick((gen.consonantBeginTime+gen.vowelBeginTime)/2.0),
				timeToTick(gen.vowelEndTime),
				gen.noteCenter,
				cdef.Kana[vowelIndex],
				"",
			)
		}
	}
	gen.reset()
	return nil
}

func (gen *generator) feedPitchBends(notes []float64) {
	bendSense := 24
	gen.vsqx.AddMCtrl(.0, "PBS", bendSense)
	t := .0
	for _, note := range notes {
		tick := int(math.Round(t / tickTime))
		dNote := note - float64(gen.noteCenter)
		bend := int(math.Round(8192.0 * dNote / float64(bendSense)))
		if bend < -8192 {
			bend = -8192
		} else if 8191 < bend {
			bend = 8191
		}
		gen.vsqx.AddMCtrl(tick, "PIT", bend)
		t += notesFramePeriod
	}
}

func (gen *generator) save(filename string) error {
	return ioutil.WriteFile(filename, gen.vsqx.Bytes(), 0644)
}

func (gen *generator) dump() {
	fmt.Println(string(gen.vsqx.Bytes()))
}

func convertAudioFile(in, out string) error {
	cmd := exec.Command(
		"/usr/bin/afconvert",
		"-f", "WAVE",
		"-d", "I16@16000",
		"-c", "1",
		"-o", out,
		in,
	)
	if captured, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err,
			"afconvert の実行中にエラーが発生しました:\n%s\n%s",
			strings.Join(cmd.Args, " "), captured,
		)
	}
	return nil
}

func Generate(wavfile, wordsfile, dictationKitName, outfile string) error {
	noteOffset := .0

	name := filepath.Base(wavfile)
	tmpdir, err := ioutil.TempDir("", name)
	if err != nil {
		return err
	}
	tmpprefix := filepath.Join(tmpdir, name)

	convertedWavFile := tmpprefix + ".wav"
	if err := convertAudioFile(wavfile, convertedWavFile); err != nil {
		return errors.Wrap(err, "音声ファイルの変換に失敗しました")
	}

	var wg sync.WaitGroup
	errch := make(chan error, 9)

	wg.Add(1)
	noteCenter := int(a3Note)
	notesDelay := .0
	notes := []float64{}
	go func() {
		defer wg.Done()
		var err error
		notes, err = wavToF0Note(convertedWavFile, tmpprefix+".f0", f0FramePeriod)
		if err != nil {
			errch <- err
			return
		}
		noteMin := 128.0
		noteMax := .0
		for i := range notes {
			notes[i] += noteOffset
			note := notes[i]
			if note < noteMin {
				noteMin = note
			}
			if noteMax < note {
				noteMax = note
			}
		}
		noteCenter = int(math.Round((noteMax + noteMin) / 2.0))

		notes = resample(notes, resampleRate)
		if 0 <= useLPF {
			notes = convolve(notes, firLPF[useLPF])
			notesDelay = float64(len(firLPF[useLPF])) / 2.0 * notesFramePeriod
		}
	}()

	wg.Add(1)
	var result *julius.Result
	go func() {
		defer wg.Done()
		var err error
		if wordsfile == "" {
			result, err = julius.Dictate(convertedWavFile, dictationKitName)
		} else {
			result, err = julius.Segmentate(convertedWavFile, wordsfile, tmpprefix)
		}
		if err != nil {
			errch <- err
			return
		}
	}()

	wg.Wait()
	if 0 < len(errch) {
		return <-errch
	}

	gen := generator{
		noteCenter: noteCenter,
		vsqx:       vsqx.New(resolution, bpm),
	}
	gen.reset()

	segsData := ""
	for _, seg := range result.Segments {
		segsData += fmt.Sprintf("%.7f %.7f %s\n", seg.BeginTime, seg.EndTime, seg.Unit)

		unit := seg.Unit
		long := strings.HasSuffix(unit, ":")
		if long {
			unit = unit[:len(unit)-1]
		}
		beginTime := seg.BeginTime + notesDelay
		endTime := seg.EndTime + notesDelay

		if s, ok := specials[unit]; ok {
			gen.flush()
			if s != "" {
				gen.vsqx.AddNote(
					64,
					timeToTick(beginTime),
					timeToTick(endTime),
					gen.noteCenter,
					s,
					"",
				)
			}
			gen.reset()
			continue
		}

		if _, ok := vowels[unit]; !ok {
			if gen.consonant != "" {
				gen.flush()
			}
			gen.setConsonant(beginTime, endTime, unit)
			continue
		}

		if splitConsonant {
			gen.flush()
		}
		gen.setVowel(beginTime, endTime, unit)
		if err := gen.flush(); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(tmpprefix+".seg", []byte(segsData), 0644); err != nil {
		return err
	}

	gen.feedPitchBends(notes)
	if outfile == "" {
		gen.dump()
		return nil
	}
	return gen.save(outfile)
}
