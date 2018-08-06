package julius

/*
#cgo LDFLAGS: -L../../cmodules/julius/libjulius -ljulius -ldl -lpthread -L../../cmodules/julius/libsent -lsent -Wl,-framework -Wl,CoreServices -Wl,-framework -Wl,CoreAudio -Wl,-framework -Wl,AudioUnit -Wl,-framework -Wl,AudioToolbox -lz -lsndfile -liconv -lm
#cgo CFLAGS: -I../../cmodules/julius/libjulius/include -I../../cmodules/julius/libsent/include
#include "julius/juliuslib.h"

void onPass1Frame(Recog *recog, void *data);
void onSegmentBegin(Recog *recog, void *data);
void onSegmentEnd(Recog *recog, void *data);
void onResult(Recog *recog, void *data);
static void _register_callbacks(Recog* recog, void* data) {
	callback_add(recog, CALLBACK_EVENT_PASS1_FRAME, onPass1Frame, data);
	callback_add(recog, CALLBACK_EVENT_SEGMENT_BEGIN, onSegmentBegin, data);
	callback_add(recog, CALLBACK_EVENT_SEGMENT_END, onSegmentEnd, data);
	callback_add(recog, CALLBACK_RESULT, onResult, data);
}

static Sentence* _read_sentence_array(Sentence* p, int index) {
	return p + index;
}
static float _read_float_array(float* p, int index) {
	return p[index];
}
static int _read_int_array(int* p, int index) {
	return p[index];
}
static unsigned char _read_uchar_array(unsigned char* p, int index) {
	return p[index];
}
static HMM_Logical* _read_hmm_logical(HMM_Logical** p, int index) {
	return p[index];
}
static HMM_Logical** _read_hmm_logical_ptr(HMM_Logical*** p, int index) {
	return p[index];
}
*/
import "C"
import (
	"fmt"
	"log"
	"strings"
	"time"
	"unsafe"

	"github.com/but80/talklistener/internal/globalopt"
)

func cStringArray(a []string) []*C.char {
	result := make([]*C.char, len(a))
	for i, s := range a {
		result[i] = C.CString(s)
	}
	return result
}

const (
	frameShiftSize = 0.01
	frameSize      = 0.025
	offsetAlign    = frameSize / 2.0
)

type Segment struct {
	BeginFrame int
	EndFrame   int
	BeginTime  float64
	EndTime    float64
	Unit       string
	Score      float64
}

type Result struct {
	Dictation [][]string
	Segments  []Segment
	frame     int
	totalSec  float64
	completed bool
}

var OnProgress func(float64, float64)

func (result *Result) DictationString() string {
	s := ""
	for _, dic := range result.Dictation {
		dic = phoneticToKana(dic)
		s += joinKana(dic) + "\n"
	}
	return s
}

func run(argv []string, wavfile string) (*Result, error) {
	if globalopt.Debug {
		C.j_enable_debug_message()
	}
	if !globalopt.Verbose {
		C.jlog_set_output(nil)
	}

	jconf := C.j_config_load_args_new(
		C.int(len(argv)),
		(**C.char)(&cStringArray(argv)[0]),
	)
	if jconf == nil {
		return nil, fmt.Errorf("Julius: 設定に失敗しました")
	}

	recog := C.j_create_instance_from_jconf(jconf)
	if jconf == nil {
		return nil, fmt.Errorf("Julius: 認識器作成に失敗しました")
	}

	result := &Result{}
	C._register_callbacks(recog, unsafe.Pointer(result))

	if C.j_adin_init(recog) == 0 {
		return nil, fmt.Errorf("Julius: 音声認識ストリームの初期化に失敗しました")
	}

	C.j_recog_info(recog)

	if ret := C.j_open_stream(recog, C.CString(wavfile)); ret != 0 {
		return nil, fmt.Errorf("Julius: 音声ファイルのオープンに失敗しました")
	}

	for {
		ret := C.j_recognize_stream(recog)
		if ret == 0 {
			break
		} else if ret == 1 {
			continue
		}
		return nil, fmt.Errorf("Julius: 音声認識に失敗しました")
	}
	for !result.completed {
		time.Sleep(10 * time.Millisecond)
	}

	C.j_recog_free(recog)
	return result, nil
}

//export onPass1Frame
func onPass1Frame(recog *C.Recog, data unsafe.Pointer) {
	result := (*Result)(data)
	samples := int(recog.speechlen)
	rate := int(recog.jconf.input.sfreq)
	totalSec := float64(samples) / float64(rate)
	OnProgress(float64(result.frame)*frameShiftSize, totalSec)
	result.frame++
}

//export onSegmentBegin
func onSegmentBegin(recog *C.Recog, data unsafe.Pointer) {
}

//export onSegmentEnd
func onSegmentEnd(recog *C.Recog, data unsafe.Pointer) {
}

//export onResult
func onResult(recog *C.Recog, data unsafe.Pointer) {
	result := (*Result)(data)
	defer func() {
		result.completed = true
	}()
	if recog == nil {
		return
	}
	proc := recog.process_list
	for proc != nil {
		result.update(proc)
		proc = proc.next
	}
}

func centerName(s string) string {
	if i := strings.Index(s, "-"); 0 <= i {
		s = s[i+1:]
	}
	if i := strings.LastIndex(s, "+"); 0 <= i {
		s = s[:i]
	}
	if i := strings.Index(s, "_"); 0 <= i {
		s = s[:i]
	}
	return s
}

func (result *Result) update(proc *C.RecogProcess) {
	if proc == nil || proc.live == 0 || proc.result.status < 0 || proc.lm == nil {
		return
	}
	winfo := proc.lm.winfo
	if winfo == nil {
		return
	}
	sentnum := int(proc.result.sentnum)
	for i := 0; i < sentnum; i++ {
		if proc.result.sent == nil {
			continue
		}
		log.Printf(
			"debug: num_frame=%d length_msec=%d",
			int(proc.result.num_frame),   ///< Number of frames of the recognized part
			int(proc.result.length_msec), ///< Length of the recognized part
		)
		sent := C._read_sentence_array(proc.result.sent, C.int(i))
		if sent == nil {
			continue
		}
		if unsafe.Pointer(&sent.word[0]) != nil {
			seqnum := int(sent.word_num)
			if len(sent.word) < seqnum {
				seqnum = len(sent.word)
			}
			dictation := []string{}
			for i := 0; i < seqnum; i++ {
				w := int(sent.word[i])
				wl := int(C._read_uchar_array(winfo.wlen, C.int(w)))
				for j := 0; j < wl; j++ {
					p := C._read_hmm_logical_ptr(winfo.wseq, C.int(w))
					if p == nil {
						continue
					}
					ws := C._read_hmm_logical(p, C.int(j))
					if ws == nil {
						continue
					}
					unit := C.GoString(ws.name)
					dictation = append(dictation, centerName(unit))
				}
			}
			result.Dictation = append(result.Dictation, dictation)
		}

		for align := sent.align; align != nil; align = align.next {
			segnum := int(align.num)
			for i := 0; i < segnum; i++ {
				score := float64(C._read_float_array(align.avgscore, C.int(i)))
				if align.unittype == C.PER_PHONEME {
					begin := int(C._read_int_array(align.begin_frame, C.int(i)))
					end := int(C._read_int_array(align.end_frame, C.int(i)))
					p := C._read_hmm_logical(align.ph, C.int(i))
					unit := C.GoString(p.name)
					// if p.is_pseudo != 0 {
					//   fmt.Printf("{%s}", p.name)
					// } else if strmatch(p.name, p.body.defined.name) {
					//   fmt.Printf("%s", p.name)
					// } else {
					//   fmt.Printf("%s[%s]", p.name, p.body.defined.name)
					// }
					seg := Segment{
						BeginFrame: begin,
						EndFrame:   end,
						BeginTime:  float64(begin)*frameShiftSize + offsetAlign,
						EndTime:    float64(end+1)*frameShiftSize + offsetAlign + frameSize,
						Unit:       centerName(unit),
						Score:      score,
					}
					result.Segments = append(result.Segments, seg)
				}
			}
		}
	}
}
