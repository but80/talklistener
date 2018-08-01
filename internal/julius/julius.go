package julius

/*
#cgo LDFLAGS: -L../../cmodules/julius/libjulius -ljulius -ldl -lpthread -L../../cmodules/julius/libsent -lsent -Wl,-framework -Wl,CoreServices -Wl,-framework -Wl,CoreAudio -Wl,-framework -Wl,AudioUnit -Wl,-framework -Wl,AudioToolbox -lz -lsndfile -liconv -lm
#cgo CFLAGS: -I../../cmodules/julius/libjulius/include -I../../cmodules/julius/libsent/include
#include "julius/juliuslib.h"

void onResult(Recog *recog, void *data);

static void _register_callback_result(Recog *recog, void *data) {
	callback_add(recog, CALLBACK_RESULT, onResult, data);
}
*/
import "C"
import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/but80/talklistener/internal/globalopt"
)

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
}

type Result struct {
	Dictation [][]string
	Segments  []Segment
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
	proc := recog.process_list
	for proc != nil {
		outputResult(proc, (*Result)(data))
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

func outputResult(proc *C.RecogProcess, result *Result) {
	if proc.live == 0 {
		return
	}
	if proc.result.status < 0 {
		fmt.Printf("no results obtained: %d\n", proc.result.status)
		return
	}
	winfo := proc.lm.winfo
	sentnum := int(proc.result.sentnum)
	for i := 0; i < sentnum; i++ {
		sent := readSentence(proc.result.sent, i)

		if unsafe.Pointer(&sent.word[0]) != nil {
			seqnum := int(sent.word_num)
			dictation := []string{}
			for i := 0; i < seqnum; i++ {
				w := int(sent.word[i])
				wl := int(readCUCharArray(winfo.wlen, w))
				for j := 0; j < wl; j++ {
					p := readHMMLogicalPtr(winfo.wseq, w)
					ws := readHMMLogical(p, j)
					unit := C.GoString(ws.name)
					dictation = append(dictation, centerName(unit))
				}
			}
			result.Dictation = append(result.Dictation, dictation)
		}

		for align := sent.align; align != nil; align = align.next {
			for i := 0; i < int(align.num); i++ {
				// align.avgscore[i]
				if align.unittype == C.PER_PHONEME {
					begin := int(readCIntArray(align.begin_frame, i))
					end := int(readCIntArray(align.end_frame, i))
					p := readHMMLogical(align.ph, i)
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

func cStringArray(a []string) []*C.char {
	result := make([]*C.char, len(a))
	for i, s := range a {
		result[i] = C.CString(s)
	}
	return result
}

func readHMMLogical(p **C.HMM_Logical, index int) *C.HMM_Logical {
	size := unsafe.Sizeof(*p)
	ptr := unsafe.Pointer(uintptr(unsafe.Pointer(p)) + size*uintptr(index))
	return *(**C.HMM_Logical)(ptr)
}

func readHMMLogicalPtr(p ***C.HMM_Logical, index int) **C.HMM_Logical {
	size := unsafe.Sizeof(*p)
	ptr := unsafe.Pointer(uintptr(unsafe.Pointer(p)) + size*uintptr(index))
	return *(***C.HMM_Logical)(ptr)
}

func readSentence(p *C.Sentence, index int) *C.Sentence {
	size := unsafe.Sizeof(*p)
	ptr := unsafe.Pointer(uintptr(unsafe.Pointer(p)) + size*uintptr(index))
	return (*C.Sentence)(ptr)
}

func readCIntArray(p *C.int, index int) C.int {
	size := unsafe.Sizeof(*p)
	ptr := unsafe.Pointer(uintptr(unsafe.Pointer(p)) + size*uintptr(index))
	return *(*C.int)(ptr)
}

func readCUCharArray(p *C.uchar, index int) C.uchar {
	size := unsafe.Sizeof(*p)
	ptr := unsafe.Pointer(uintptr(unsafe.Pointer(p)) + size*uintptr(index))
	return *(*C.uchar)(ptr)
}
