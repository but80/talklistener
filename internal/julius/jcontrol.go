package julius

import (
	"bufio"
	"encoding/xml"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"golang.org/x/xerrors"
)

type WHypo struct {
	XMLName xml.Name `xml:"WHYPO"`
	Word    string   `xml:"WORD,attr"`
	ClassID string   `xml:"CLASSID,attr"`
	Phone   string   `xml:"PHONE,attr"`
	CM      float64  `xml:"CM,attr"`
}

type SHypo struct {
	XMLName xml.Name `xml:"SHYPO"`
	WHypo   []*WHypo `xml:"WHYPO"`
	Rank    int      `xml:"RANK,attr"`
	Score   float64  `xml:"SCORE,attr"`
}

type RecogOut struct {
	XMLName xml.Name `xml:"RECOGOUT"`
	SHypo   *SHypo   `xml:"SHYPO"`
}

func jcontrol() ([][]string, error) {
	retry := 80
	interval := time.Millisecond * 250

	var conn net.Conn
	var err error
	for i := 0; i < retry; i++ {
		conn, err = net.Dial("tcp", "localhost:10500")
		if err == nil {
			break
		}
		time.Sleep(interval)
	}
	if err != nil {
		return nil, xerrors.Errorf("juliusサーバに接続できません: %w", err)
	}
	defer conn.Close()
	dec := japanese.ShiftJIS.NewDecoder()
	trans := transform.NewReader(conn, dec)
	reader := bufio.NewReader(trans)

	log.Printf("debug: juliusサーバに接続しました")
	var xmlsrc []string
	var result [][]string
	for {
		line, err := reader.ReadString('\n')
		if xerrors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, xerrors.Errorf("juliusサーバからのデータ受信中にエラーが発生しました: %w", err)
		}
		line = strings.TrimRight(line, "\r\n")
		if line != "." {
			xmlsrc = append(xmlsrc, line)
			continue
		}
		var out RecogOut
		// log.Printf("debug: XML: %s", xmlsrc)
		if err := xml.Unmarshal([]byte(strings.Join(xmlsrc, "\n")), &out); err == nil && out.SHypo != nil {
			var words []string
			var phones []string
			for _, h := range out.SHypo.WHypo {
				words = append(words, h.Word)
				for _, p := range strings.Split(h.Phone, " ") {
					phones = append(phones, centerName(p))
				}
			}
			log.Printf("info: 発話内容: %s (%s)", strings.Join(words, ","), strings.Join(phones, ","))
			result = append(result, phones)
		}
		xmlsrc = nil
	}
	return result, nil
}
