package main

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/but80/talklistener/internal/julius"
	"github.com/pkg/errors"
)

// ArchiveURL: "https://ja.osdn.net/frs/redir.php?m=jaist&f=julius%2F60416%2Fdictation-kit-v4.3.1-osx.tgz"

type config struct {
	kit      string
	opt      string
	filename string
}

var ConfigNames = []string{
	"std-gmm",
	"std-dnn",
	"ssr-dnn",
}

var configs = map[string]config{
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

func Dictate(wavfile, filename string) (*julius.Result, error) {
	u, err := user.Current()
	if err != nil {
		return nil, errors.Wrap(err, "ホームディレクトリを特定できません")
	}
	datadir := filepath.Join(u.HomeDir, ".talklistener")
	conf := configs[filename]
	kitdir := filepath.Join(datadir, conf.kit)
	argv := []string{
		"julius",
		"-C", filepath.Join(kitdir, "main.jconf"),
		conf.opt, filepath.Join(kitdir, conf.filename),
		"-palign",
		"-input", "file",
	}
	return julius.Run(argv, wavfile)
}

func main() {
	result, err := Dictate("../../input/zassou1.wav", "ssr-dnn")
	if err != nil {
		panic(err)
	}
	for _, dic := range result.Dictation {
		fmt.Printf("%s\n", strings.Join(dic, " "))
	}
	fmt.Printf("%#v\n", result.Segments)
}
