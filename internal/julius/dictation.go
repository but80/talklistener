package julius

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/pkg/errors"
)

type dictationKit struct {
	kit      string
	url      string
	ext      string
	archiver string
	opt      string
	filename string
}

var DictationModelNames = []string{
	"std-gmm",
	"std-dnn",
	"ssr-dnn",
}

var configs = map[string]dictationKit{
	"std-gmm": {
		kit:      "dictation-kit-v4.3.1-osx",
		url:      "https://ja.osdn.net/frs/redir.php?m=jaist&f=julius%2F60416%2Fdictation-kit-v4.3.1-osx.tgz",
		ext:      "tgz",
		archiver: "tar xvzf",
		opt:      "-C",
		filename: "am-gmm.jconf",
	},
	"std-dnn": {
		kit:      "dictation-kit-v4.3.1-osx",
		url:      "https://ja.osdn.net/frs/redir.php?m=jaist&f=julius%2F60416%2Fdictation-kit-v4.3.1-osx.tgz",
		ext:      "tgz",
		archiver: "tar xvzf",
		opt:      "-C",
		filename: "am-dnn.jconf",
	},
	"ssr-dnn": {
		kit:      "ssr-kit-v4.4.2.1a",
		url:      "https://ja.osdn.net/frs/redir.php?m=iij&f=julius%2F68910%2Fssr-kit-v4.4.2.1a.zip",
		ext:      "zip",
		archiver: "unzip",
		opt:      "-dnnconf",
		filename: "main.dnnconf",
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

func Dictate(wavfile, model string) (*Result, error) {
	log.Println("発話内容を推定中...")

	u, err := user.Current()
	if err != nil {
		return nil, errors.Wrap(err, "ホームディレクトリを特定できません")
	}
	datadir := filepath.Join(u.HomeDir, ".talklistener")
	conf, ok := configs[model]
	if !ok {
		return nil, fmt.Errorf("ディクテーションモデル %s は定義されていません", model)
	}
	kitdir := filepath.Join(datadir, conf.kit)
	kit := filepath.Join(kitdir, conf.filename)
	if _, err := os.Stat(kit); err != nil {
		return nil, fmt.Errorf(downloadMsg[1:],
			model, conf.url, kitdir,
			conf.kit, conf.ext, conf.url,
			conf.archiver, conf.kit, conf.ext,
		)
	}
	argv := []string{
		"julius",
		"-C", filepath.Join(kitdir, "main.jconf"),
		conf.opt, kit,
		"-palign",
		"-input", "file",
	}
	return run(argv, wavfile)
}
