# TalkListener

## 必須環境

- macOS Sierra
- Go 1.10

## ビルド手順

```bash
go run mage.go build
./talklistener
```

## 使用方法

```
USAGE: talklistener <input wav> <input text> [<output vsqx>]
```

- 入力音源 `<input wav>` は、以下のスペックのwavファイルである必要があります。
  - サンプリング周波数 16,000 Hz
  - 量子化ビット数 16 bit
  - モノラル
- 入力テキスト `<input text>` には、認識させたい読みをひらがなで記述します。
  - 文字エンコーディングは Unicode です。
  - 間隔が開く箇所には ` sp ` と記述します（左右に半角スペースが必要です）
- `<output vsqx>` を省略すると標準出力に結果を出力します。
- 具体的な使用例は [examples/](./examples) を参考にしてください。

## Tips・注意事項

- 入力ファイルのフォーマットが合わない場合は、[SoX](http://brewformulas.org/Sox) を利用するとCLIで簡単にコンバートできます。<br>例: `sox input.aiff -r 16000 output.wav`
- 音声が長すぎるとエラーになる場合があります。その場合は、細切れにして別々に処理してください。

## ライセンス

- 本ソフトウェアは三条項BSDライセンスです。[LICENSE](./LICENSE) をお読みください。
- 本ソフトウェアを利用して制作したソフトウェアのソースコード および バイナリに付属するドキュメントには、本ソフトウェアのライセンス表記に加え、以下の WORLD, Julius の各ライセンス表記を含める必要があります。
  - イントネーションの抽出に [音声分析変換合成システム WORLD](https://github.com/mmorise/World) を使用しています。WORLD の著作権およびライセンスについては [LICENSE-world.txt](./LICENSE-world.txt) をお読みください。
  - 発音タイミングの抽出に [大語彙連続音声認識エンジン Julius](https://github.com/julius-speech/julius) および [segmentation-kit](https://github.com/julius-speech/segmentation-kit) に含まれる音響モデルを使用しています。Julius の著作権およびライセンスについては [LICENSE-julius.txt](./LICENSE-julius.txt) をお読みください。
