package julius

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/xerrors"
)

type dictationKit struct {
	kit      string
	url      string
	ext      string
	archiver string
	bin      map[string]string
	opts     []string
}

var DictationModelNames = []string{
	"dictation",
	"ssr",
	"lsr",
}

var configs = map[string]dictationKit{
	//
	"dictation": {
		kit:      "dictation-kit-v4.5",
		url:      "https://osdn.net/frs/redir.php?m=ymu&f=julius%2F71011%2Fdictation-kit-4.5.zip",
		ext:      "zip",
		archiver: "unzip",
		bin: map[string]string{
			"darwin":  "bin/osx/julius",
			"linux":   "bin/linux/julius",
			"windows": "bin/windows/julius.exe",
		},
		opts: []string{
			"-C", "./main.jconf",
			"-C", "./am-dnn.jconf",
			"-demo",
			"-dnnconf", "./julius.dnnconf",
		},
	},
	"ssr": {
		kit:      "ssr-kit-v4.5",
		url:      "https://osdn.net/frs/redir.php?m=iij&f=julius%2F71011%2Fssr-kit-v4.5.zip",
		ext:      "zip",
		archiver: "unzip",
		bin: map[string]string{
			"darwin":  "bin/osx/julius",
			"linux":   "bin/linux/julius",
			"windows": "bin/windows/julius.exe",
		},
		opts: []string{
			"-C", "./main.jconf",
			"-dnnconf", "./main.dnnconf",
		},
	},
	"lsr": {
		kit:      "lsr-kit-v4.5",
		url:      "https://osdn.net/frs/redir.php?m=iij&f=julius%2F71011%2Flsr-kit-v4.5.zip",
		ext:      "zip",
		archiver: "unzip",
		bin: map[string]string{
			"darwin":  "bin/osx/julius",
			"linux":   "bin/linux/julius",
			"windows": "bin/windows/julius.exe",
		},
		opts: []string{
			"-C", "./main.jconf",
			"-dnnconf", "./main.dnnconf",
		},
	},
}

const downloadMsg = `
ディクテーションモデル %s が見つかりません。
%s
からディクテーションキットのアーカイブをダウンロードし、展開したディレクトリを
%s
に配置して下さい。以下のコマンドで実行できます。

mkdir -p ~/.talklistener; cd ~/.talklistener; curl -vLo %s.%s '%s' && %s %s.%s
`

func unzip(src io.Reader, dest string) error {
	dest = filepath.Dir(dest)

	buf := bytes.NewBuffer([]byte{})
	if _, err := io.Copy(buf, src); err != nil {
		return err
	}
	b := buf.Bytes()
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return err
	}

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func downloadKit(url, dest string) bool {
	log.Print("info: ディクテーションキットをダウンロード中...")
	resp, err := http.Get(url)
	if err != nil {
		log.Print("warn: ディクテーションキットのダウンロードに失敗しました: %s", err.Error())
		return false
	}
	defer resp.Body.Close()

	log.Print("info: ディクテーションキットのアーカイブを解凍中...")
	if err := unzip(resp.Body, dest); err != nil {
		log.Print("warn: ディクテーションキットのアーカイブ解凍に失敗しました: %s", err.Error())
		return false
	}

	return true
}

func Dictate(wavfile, model string) (*Result, error) {
	log.Print("info: 発話内容を推定中...")

	u, err := user.Current()
	if err != nil {
		return nil, xerrors.Errorf("ホームディレクトリを特定できません: %w", err)
	}
	datadir := filepath.Join(u.HomeDir, ".talklistener")
	conf, ok := configs[model]
	if !ok {
		return nil, fmt.Errorf("ディクテーションモデル %s は定義されていません", model)
	}
	kitdir := filepath.Join(datadir, conf.kit)
	if _, err := os.Stat(kitdir); err != nil {
		if ok := downloadKit(conf.url, kitdir); !ok {
			return nil, fmt.Errorf(downloadMsg[1:],
				model, conf.url, kitdir,
				conf.kit, conf.ext, conf.url,
				conf.archiver, conf.kit, conf.ext,
			)
		}
	}
	bin, ok := conf.bin[runtime.GOOS]
	if !ok {
		return nil, fmt.Errorf("このプラットフォーム %s ではディクテーションモデル %s を利用できません", runtime.GOOS, model)
	}
	bin = filepath.Join(kitdir, bin)

	argv := []string{bin}
	argv = append(argv, conf.opts...)
	argv = append(argv,
		"-palign",
		"-input", "file",
	)

	// 引数中の相対パスを絶対パスに変換
	for i := range argv {
		v := argv[i]
		if strings.HasPrefix(v, "./") {
			argv[i] = filepath.Join(kitdir, v[2:])
		}
	}
	log.Printf("debug: running julius: %+v", argv)

	return run(argv, wavfile)
}
