package main

import (
	"os"
	"strings"

	"github.com/but80/talklistener/internal/generator"
	"github.com/but80/talklistener/internal/globalopt"
	"github.com/but80/talklistener/internal/julius"
	"github.com/comail/colog"
	"github.com/urfave/cli"

	// Go >= 1.10 required
	_ "github.com/theckman/goconstraint/go1.10/gte"
)

var version = "unknown"

const description = `
   - <音声ファイル> は .wav .aiff .mp3 等のフォーマットに対応しています。
     詳細は afconvert のヘルプを afconvert -hf にてお読みください。
   - イントネーションの抽出に「音声分析変換合成システム WORLD」
     https://github.com/mmorise/World を使用しています。
   - 発音タイミングの抽出に「大語彙連続音声認識エンジン Julius」
     https://github.com/julius-speech/julius
     および、以下の関連データを使用しています。
     - 音素セグメンテーションキット https://github.com/julius-speech/segmentation-kit
     - ディクテーションキット https://github.com/julius-speech/dictation-kit
`

func main() {
	app := cli.NewApp()
	app.Name = "talklistener"
	app.Version = version
	app.Usage = "話し声を録音したwavファイルと、その読みを記述したテキストの組み合わせから、Vocaloid3シーケンスを生成します"
	app.Description = strings.TrimSpace(description)
	app.Authors = []cli.Author{
		{
			Name:  "but80",
			Email: "mersenne.sister@gmail.com",
		},
	}
	app.HelpName = "talklistener"
	app.UsageText = "talklistener [オプション...] <音声ファイル>"
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "transpose, p",
			Usage: `出力VSQX内の全ノートの音高をずらします（単位：セント）`,
		},
		cli.BoolFlag{
			Name:  "redictate, R",
			Usage: "発話内容の再認識を行い、その結果をテキストファイルに上書き保存します",
		},
		cli.StringFlag{
			Name:  "dictation-model, d",
			Usage: "発話内容の認識に使用するモデル (" + strings.Join(julius.DictationModelNames, ", ") + ")",
			Value: "ssr-dnn",
		},
		cli.StringFlag{
			Name:  "out, o",
			Usage: `出力VSQXを指定した名前で保存します（省略時は "音声ファイル名.vsqx"）`,
		},
		cli.StringFlag{
			Name:  "text, t",
			Usage: `テキストファイルを指定した名前で保存・ロードします（省略時は "音声ファイル名.txt"）`,
		},
		cli.BoolFlag{
			Name:  "leave-obj, l",
			Usage: `中間オブジェクトを削除せず、ディレクトリ "音声ファイル名.tlo/" に保存します`,
		},
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "詳細を表示します",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "デバッグ情報を表示します",
		},
		cli.BoolFlag{
			Name:  "version",
			Usage: "バージョン番号を表示します",
		},
	}
	app.HideVersion = true

	app.Action = func(ctx *cli.Context) error {
		if ctx.Bool("version") {
			cli.ShowVersion(ctx)
			return nil
		}
		if ctx.NArg() < 1 {
			cli.ShowAppHelpAndExit(ctx, 1)
		}
		wavfile := ctx.Args()[0]
		txtfile := ctx.String("text")
		outfile := ctx.String("out")
		globalopt.Verbose = ctx.Bool("verbose")
		globalopt.Debug = ctx.Bool("debug")
		if globalopt.Debug {
			colog.SetMinLevel(colog.LDebug)
		} else if globalopt.Verbose {
			colog.SetMinLevel(colog.LInfo)
		} else {
			colog.SetMinLevel(colog.LWarning)
		}
		if err := generator.Generate(
			wavfile, txtfile, ctx.String("dictation-model"), outfile,
			ctx.Bool("redictate"), ctx.Bool("leave-obj"), ctx.Int("transpose"),
		); err != nil {
			return cli.NewExitError(err, 1)
		}
		return nil
	}

	colog.Register()
	app.Run(os.Args)
}
