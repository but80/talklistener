package julius

/*
#cgo LDFLAGS: -L../../cmodules/julius/libjulius -ljulius -ldl -lpthread -L../../cmodules/julius/libsent -lsent -Wl,-framework -Wl,CoreServices -Wl,-framework -Wl,CoreAudio -Wl,-framework -Wl,AudioUnit -Wl,-framework -Wl,AudioToolbox -lz -lsndfile -liconv -lm
#cgo CFLAGS: -I../../cmodules/julius/libjulius/include -I../../cmodules/julius/libsent/include
#include "julius/juliuslib.h"

void onResult(Recog *recog, void *data);

static void _register_callback_result(Recog* recog, void* data) {
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
	"strings"
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
	framePeriod = 0.01
	offsetAlign = 0.0125 // offset for result in ms: 25ms / 2
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
}

func (result *Result) DictationString() string {
	s := ""
	for _, dic := range result.Dictation {
		s += strings.Join(dic, " ") + "\n"
	}
	return s
}

func run(argv []string, wavfile string) (*Result, error) {
	if globalopt.Verbose {
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
	C._register_callback_result(recog, unsafe.Pointer(result))

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

	C.j_recog_free(recog)
	return result, nil
}

//export onResult
func onResult(recog *C.Recog, data unsafe.Pointer) {
	if recog == nil {
		return
	}
	result := (*Result)(data)
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
						BeginTime:  float64(begin) * framePeriod,
						EndTime:    float64(end+1)*framePeriod + offsetAlign,
						Unit:       unit,
						Score:      score,
					}
					if begin != 0 {
						seg.BeginTime += offsetAlign
					}
					result.Segments = append(result.Segments, seg)
				}
			}
		}
	}
}
