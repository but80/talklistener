package julius

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/but80/talklistener/internal/assets"
)

const (
	hmmDefs = "/cmodules/segmentation-kit/models/hmmdefs_monof_mix16_gid.binhmm" // monophone model
	// hmmDefs = "/cmodules/segmentation-kit/models/hmmdefs_ptm_gid.binhmm" // triphone model
)

var kanaToPhoneticTable = [][2]string{
	// 3文字以上からなる変換規則
	{"う゛ぁ", " b a"},
	{"う゛ぃ", " b i"},
	{"う゛ぇ", " b e"},
	{"う゛ぉ", " b o"},
	{"う゛ゅ", " by u"},

	// 2文字からなる変換規則
	{"ぅ゛", " b u"},

	{"あぁ", " a a"},
	{"いぃ", " i i"},
	{"いぇ", " i e"},
	{"いゃ", " y a"},
	{"うぅ", " u:"},
	{"えぇ", " e e"},
	{"おぉ", " o:"},
	{"かぁ", " k a:"},
	{"きぃ", " k i:"},
	{"くぅ", " k u:"},
	{"くゃ", " ky a"},
	{"くゅ", " ky u"},
	{"くょ", " ky o"},
	{"けぇ", " k e:"},
	{"こぉ", " k o:"},
	{"がぁ", " g a:"},
	{"ぎぃ", " g i:"},
	{"ぐぅ", " g u:"},
	{"ぐゃ", " gy a"},
	{"ぐゅ", " gy u"},
	{"ぐょ", " gy o"},
	{"げぇ", " g e:"},
	{"ごぉ", " g o:"},
	{"さぁ", " s a:"},
	{"しぃ", " sh i:"},
	{"すぅ", " s u:"},
	{"すゃ", " sh a"},
	{"すゅ", " sh u"},
	{"すょ", " sh o"},
	{"せぇ", " s e:"},
	{"そぉ", " s o:"},
	{"ざぁ", " z a:"},
	{"じぃ", " j i:"},
	{"ずぅ", " z u:"},
	{"ずゃ", " zy a"},
	{"ずゅ", " zy u"},
	{"ずょ", " zy o"},
	{"ぜぇ", " z e:"},
	{"ぞぉ", " z o:"},
	{"たぁ", " t a:"},
	{"ちぃ", " ch i:"},
	{"つぁ", " ts a"},
	{"つぃ", " ts i"},
	{"つぅ", " ts u:"},
	{"つゃ", " ch a"},
	{"つゅ", " ch u"},
	{"つょ", " ch o"},
	{"つぇ", " ts e"},
	{"つぉ", " ts o"},
	{"てぇ", " t e:"},
	{"とぉ", " t o:"},
	{"だぁ", " d a:"},
	{"ぢぃ", " j i:"},
	{"づぅ", " d u:"},
	{"づゃ", " zy a"},
	{"づゅ", " zy u"},
	{"づょ", " zy o"},
	{"でぇ", " d e:"},
	{"どぉ", " d o:"},
	{"なぁ", " n a:"},
	{"にぃ", " n i:"},
	{"ぬぅ", " n u:"},
	{"ぬゃ", " ny a"},
	{"ぬゅ", " ny u"},
	{"ぬょ", " ny o"},
	{"ねぇ", " n e:"},
	{"のぉ", " n o:"},
	{"はぁ", " h a:"},
	{"ひぃ", " h i:"},
	{"ふぅ", " f u:"},
	{"ふゃ", " hy a"},
	{"ふゅ", " hy u"},
	{"ふょ", " hy o"},
	{"へぇ", " h e:"},
	{"ほぉ", " h o:"},
	{"ばぁ", " b a:"},
	{"びぃ", " b i:"},
	{"ぶぅ", " b u:"},
	// { "ふゃ", " hy a" },
	{"ぶゅ", " by u"},
	// { "ふょ", " hy o" },
	{"べぇ", " b e:"},
	{"ぼぉ", " b o:"},
	{"ぱぁ", " p a:"},
	{"ぴぃ", " p i:"},
	{"ぷぅ", " p u:"},
	{"ぷゃ", " py a"},
	{"ぷゅ", " py u"},
	{"ぷょ", " py o"},
	{"ぺぇ", " p e:"},
	{"ぽぉ", " p o:"},
	{"まぁ", " m a:"},
	{"みぃ", " m i:"},
	{"むぅ", " m u:"},
	{"むゃ", " my a"},
	{"むゅ", " my u"},
	{"むょ", " my o"},
	{"めぇ", " m e:"},
	{"もぉ", " m o:"},
	{"やぁ", " y a:"},
	{"ゆぅ", " y u:"},
	{"ゆゃ", " y a:"},
	{"ゆゅ", " y u:"},
	{"ゆょ", " y o:"},
	{"よぉ", " y o:"},
	{"らぁ", " r a:"},
	{"りぃ", " r i:"},
	{"るぅ", " r u:"},
	{"るゃ", " ry a"},
	{"るゅ", " ry u"},
	{"るょ", " ry o"},
	{"れぇ", " r e:"},
	{"ろぉ", " r o:"},
	{"わぁ", " w a:"},
	{"をぉ", " o:"},

	{"う゛", " b u"},
	{"でぃ", " d i"},
	// { "でぇ", " d e:" },
	{"でゃ", " dy a"},
	{"でゅ", " dy u"},
	{"でょ", " dy o"},
	{"てぃ", " t i"},
	// { "てぇ", " t e:" },
	{"てゃ", " ty a"},
	{"てゅ", " ty u"},
	{"てょ", " ty o"},
	{"すぃ", " s i"},
	{"ずぁ", " z u a"},
	{"ずぃ", " z i"},
	// { "ずぅ", " z u" },
	// { "ずゃ", " zy a" },
	// { "ずゅ", " zy u" },
	// { "ずょ", " zy o" },
	{"ずぇ", " z e"},
	{"ずぉ", " z o"},
	{"きゃ", " ky a"},
	{"きゅ", " ky u"},
	{"きょ", " ky o"},
	{"しゃ", " sh a"},
	{"しゅ", " sh u"},
	{"しぇ", " sh e"},
	{"しょ", " sh o"},
	{"ちゃ", " ch a"},
	{"ちゅ", " ch u"},
	{"ちぇ", " ch e"},
	{"ちょ", " ch o"},
	{"とぅ", " t u"},
	{"とゃ", " ty a"},
	{"とゅ", " ty u"},
	{"とょ", " ty o"},
	{"どぁ", " d o a"},
	{"どぅ", " d u"},
	{"どゃ", " dy a"},
	{"どゅ", " dy u"},
	{"どょ", " dy o"},
	// { "どぉ", " d o:" },
	{"にゃ", " ny a"},
	{"にゅ", " ny u"},
	{"にょ", " ny o"},
	{"ひゃ", " hy a"},
	{"ひゅ", " hy u"},
	{"ひょ", " hy o"},
	{"みゃ", " my a"},
	{"みゅ", " my u"},
	{"みょ", " my o"},
	{"りゃ", " ry a"},
	{"りゅ", " ry u"},
	{"りょ", " ry o"},
	{"ぎゃ", " gy a"},
	{"ぎゅ", " gy u"},
	{"ぎょ", " gy o"},
	{"ぢぇ", " j e"},
	{"ぢゃ", " j a"},
	{"ぢゅ", " j u"},
	{"ぢょ", " j o"},
	{"じぇ", " j e"},
	{"じゃ", " j a"},
	{"じゅ", " j u"},
	{"じょ", " j o"},
	{"びゃ", " by a"},
	{"びゅ", " by u"},
	{"びょ", " by o"},
	{"ぴゃ", " py a"},
	{"ぴゅ", " py u"},
	{"ぴょ", " py o"},
	{"うぁ", " u a"},
	{"うぃ", " w i"},
	{"うぇ", " w e"},
	{"うぉ", " w o"},
	{"ふぁ", " f a"},
	{"ふぃ", " f i"},
	// { "ふぅ", " f u" },
	// { "ふゃ", " hy a" },
	// { "ふゅ", " hy u" },
	// { "ふょ", " hy o" },
	{"ふぇ", " f e"},
	{"ふぉ", " f o"},

	// 1音からなる変換規則
	{"あ", " a"},
	{"い", " i"},
	{"う", " u"},
	{"え", " e"},
	{"お", " o"},
	{"か", " k a"},
	{"き", " k i"},
	{"く", " k u"},
	{"け", " k e"},
	{"こ", " k o"},
	{"さ", " s a"},
	{"し", " sh i"},
	{"す", " s u"},
	{"せ", " s e"},
	{"そ", " s o"},
	{"た", " t a"},
	{"ち", " ch i"},
	{"つ", " ts u"},
	{"て", " t e"},
	{"と", " t o"},
	{"な", " n a"},
	{"に", " n i"},
	{"ぬ", " n u"},
	{"ね", " n e"},
	{"の", " n o"},
	{"は", " h a"},
	{"ひ", " h i"},
	{"ふ", " f u"},
	{"へ", " h e"},
	{"ほ", " h o"},
	{"ま", " m a"},
	{"み", " m i"},
	{"む", " m u"},
	{"め", " m e"},
	{"も", " m o"},
	{"ら", " r a"},
	{"り", " r i"},
	{"る", " r u"},
	{"れ", " r e"},
	{"ろ", " r o"},
	{"が", " g a"},
	{"ぎ", " g i"},
	{"ぐ", " g u"},
	{"げ", " g e"},
	{"ご", " g o"},
	{"ざ", " z a"},
	{"じ", " j i"},
	{"ず", " z u"},
	{"ぜ", " z e"},
	{"ぞ", " z o"},
	{"だ", " d a"},
	{"ぢ", " j i"},
	{"づ", " z u"},
	{"で", " d e"},
	{"ど", " d o"},
	{"ば", " b a"},
	{"び", " b i"},
	{"ぶ", " b u"},
	{"べ", " b e"},
	{"ぼ", " b o"},
	{"ぱ", " p a"},
	{"ぴ", " p i"},
	{"ぷ", " p u"},
	{"ぺ", " p e"},
	{"ぽ", " p o"},
	{"や", " y a"},
	{"ゆ", " y u"},
	{"よ", " y o"},
	{"わ", " w a"},
	{"ゐ", " i"},
	{"ゑ", " e"},
	{"ん", " N"},
	{"っ", " q"},
	{"ー", ":"},

	// ここまでに処理されてない ぁぃぅぇぉ はそのまま大文字扱い
	{"ぁ", " a"},
	{"ぃ", " i"},
	{"ぅ", " u"},
	{"ぇ", " e"},
	{"ぉ", " o"},
	{"ゎ", " w a"},
	// { "ぉ", " o" },

	//その他特別なルール
	{"を", " o"},
	{`\s*:(\s*:)+`, ":"},
}

func kanaToPhonetic(line string) string {
	line = strings.TrimSpace(line)
	for _, t := range kanaToPhoneticTable {
		rx, _ := regexp.Compile(t[0])
		line = rx.ReplaceAllString(line, t[1])
	}
	line = strings.TrimSpace(line)
	return line
}

func wordsToDict(infile, outfile string) ([]string, error) {
	words, err := loadWords(infile)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(outfile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	for i, w := range words {
		_, err := fmt.Fprintf(file, "%d [w_%d] %s\n", i, i, w)
		if err != nil {
			file.Close()
			return nil, err
		}
	}
	return words, file.Close()
}

func loadWords(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	words := []string{"silB"}
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		words = append(words, kanaToPhonetic(line))
	}
	words = append(words, "silE")
	return words, nil
}

func generateDFA(num int, outfile string) error {
	file, err := os.OpenFile(outfile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	num--
	first := 1
	for i := 0; i <= num; i++ {
		if _, err := fmt.Fprintf(file, "%d %d %d 0 %d\n", i, num-i, i+1, first); err != nil {
			file.Close()
			return err
		}
		first = 0
	}
	if _, err := fmt.Fprintf(file, "%d -1 -1 1 0\n", num+1); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func saveAssetAsFile(assetname, filename string) error {
	file, err := assets.Assets.Open(assetname)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

func Segmentate(wavfile, wordsfile, tmpprefix string) ([]Segment, error) {
	words, err := wordsToDict(wordsfile, tmpprefix+".dict")
	if err != nil {
		return nil, err
	}
	generateDFA(len(words), tmpprefix+".dfa")

	hmmTmpName := tmpprefix + filepath.Base(hmmDefs)
	if err := saveAssetAsFile(hmmDefs, hmmTmpName); err != nil {
		return nil, err
	}

	argv := []string{
		"julius",
		"-h", hmmTmpName, // HMM definition
		"-dfa", tmpprefix + ".dfa", // DFA grammar
		"-v", tmpprefix + ".dict", // dictionary
		"-palign", // optionally output phoneme alignments
		"-input", "file",
	}
	segs, err := Run(argv, wavfile)
	if err != nil {
		return nil, err
	}
	return segs.Segments, nil
}
