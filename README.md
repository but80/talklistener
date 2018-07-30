# TalkListener

話し声を録音したwavファイルと、その読みを記述したテキストの組み合わせから、Vocaloid3シーケンスを生成するCLIツールです。

- [デモ](https://twitter.com/bucchigiri/status/1023193037193719808)

## 必須環境

- macOS Sierra

## インストール手順

### macOS + Homebrew でのインストール

```bash
brew tap but80/tap
brew install but80/tap/talklistener
```

### Go 1.10 でのインストール（開発者向け）

```bash
go run mage.go install

# または

go run mage.go build
mv ./talklistener /usr/local/bin/
```

## 使用方法

```
USAGE: talklistener <input wav> <input text> [<output vsqx>]
```

- 入力音源ファイル `<input wav>` は、以下のスペックのwavファイルである必要があります。
  - サンプリング周波数 16,000 Hz
  - 量子化ビット数 16 bit
  - モノラル
- 入力テキストファイル `<input text>` には、認識させたい読みをひらがなで記述します。
  - 文字エンコーディングは Unicode です。
  - 間隔が開く箇所には ` sp ` と記述します（左右に半角スペースが必要です）
- `<output vsqx>` を省略すると標準出力に結果を出力します。
- 具体的な使用例は [examples/](./examples) を参考にしてください。

## Tips・注意事項

- 入力ファイルのフォーマットが合わない場合は、[SoX](http://brewformulas.org/Sox) を利用するとCLIで簡単にコンバートできます。<br>例: `sox input.aiff -r 16000 output.wav`
- 音声が長すぎるとエラーになる場合があります。その場合は、細切れにして別々に処理してください。
- 出力されるVSQXに設定されたシンガーは、現バージョンでは「結月ゆかり 穏」固定です。Vocaloidエディタで読み込み後、目的のシンガーに変更してください。

## ライセンス

- 本ソフトウェアは三条項BSDライセンスです。[LICENSE](./LICENSE) をお読みください。
- イントネーションの抽出に [音声分析変換合成システム WORLD](https://github.com/mmorise/World) を使用しています。WORLD の著作権およびライセンスについては [LICENSE-world.txt](./LICENSE-world.txt) をお読みください。
- 発音タイミングの抽出に [大語彙連続音声認識エンジン Julius](https://github.com/julius-speech/julius) および [segmentation-kit](https://github.com/julius-speech/segmentation-kit) に含まれる音響モデルを使用しています。Julius の著作権およびライセンスについては [LICENSE-julius.txt](./LICENSE-julius.txt) をお読みください。
- 本ソフトウェアを利用して制作したソフトウェアのソースコード および バイナリに付属するドキュメントには、本ソフトウェアのライセンス表記に加え、上記 WORLD, Julius の各ライセンス表記を併せて含める必要があります。
