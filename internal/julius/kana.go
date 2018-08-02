package julius

import (
	"regexp"
	"strings"
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

var Vowels = map[string]int{
	"a": 0,
	"i": 1,
	"u": 2,
	"e": 3,
	"o": 4,
}

type Consonant struct {
	VSQXPhoneme string
	Kana        []string
}

var Consonants = map[string]*Consonant{
	"":   {VSQXPhoneme: "", Kana: []string{"あ", "い", "う", "え", "お"}},
	"k":  {VSQXPhoneme: "k", Kana: []string{"か", "き", "く", "け", "こ"}},
	"s":  {VSQXPhoneme: "s", Kana: []string{"さ", "すぃ", "す", "せ", "そ"}},
	"t":  {VSQXPhoneme: "t", Kana: []string{"た", "てぃ", "とぅ", "て", "と"}},
	"n":  {VSQXPhoneme: "n", Kana: []string{"な", "に", "ぬ", "ね", "の"}},
	"h":  {VSQXPhoneme: "h", Kana: []string{"は", "ひ", "ふ", "へ", "ほ"}},
	"f":  {VSQXPhoneme: `p\`, Kana: []string{"ふぁ", "ふぃ", "ふ", "ふぇ", "ふぉ"}},
	"m":  {VSQXPhoneme: "m", Kana: []string{"ま", "み", "む", "め", "も"}},
	"y":  {VSQXPhoneme: "j", Kana: []string{"や", "い", "ゆ", "いぇ", "よ"}},
	"r":  {VSQXPhoneme: "4", Kana: []string{"ら", "り", "る", "れ", "ろ"}},
	"w":  {VSQXPhoneme: "w", Kana: []string{"わ", "うぃ", "う", "うぇ", "うぉ"}},
	"g":  {VSQXPhoneme: "g", Kana: []string{"が", "ぎ", "ぐ", "げ", "ご"}},
	"z":  {VSQXPhoneme: "z", Kana: []string{"ざ", "ずぃ", "ず", "ぜ", "ぞ"}},
	"d":  {VSQXPhoneme: "d", Kana: []string{"だ", "でぃ", "どぅ", "で", "ど"}},
	"b":  {VSQXPhoneme: "b", Kana: []string{"ば", "び", "ぶ", "べ", "ぼ"}},
	"ky": {VSQXPhoneme: "k'", Kana: []string{"きゃ", "き", "きゅ", "きぇ", "きょ"}},
	"sh": {VSQXPhoneme: "S", Kana: []string{"しゃ", "し", "しゅ", "しぇ", "しょ"}},
	"ty": {VSQXPhoneme: "t'", Kana: []string{"てゃ", "てぃ", "てゅ", "て", "てょ"}},
	"ch": {VSQXPhoneme: "tS", Kana: []string{"ちゃ", "ち", "ちゅ", "ちぇ", "ちょ"}},
	"ny": {VSQXPhoneme: "n'", Kana: []string{"にゃ", "に", "にゅ", "にぇ", "にょ"}},
	"hy": {VSQXPhoneme: "C", Kana: []string{"ひゃ", "ひ", "ひゅ", "ひぇ", "ひょ"}},
	"my": {VSQXPhoneme: "m'", Kana: []string{"みゃ", "み", "みゅ", "みぇ", "みょ"}},
	"ry": {VSQXPhoneme: "4'", Kana: []string{"りゃ", "り", "りゅ", "りぇ", "りょ"}},
	"gy": {VSQXPhoneme: "g'", Kana: []string{"ぎゃ", "ぎ", "ぎゅ", "ぎぇ", "ぎょ"}},
	"j":  {VSQXPhoneme: "dZ", Kana: []string{"じゃ", "じ", "じゅ", "じぇ", "じょ"}},
	"dy": {VSQXPhoneme: "d'", Kana: []string{"でゃ", "でぃ", "でゅ", "で", "でょ"}},
	"by": {VSQXPhoneme: "b'", Kana: []string{"びゃ", "び", "びゅ", "びぇ", "びょ"}},
	"p":  {VSQXPhoneme: "p", Kana: []string{"ぱ", "ぴ", "ぷ", "ぺ", "ぽ"}},
	"py": {VSQXPhoneme: "p'", Kana: []string{"ぴゃ", "ぴ", "ぴゅ", "ぴぇ", "ぴょ"}},
	"ts": {VSQXPhoneme: "ts", Kana: []string{"つぁ", "つぃ", "つ", "つぇ", "つぉ"}},
	"zy": {VSQXPhoneme: "z'", Kana: []string{"ずゃ", "ずぃ", "ず", "ずぇ", "ずぉ"}},
}

var specials = map[string]string{
	"q":    "っ",
	"sp":   "sp",
	"silB": "",
	"silE": "",
	"N":    "ん",
}

var SpecialsForVSQX = map[string]string{
	"q":    "",
	"sp":   "",
	"silB": "",
	"silE": "",
	"N":    "ん",
}

func splitLong(vs string) (string, bool) {
	if strings.HasSuffix(vs, ":") {
		return vs[:len(vs)-1], true
	}
	return vs, false
}

func splitVowel(vs string) (vi int, long bool, ok bool) {
	vs, long = splitLong(vs)
	if vi, ok = Vowels[vs]; ok {
		return
	}
	vi = -1
	return
}

func phoneticToKana(p []string) []string {
	result := []string{}
	for i := 0; i < len(p); i++ {
		v, long, ok := splitVowel(p[i])
		if ok {
			result = append(result, Consonants[""].Kana[v])
			if long {
				result = append(result, "ー")
			}
			continue
		}
		c := p[i]
		if i+1 < len(p) {
			v, long, _ = splitVowel(p[i+1])
		}
		cons, ok := Consonants[c]
		if ok && 0 <= v && v < len(cons.Kana) {
			result = append(result, cons.Kana[v])
			if long {
				result = append(result, "ー")
			}
			i++
			continue
		}
		if s, ok := specials[c]; ok {
			c = s
		}
		if c != "" {
			result = append(result, c)
		}
	}
	return result
}

var joinKanaRx1 = regexp.MustCompile(`^\w+$`)
var joinKanaRx2 = regexp.MustCompile(`\s+`)

func joinKana(k []string) string {
	result := ""
	for _, s := range k {
		if joinKanaRx1.MatchString(s) {
			result += " " + s + " "
		} else {
			result += s
		}
	}
	return joinKanaRx2.ReplaceAllString(strings.TrimSpace(result), " ")
}
