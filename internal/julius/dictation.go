package julius

import (
	"fmt"
	"os/user"
	"path/filepath"

	"github.com/pkg/errors"
)

// ArchiveURL: "https://ja.osdn.net/frs/redir.php?m=jaist&f=julius%2F60416%2Fdictation-kit-v4.3.1-osx.tgz"

type dictationKit struct {
	kit      string
	opt      string
	filename string
}

var DictationKitNames = []string{
	"std-gmm",
	"std-dnn",
	"ssr-dnn",
}

var configs = map[string]dictationKit{
	"std-gmm": {
		kit:      "dictation-kit-v4.3.1-osx",
		opt:      "-C",
		filename: "am-gmm.jconf",
	},
	"std-dnn": {
		kit:      "dictation-kit-v4.3.1-osx",
		opt:      "-C",
		filename: "am-dnn.jconf",
	},
	"ssr-dnn": {
		kit:      "ssr-kit-v4.4.2.1a",
		opt:      "-dnnconf",
		filename: "main.dnnconf",
	},
}

func Dictate(wavfile, kitName string) (*Result, error) {
	u, err := user.Current()
	if err != nil {
		return nil, errors.Wrap(err, "ホームディレクトリを特定できません")
	}
	datadir := filepath.Join(u.HomeDir, ".talklistener")
	conf, ok := configs[kitName]
	if !ok {
		return nil, fmt.Errorf("ディクテーションキット %s は定義されていません", kitName)
	}
	kitdir := filepath.Join(datadir, conf.kit)
	argv := []string{
		"julius",
		"-C", filepath.Join(kitdir, "main.jconf"),
		conf.opt, filepath.Join(kitdir, conf.filename),
		"-palign",
		"-input", "file",
	}
	return run(argv, wavfile)
}
