package vsqx

import (
	"encoding/xml"
	"io/ioutil"
	"math"
	"sort"
)

// https://github.com/KentoW/json2vsqx を参考にさせていただきました。

type CData struct {
	Data string `xml:",cdata"`
	Lock int    `xml:"lock,attr,omitempty"`
}

type VoiceParam struct {
	XMLName xml.Name `xml:"vVoiceParam"`
	BRE     int      `xml:"bre"`
	BRI     int      `xml:"bri"`
	CLE     int      `xml:"cle"`
	GEN     int      `xml:"gen"`
	OPE     int      `xml:"ope"`
}

type Voice struct {
	XMLName    xml.Name `xml:"vVoice"`
	BS         int      `xml:"vBS"`
	PC         int      `xml:"vPC"`
	CompID     CData    `xml:"compID"`
	VoiceName  CData    `xml:"vVoiceName"`
	VoiceParam VoiceParam
}

type VoiceTable struct {
	XMLName xml.Name `xml:"vVoiceTable"`
	Voice   Voice
}

type MasterUnit struct {
	XMLName  xml.Name `xml:"masterUnit"`
	OutDev   int      `xml:"outDev"`
	RetLevel int      `xml:"retLevel"`
	Vol      int      `xml:"vol"`
}

type VSUnit struct {
	XMLName    xml.Name `xml:"vsUnit"`
	VSTrackNo  int      `xml:"vsTrackNo"`
	InGain     int      `xml:"inGain"`
	SendLevel  int      `xml:"sendLevel"`
	SendEnable int      `xml:"sendEnable"`
	Mute       int      `xml:"mute"`
	Solo       int      `xml:"solo"`
	Pan        int      `xml:"pan"`
	Vol        int      `xml:"vol"`
}

type SEUnit struct {
	XMLName    xml.Name `xml:"seUnit"`
	InGain     int      `xml:"inGain"`
	SendLevel  int      `xml:"sendLevel"`
	SendEnable int      `xml:"sendEnable"`
	Mute       int      `xml:"mute"`
	Solo       int      `xml:"solo"`
	Pan        int      `xml:"pan"`
	Vol        int      `xml:"vol"`
}

type KaraokeUnit struct {
	XMLName xml.Name `xml:"karaokeUnit"`
	InGain  int      `xml:"inGain"`
	Mute    int      `xml:"mute"`
	Solo    int      `xml:"solo"`
	Vol     int      `xml:"vol"`
}

type Mixer struct {
	XMLName     xml.Name `xml:"mixer"`
	MasterUnit  MasterUnit
	VSUnit      VSUnit
	SEUnit      SEUnit
	KaraokeUnit KaraokeUnit
}

type TimeSig struct {
	XMLName xml.Name `xml:"timeSig"`
	PosMes  int      `xml:"posMes"`
	Nume    int      `xml:"nume"`
	Denomi  int      `xml:"denomi"`
}

type Tempo struct {
	XMLName xml.Name `xml:"tempo"`
	PosTick int      `xml:"posTick"`
	BPM     int      `xml:"bpm"`
}

type MasterTrack struct {
	XMLName    xml.Name `xml:"masterTrack"`
	SeqName    CData    `xml:"seqName"`
	Comment    CData    `xml:"comment"`
	Resolution int      `xml:"resolution"`
	PreMeasure int      `xml:"preMeasure"`
	TimeSig    TimeSig
	Tempo      Tempo
}

type StylePlugin struct {
	XMLName         xml.Name `xml:"stylePlugin"`
	StylePluginID   CData    `xml:"stylePluginID"`
	StylePluginName CData    `xml:"stylePluginName"`
	Version         CData    `xml:"version"`
}

type Attr struct {
	XMLName xml.Name `xml:"attr"`
	Value   int      `xml:",chardata"`
	ID      string   `xml:"id,attr"`
}

type Singer struct {
	XMLName xml.Name `xml:"singer"`
	PosTick int      `xml:"posTick"`
	BS      int      `xml:"vBS"`
	PC      int      `xml:"vPC"`
}

type Note struct {
	XMLName   xml.Name `xml:"note"`
	PosTick   int      `xml:"posTick"`
	DurTick   int      `xml:"durTick"`
	NoteNum   int      `xml:"noteNum"`
	Velocity  int      `xml:"velocity"`
	Lyric     CData    `xml:"lyric"`
	Phnms     CData    `xml:"phnms"`
	NoteStyle []Attr   `xml:"noteStyle>attr"`
}

type MCtrl struct {
	XMLName xml.Name `xml:"mCtrl"`
	PosTick int      `xml:"posTick"`
	Attr    []Attr   `xml:"attr"`
}

type MusicalPart struct {
	XMLName     xml.Name `xml:"musicalPart"`
	PosTick     int      `xml:"posTick"`
	PlayTime    int      `xml:"playTime"`
	PartName    CData    `xml:"partName"`
	Comment     CData    `xml:"comment"`
	StylePlugin StylePlugin
	PartStyle   []Attr `xml:"partStyle>attr"`
	Singer      Singer
	MCtrl       []MCtrl `xml:"mCtrl,omitempty"`
	Note        []Note  `xml:"note"`
}

type VSTrack struct {
	XMLName     xml.Name `xml:"vsTrack"`
	VSTrackNo   int      `xml:"vsTrackNo"`
	TrackName   CData    `xml:"trackName"`
	Comment     CData    `xml:"comment"`
	MusicalPart MusicalPart
}

type SETrack struct {
	XMLName xml.Name `xml:"seTrack"`
}

type KaraokeTrack struct {
	XMLName xml.Name `xml:"karaokeTrack"`
}

type AUX struct {
	XMLName xml.Name `xml:"aux"`
	AUXID   CData    `xml:"auxID"`
	Content CData    `xml:"content"`
}

type VSQ3 struct {
	XMLName        xml.Name `xml:"vsq3"`
	XMLNS          string   `xml:"xmlns,attr"`
	XSI            string   `xml:"xmlns:xsi,attr"`
	SchemaLocation string   `xml:"xsi:schemaLocation,attr"`
	Vender         CData    `xml:"vender"`
	Version        CData    `xml:"version"`
	VoiceTable     VoiceTable
	Mixer          Mixer
	MasterTrack    MasterTrack
	VSTrack        VSTrack
	SETrack        SETrack
	KaraokeTrack   KaraokeTrack
	AUX            AUX

	noteCount int `xml:"-"`
}

func Load(filename string) (*VSQ3, error) {
	x, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var result VSQ3
	if err := xml.Unmarshal(x, &result); err != nil {
		return nil, err
	}
	result.normalize("")
	return &result, nil
}

func New(singer string, resolution int, bpm float64) *VSQ3 {
	var vsq3 VSQ3
	vsq3.normalize(singer)
	vsq3.Mixer.VSUnit.SendLevel = -898
	vsq3.Mixer.VSUnit.Pan = 64
	vsq3.Mixer.SEUnit.SendLevel = -898
	vsq3.Mixer.SEUnit.Pan = 64
	vsq3.Mixer.KaraokeUnit.Vol = -129
	vsq3.MasterTrack.Resolution = resolution
	vsq3.MasterTrack.PreMeasure = 4
	vsq3.MasterTrack.TimeSig.Nume = 4
	vsq3.MasterTrack.TimeSig.Denomi = 4
	vsq3.MasterTrack.Tempo.BPM = int(math.Round(bpm * 100))
	vsq3.VSTrack.MusicalPart.PosTick = 7680
	vsq3.VSTrack.MusicalPart.PlayTime = 614400 // ?
	vsq3.VSTrack.MusicalPart.PartStyle = []Attr{
		{ID: "accent", Value: 50},
		{ID: "bendDep", Value: 8},
		{ID: "bendLen", Value: 0},
		{ID: "decay", Value: 50},
		{ID: "fallPort", Value: 0},
		{ID: "opening", Value: 127},
		{ID: "risePort", Value: 0},
	}

	return &vsq3
}

type singerDef struct {
	compID string
	bs     int
}

var DefaultSinger = "Yukari_Onn"

// Pull Request募集中
var singerDefs = map[string]singerDef{
	"CUL":                {compID: "BCBG86S4FSYMTCBK", bs: 0},
	"DEX":                {compID: "BEPP62G3DDXLRECA", bs: 1},
	"IA":                 {compID: "BLRGDDR4M3WM2LC6", bs: 0},
	"Iroha(V2)":          {compID: "BMKN7HT9EWTTSMCL", bs: 0},
	"KAITO_V3_English":   {compID: "BNGW7FG7E5TRSNC3", bs: 1},
	"KAITO_V3_Soft":      {compID: "BKGKCC96L2TPZKAC", bs: 0},
	"KAITO_V3_Straight":  {compID: "BDPEA722HT3KXDC4", bs: 0},
	"KAITO_V3_Whisper":   {compID: "BDHEB7W2KTWKYDC5", bs: 0},
	"LEN_V4X_Cold":       {compID: "BMGD88HZFLTHTMC7", bs: 0},
	"LEN_V4X_Power_EVEC": {compID: "BKPLC6S7LH3RZKC8", bs: 0},
	"LEN_V4X_Serious":    {compID: "BKFFF663PHSL4KB6", bs: 0},
	"LEN_V4_English":     {compID: "BMFX98L8GLSSWMD3", bs: 1},
	"Len_ACT2(V2)":       {compID: "BMLBDHXXMWYF2MBE", bs: 0},
	"Luka_ENG(V2)":       {compID: "BHLNEE62NRYK3HD2", bs: 1},
	"Luka_JPN(V2)":       {compID: "BCMDC9MZLKZHZCB4", bs: 0},
	"Miku(V2)":           {compID: "BHHN4EF9BRWTNHAB", bs: 0},
	"RIN_V4X_Power_EVEC": {compID: "BKKP765AEHXWSKDB", bs: 0},
	"RIN_V4X_Sweet":      {compID: "BLECA76YHKRGXLB7", bs: 0},
	"RIN_V4X_Warm":       {compID: "BDHEBZG2KCWKYDC5", bs: 0},
	"RIN_V4_English":     {compID: "BXENFF42PWRK4XE7", bs: 1},
	"Rin_ACT2(V2)":       {compID: "BEKF6B63DMXLRECA", bs: 0},
	"VY1V3":              {compID: "BDRE87E2FTTKTDBA", bs: 0},
	"VY2V3":              {compID: "BCXDC6CZLSZHZCB4", bs: 0},
	"VY2V3_falsetto":     {compID: "BDSEB7L2KTWKYDC5", bs: 0},
	"Yukari":             {compID: "BMGK9EC6G4RPWMB3", bs: 0},
	"Yukari_Jun":         {compID: "BDKCEZEYNCTG3DBF", bs: 0},
	"Yukari_Lin":         {compID: "BKLM76B8EHWSSKBB", bs: 0},
	"Yukari_Onn":         {compID: "BNRCB9XYKM2GYNCE", bs: 0},
}

func Singers() []string {
	result := []string{}
	for s, d := range singerDefs {
		if d.bs != 0 {
			// TODO: 英語DB対応
			continue
		}
		result = append(result, s)
	}
	sort.Strings(result)
	return result
}

func IsValidSinger(singer string) bool {
	d, ok := singerDefs[singer]
	if !ok {
		return false
	}
	// TODO: 英語DB対応
	return d.bs == 0
}

func (vsq3 *VSQ3) isEnglish() bool {
	return vsq3.VoiceTable.Voice.BS == 1
}

func (vsq3 *VSQ3) normalize(singer string) {
	if vsq3.XMLNS == "" {
		vsq3.XMLNS = "http://www.yamaha.co.jp/vocaloid/schema/vsq3/"
	}
	if vsq3.XSI == "" {
		vsq3.XSI = "http://www.w3.org/2001/XMLSchema-instance"
	}
	if vsq3.SchemaLocation == "" {
		vsq3.SchemaLocation = "http://www.yamaha.co.jp/vocaloid/schema/vsq3/vsq3.xsd"
	}
	if vsq3.Vender.Data == "" {
		vsq3.Vender.Data = "Yamaha corporation"
	}
	if vsq3.Version.Data == "" {
		vsq3.Version.Data = "3.0.0.11"
	}
	if _, ok := singerDefs[singer]; ok {
		vsq3.VoiceTable.Voice.BS = singerDefs[singer].bs
		vsq3.VoiceTable.Voice.CompID.Data = singerDefs[singer].compID
		vsq3.VoiceTable.Voice.VoiceName.Data = singer
	}
	if vsq3.MasterTrack.SeqName.Data == "" {
		vsq3.MasterTrack.SeqName.Data = "none"
	}
	if vsq3.MasterTrack.Comment.Data == "" {
		vsq3.MasterTrack.Comment.Data = "none"
	}
	if vsq3.VSTrack.TrackName.Data == "" {
		vsq3.VSTrack.TrackName.Data = "Track"
	}
	if vsq3.VSTrack.Comment.Data == "" {
		vsq3.VSTrack.Comment.Data = "Track"
	}
	if vsq3.VSTrack.MusicalPart.PartName.Data == "" {
		vsq3.VSTrack.MusicalPart.PartName.Data = "NewPart"
	}
	if vsq3.VSTrack.MusicalPart.Comment.Data == "" {
		vsq3.VSTrack.MusicalPart.Comment.Data = "New Musical Part"
	}
	if vsq3.VSTrack.MusicalPart.StylePlugin.StylePluginID.Data == "" {
		vsq3.VSTrack.MusicalPart.StylePlugin.StylePluginID.Data = "ACA9C502-A04B-42b5-B2EB-5CEA36D16FCE"
	}
	if vsq3.VSTrack.MusicalPart.StylePlugin.StylePluginName.Data == "" {
		vsq3.VSTrack.MusicalPart.StylePlugin.StylePluginName.Data = "VOCALOID2 Compatible Style"
	}
	if vsq3.VSTrack.MusicalPart.StylePlugin.Version.Data == "" {
		vsq3.VSTrack.MusicalPart.StylePlugin.Version.Data = "3.0.0.1"
	}
	if vsq3.AUX.AUXID.Data == "" {
		vsq3.AUX.AUXID.Data = "AUX_VST_HOST_CHUNK_INFO"
	}
	if vsq3.AUX.Content.Data == "" {
		vsq3.AUX.Content.Data = "VlNDSwAAAAADAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	}
}

var phonemes = map[string]string{
	"あ": "a", "い": "i", "う": "M", "え": "e", "お": "o",
	"か": "k a", "き": "k' i", "く": "k M", "け": "k e", "こ": "k o",
	"さ": "s a", "し": "S i", "す": "s M", "せ": "s e", "そ": "s o",
	"た": "t a", "ち": "tS i", "つ": "ts M", "て": "t e", "と": "t o",
	"な": "n a", "に": "J i", "ぬ": "n M", "ね": "n e", "の": "n o",
	"は": "h a", "ひ": "C i", "ふ": "p\\ M", "へ": "h e", "ほ": "h o",
	"ま": "m a", "み": "m i", "む": "m M", "め": "m e", "も": "m o",
	"ら": "4 a", "り": "4' i", "る": "4 M", "れ": "4 e", "ろ": "4 o",
	"が": "g a", "ぎ": "g' i", "ぐ": "g M", "げ": "g e", "ご": "g o",
	"ざ": "dz a", "じ": "dZ i", "ず": "dz M", "ぜ": "dz e", "ぞ": "dz o",
	"だ": "d a", "ぢ": "dZ i", "づ": "dz M", "で": "d e", "ど": "d o",
	"ば": "b a", "び": "b' i", "ぶ": "b M", "べ": "b e", "ぼ": "b o",
	"ぱ": "p a", "ぴ": "p' i", "ぷ": "p M", "ぺ": "p e", "ぽ": "p o",
	"や": "j a", "ゆ": "j M", "よ": "j o",
	"わ": "w a", "ゐ": "w i", "ゑ": "w e", "を": "o", "ん": "N\\",
	"ふぁ": `p\ a`, "つぁ": "ts a",
	"うぃ": "w i", "すぃ": "s i", "ずぃ": "dz i", "つぃ": "ts i", "てぃ": "t' i",
	"でぃ": "d' i", "ふぃ": `p\' i`,
	"とぅ": "t M", "どぅ": "d M",
	"いぇ": "j e", "うぇ": "w e", "きぇ": "k' e", "しぇ": "S e", "ちぇ": "tS e",
	"つぇ": "ts e", "てぇ": "t' e", "にぇ": "J e", "ひぇ": "C e", "みぇ": "m' e",
	"りぇ": "4' e", "ぎぇ": "g' e", "じぇ": "dZ e", "でぇ": "d' e", "びぇ": "b' e",
	"ぴぇ": "p' e", "ふぇ": `p\ e`,
	"うぉ": "w o", "つぉ": "ts o", "ふぉ": `p\ o`,
	"きゃ": "k' a", "しゃ": "S a", "ちゃ": "tS a", "てゃ": "t' a", "にゃ": "J a",
	"ひゃ": "C a", "みゃ": "m' a", "りゃ": "4' a", "ぎゃ": "N' a", "じゃ": "dZ a",
	"でゃ": "d' a", "びゃ": "b' a", "ぴゃ": "p' a", "ふゃ": `p\' a`,
	"きゅ": "k' M", "しゅ": "S M", "ちゅ": "tS M", "てゅ": "t' M", "にゅ": "J M",
	"ひゅ": "C M", "みゅ": "m' M", "りゅ": "4' M", "ぎゅ": "g' M", "じゅ": "dZ M",
	"でゅ": "d' M", "びゅ": "b' M", "ぴゅ": "p' M", "ふゅ": `p\' M`,
	"きょ": "k' o", "しょ": "S o", "ちょ": "tS o", "てょ": "t' o", "にょ": "J o",
	"ひょ": "C o", "みょ": "m' o", "りょ": "4' o", "ぎょ": "N' o", "じょ": "dZ o",
	"でょ": "d' o", "びょ": "b' o", "ぴょ": "p' o",
}

func (vsq3 *VSQ3) AddNote(velocity, beginTick, endTick, note int, lyrics, phnms string) {
	phnmsLock := 1
	if phnms == "" {
		if p, ok := phonemes[lyrics]; ok {
			phnms = p
		} else {
			phnmsLock = 0
			phnms = "4 a"
		}
	}
	vsq3.LimitLastNote(beginTick)
	vsq3.VSTrack.MusicalPart.Note = append(vsq3.VSTrack.MusicalPart.Note, Note{
		PosTick:  beginTick,
		DurTick:  endTick - beginTick,
		NoteNum:  note,
		Velocity: velocity,
		Lyric:    CData{Data: lyrics},
		Phnms:    CData{Data: phnms, Lock: phnmsLock},
		NoteStyle: []Attr{
			{ID: "accent", Value: 50},
			{ID: "bendDep", Value: 8},
			{ID: "bendLen", Value: 0},
			{ID: "decay", Value: 50},
			{ID: "fallPort", Value: 0},
			{ID: "opening", Value: 127},
			{ID: "risePort", Value: 0},
			{ID: "vibLen", Value: 0},
			{ID: "vibType", Value: 0},
		},
	})
	vsq3.noteCount++
}

func (vsq3 *VSQ3) ExtendLastNote(toTick, ifAfterTick int) bool {
	part := vsq3.VSTrack.MusicalPart
	n := len(part.Note)
	if n < 1 {
		return false
	}
	tail := part.Note[n-1].PosTick + part.Note[n-1].DurTick
	if tail < ifAfterTick {
		return false
	}
	part.Note[n-1].DurTick = toTick - part.Note[n-1].PosTick
	return true
}

func (vsq3 *VSQ3) LimitLastNote(toTick int) bool {
	part := vsq3.VSTrack.MusicalPart
	n := len(part.Note)
	if n < 1 {
		return false
	}
	l := toTick - part.Note[n-1].PosTick
	if part.Note[n-1].DurTick < l {
		return false
	}
	part.Note[n-1].DurTick = l
	return true
}

func (vsq3 *VSQ3) NoteCount() int {
	return vsq3.noteCount
}

func (vsq3 *VSQ3) AddMCtrl(tick int, id string, value int) {
	vsq3.VSTrack.MusicalPart.MCtrl = append(vsq3.VSTrack.MusicalPart.MCtrl, MCtrl{
		PosTick: tick,
		Attr: []Attr{{
			ID:    id,
			Value: value,
		}},
	})
}

func (vsq3 *VSQ3) Bytes() []byte {
	result, _ := xml.MarshalIndent(vsq3, "", "    ")
	return append([]byte(xml.Header), result...)
}

func (vsq3 *VSQ3) String() string {
	return string(vsq3.Bytes())
}
