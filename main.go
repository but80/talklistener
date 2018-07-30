package main

import (
	"os"

	_ "github.com/theckman/goconstraint/go1.10/gte"
	"github.com/urfave/cli"
)

var version string

const description = `
   - イントネーションの抽出に「音声分析変換合成システム WORLD」
     https://github.com/mmorise/World を使用しています
   - 発音タイミングの抽出に「大語彙連続音声認識エンジン Julius」
     https://github.com/julius-speech/julius
     および segmentation-kit https://github.com/julius-speech/segmentation-kit
     に含まれる音響モデルを使用しています
`

func main() {
	app := cli.NewApp()
	app.Name = "talklistener"
	app.Version = version
	app.Usage = "話し声を録音したwavファイルと、その読みを記述したテキストの組み合わせから、Vocaloid3シーケンスを生成します"
	app.Authors = []cli.Author{
		{
			Name:  "but80",
			Email: "mersenne.sister@gmail.com",
		},
	}
	app.HelpName = "talklistener"
	app.UsageText = "talklistener <音声ファイル.wav> <テキストファイル.txt> [<出力ファイル.vsqx>]"
	app.Description = description[4:]

	app.Action = func(ctx *cli.Context) error {
		if ctx.NArg() < 2 {
			cli.ShowAppHelpAndExit(ctx, 1)
		}
		wavfile := ctx.Args()[0]
		txtfile := ctx.Args()[1]
		outfile := ""
		if 3 <= ctx.NArg() {
			outfile = ctx.Args()[2]
		}
		if err := generate(wavfile, txtfile, outfile); err != nil {
			return cli.NewExitError(err, 1)
		}
		return nil
	}

	app.Run(os.Args)
}
