package generator

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	parallel         = true
	extendNoteTime   = 0.05
	shiftBendTime    = 0.0
	durationAtMinVel = 0.20
	durationAtMaxVel = 0.00
)

func timeToTick(time float64) int {
	return int(math.Round(time / tickTime))
}

func durationToVelocity(dur float64) int {
	velocity := 127 - int(math.Round((dur-durationAtMaxVel)*127.0/(durationAtMinVel-durationAtMaxVel)))
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
	var cons *julius.Consonant

	if gen.vowel != "" {
		if vowelIndex, ok = julius.Vowels[gen.vowel]; !ok {
			return fmt.Errorf("[%s] [%s] の発音を特定できません", gen.consonant, gen.vowel)
		}
	}
	if gen.consonant != "" {
		if cons, ok = julius.Consonants[gen.consonant]; !ok || len(cons.Kana) <= vowelIndex {
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
				timeToTick(gen.vowelEndTime+extendNoteTime),
				gen.noteCenter,
				julius.Consonants[""].Kana[vowelIndex],
				"",
			)
		}
	} else {
		if gen.vowel == "" || gen.vowelBeginTime < .0 {
			// 子音のみ
			gen.vsqx.AddNote(
				durationToVelocity(gen.consonantEndTime-gen.consonantBeginTime),
				timeToTick(gen.consonantBeginTime),
				timeToTick(gen.consonantEndTime+extendNoteTime),
				gen.noteCenter,
				gen.consonant,
				cons.VSQXPhoneme,
			)
		} else {
			// 子音＋母音
			begin := timeToTick(gen.vowelBeginTime)
			end := timeToTick(gen.vowelEndTime + extendNoteTime)
			gen.vsqx.ExtendLastNote(begin, timeToTick(gen.consonantBeginTime))
			gen.vsqx.AddNote(
				durationToVelocity(gen.vowelBeginTime-gen.consonantBeginTime),
				begin,
				end,
				gen.noteCenter,
				cons.Kana[vowelIndex],
				"",
			)
		}
	}
	gen.reset()
	return nil
}

func (gen *generator) feedPitchBends(notes []float64, timeOffset float64) {
	bendSense := 24
	gen.vsqx.AddMCtrl(.0, "PBS", bendSense)
	t := timeOffset
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
	log.Print("info: 音声ファイルのフォーマットを変換中...")
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

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func isNewer(this, that string) bool {
	sThis, err := os.Stat(this)
	if err != nil {
		return false
	}
	sThat, err := os.Stat(that)
	if err != nil {
		return false
	}
	return sThis.ModTime().After(sThat.ModTime())
}

func isEmpty(filename string) bool {
	s, err := os.Stat(filename)
	return err != nil || s.Size() == 0
}

var removeExtRx = regexp.MustCompile(`\.[^\.]+$`)

func removeExt(filename string) string {
	return removeExtRx.ReplaceAllString(filename, "")
}

type GenerateOptions struct {
	AudioFile      string
	TextFile       string
	OutFile        string
	Singer         string
	F0LPFCutoff    string
	DictationModel string
	SplitConsonant bool
	Transpose      int
	Redictate      bool
	Recache        bool
}

// Generate は、話し声を録音した音声ファイルからVocaloid3シーケンスを生成します。
func Generate(opts *GenerateOptions) error {
	if !vsqx.IsValidSinger(opts.Singer) {
		log.Printf("warn: シンガー %s は定義されていません", opts.Singer)
		opts.Singer = vsqx.DefaultSinger
	}

	if !exists(opts.AudioFile) {
		return fmt.Errorf("%s が見つかりません", opts.AudioFile)
	}
	if opts.TextFile == "" {
		opts.TextFile = removeExt(opts.AudioFile) + ".txt"
	}
	if opts.OutFile == "" {
		opts.OutFile = removeExt(opts.AudioFile) + ".vsqx"
	}

	name := filepath.Base(opts.AudioFile)
	objdir := removeExt(opts.AudioFile) + ".tlo"
	if opts.Recache {
		if err := os.RemoveAll(objdir); err != nil {
			return errors.Wrap(err, "キャッシュディレクトリの作成に失敗しました")
		}
	}
	if _, err := os.Stat(objdir); err != nil {
		if err := os.Mkdir(objdir, 0755); err != nil {
			return errors.Wrap(err, "キャッシュディレクトリの作成に失敗しました")
		}
	}
	objPrefix := filepath.Join(objdir, name)

	convertedWavFile := objPrefix + ".wav"
	if isNewer(convertedWavFile, opts.AudioFile) {
		log.Printf("info: フォーマット変換済み音声ファイルのキャッシュを使用します: %s", convertedWavFile)
	} else {
		if err := convertAudioFile(opts.AudioFile, convertedWavFile); err != nil {
			return errors.Wrap(err, "音声ファイルの変換に失敗しました")
		}
	}

	var wg sync.WaitGroup
	errch := make(chan error, 9)

	wg.Add(1)
	noteCenter := int(a3Note)
	notesDelay := .0
	notes := []float64{}
	f0done := false
	go func() {
		defer wg.Done()
		defer func() {
			f0done = true
		}()
		var err error
		f0file := objPrefix + ".f0"
		if isNewer(f0file, convertedWavFile) {
			log.Printf("info: 推定済み基本周波数のキャッシュを使用します: %s", f0file)
			notes, err = loadF0Note(f0file)
		} else {
			notes, err = wavToF0Note(convertedWavFile, f0file, f0FramePeriod)
		}
		if err != nil {
			errch <- errors.Wrap(err, "基本周波数の推定に失敗しました")
			return
		}
		noteMin := 128.0
		noteMax := .0
		noteOffset := float64(opts.Transpose) / 100.0
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

		log.Print("info: 基本周波数の変動をフィルタリング中...")
		notes = resample(notes, resampleRate)
		if opts.F0LPFCutoff != "" {
			notes = convolve(notes, firLPF[opts.F0LPFCutoff])
			notesDelay = float64(len(firLPF[opts.F0LPFCutoff])) / 2.0 * notesFramePeriod
		}
	}()

	if !parallel {
		wg.Wait()
	}

	wg.Add(1)
	var result *julius.Result
	go func() {
		defer wg.Done()
		lastSec := -1
		julius.OnProgress = func(progress, total float64) {
			sec := int(progress/10.0) * 10
			if lastSec != sec {
				log.Printf("info: 進捗: %d / %d 秒", sec, int(math.Ceil(total)))
				lastSec = sec
			}
		}
		var err error
		if isEmpty(opts.TextFile) || opts.Redictate {
			result, err = julius.Dictate(convertedWavFile, opts.DictationModel)
			if err != nil {
				errch <- errors.Wrap(err, "発話内容の推定に失敗しました")
				return
			}
			b := []byte(result.DictationString())
			if len(b) == 0 {
				errch <- errors.Wrap(err, "音声ファイル中に認識可能な発話がありませんでした")
				return
			}
			if err := ioutil.WriteFile(opts.TextFile, b, 0644); err != nil {
				errch <- errors.Wrap(err, "推定した発話内容の保存に失敗しました")
				return
			}
		}
		result, err = julius.Segmentate(convertedWavFile, opts.TextFile, objPrefix)
		if err != nil {
			errch <- errors.Wrap(err, "発音タイミングの推定に失敗しました")
			return
		}
		if !f0done {
			log.Printf("info: 基本周波数の推定が継続中です。しばらくお待ち下さい...")
		}
	}()

	wg.Wait()
	if 0 < len(errch) {
		return <-errch
	}

	gen := generator{
		noteCenter: noteCenter,
		vsqx:       vsqx.New(opts.Singer, resolution, bpm),
	}
	gen.reset()

	log.Print("info: VSQXを生成中...")
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

		if unit == "q" {
			gen.flush()
			gen.vsqx.AddNote(
				64,
				timeToTick(beginTime),
				timeToTick(endTime),
				gen.noteCenter,
				"っ",
				"Sil",
			)
			continue
		}

		if s, ok := julius.SpecialsForVSQX[unit]; ok {
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

		if _, ok := julius.Vowels[unit]; !ok {
			if gen.consonant != "" {
				gen.flush()
			}
			gen.setConsonant(beginTime, endTime, unit)
			continue
		}

		if opts.SplitConsonant {
			if err := gen.flush(); err != nil {
				return errors.Wrap(err, "テキストファイルの内容が不正です")
			}
		}
		gen.setVowel(beginTime, endTime, unit)
		if err := gen.flush(); err != nil {
			return errors.Wrap(err, "テキストファイルの内容が不正です")
		}
	}
	if err := gen.flush(); err != nil {
		return errors.Wrap(err, "テキストファイルの内容が不正です")
	}

	if err := ioutil.WriteFile(objPrefix+".seg", []byte(segsData), 0644); err != nil {
		return errors.Wrap(err, "セグメンテーションキャッシュファイルの保存に失敗しました")
	}

	gen.feedPitchBends(notes, shiftBendTime)
	if err := gen.save(opts.OutFile); err != nil {
		return errors.Wrap(err, "VSQXの保存に失敗しました")
	}

	log.Printf("info: 出力ノート数: %d", gen.vsqx.NoteCount())
	log.Print("info: 完了")
	return nil
}
