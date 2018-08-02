package julius

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/but80/talklistener/internal/assets"
)

const (
	hmmDefs = "/cmodules/segmentation-kit/models/hmmdefs_monof_mix16_gid.binhmm" // monophone model
	// hmmDefs = "/cmodules/segmentation-kit/models/hmmdefs_ptm_gid.binhmm" // triphone model
)

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

func Segmentate(wavfile, wordsfile, objPrefix string) (*Result, error) {
	log.Println("発音タイミングを推定中...")

	words, err := wordsToDict(wordsfile, objPrefix+".dict")
	if err != nil {
		return nil, err
	}
	generateDFA(len(words), objPrefix+".dfa")

	hmmTmpName := objPrefix + ".binhmm"
	if err := saveAssetAsFile(hmmDefs, hmmTmpName); err != nil {
		return nil, err
	}

	argv := []string{
		"julius",
		"-h", hmmTmpName, // HMM definition
		"-dfa", objPrefix + ".dfa", // DFA grammar
		"-v", objPrefix + ".dict", // dictionary
		"-palign", // optionally output phoneme alignments
		"-input", "file",
	}
	return run(argv, wavfile)
}
