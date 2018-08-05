# TalkListener

話し声を録音したwavファイルからVocaloid3シーケンスを生成するCLIツールです。

- [デモ](https://twitter.com/bucchigiri/status/1023193037193719808)

## 必須環境

- macOS Sierra 以降

## インストール手順

### 手作業でのインストール

1. [リリースページ](https://github.com/but80/talklistener/releases) から `talklistener_X.X.X_darwin_amd64.tar.gz` をダウンロード
2. アーカイブを展開
3. 展開されたディレクトリに含まれる `talklistener` を任意の場所に移動

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
sudo mv ./talklistener /usr/local/bin/
```

Julius, WORLD が依存しているライブラリが必要になりますので、不足している場合は事前にインストールしてください。

## 使用方法

```
NAME:
   talklistener - 話し声を録音したwavファイルからVocaloid3シーケンスを生成します

USAGE:
   talklistener [オプション...] <音声ファイル>

DESCRIPTION:
   - <音声ファイル> は .wav .aiff .mp3 等のフォーマットに対応しています。
     詳細は afconvert のヘルプを afconvert -hf にてお読みください。
   - イントネーションの抽出に「音声分析変換合成システム WORLD」
     https://github.com/mmorise/World を使用しています。
   - 発音タイミングの抽出に「大語彙連続音声認識エンジン Julius」
     https://github.com/julius-speech/julius
     および、以下の関連データを使用しています。
     - 音素セグメンテーションキット https://github.com/julius-speech/segmentation-kit
     - ディクテーションキット https://github.com/julius-speech/dictation-kit

   シンガー一覧:
     CUL, IA, Iroha(V2), KAITO_V3_Soft, KAITO_V3_Straight, KAITO_V3_Whisper,
     LEN_V4X_Cold, LEN_V4X_Power_EVEC, LEN_V4X_Serious, Len_ACT2(V2),
     Luka_JPN(V2), Miku(V2), RIN_V4X_Power_EVEC, RIN_V4X_Sweet, RIN_V4X_Warm,
     Rin_ACT2(V2), VY1V3, VY2V3, VY2V3_falsetto, Yukari, Yukari_Jun,
     Yukari_Lin, Yukari_Onn

AUTHOR:
   but80 <mersenne.sister@gmail.com>

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --singer value, -s value           シンガー (default: "Yukari_Onn")
   --transpose value, -t value        出力VSQX内の全ノートの音高をずらします（単位：セント） (default: 0)
   --split-consonant, -c              子音を母音とは別のノートに分割配置します
   --redictate, -R                    発話内容の再認識を行い、その結果をテキストファイルに上書き保存します
   --f0-cutoff value, -f value        基本周波数の変動にかけるLPFのカットオフ周波数 (0.5, 1.0, 1.5, 2.0, 2.5, 3.0) (default: "1.0")
   --dictation-model value, -d value  発話内容の認識に使用するモデル (std-gmm, std-dnn, ssr-dnn) (default: "ssr-dnn")
   --out value                        出力VSQXを指定した名前で保存します（省略時は "音声ファイル名.vsqx"）
   --text value                       テキストファイルを指定した名前で保存・ロードします（省略時は "音声ファイル名.txt"）
   --recache, -r                      キャッシュ "音声ファイル名.tlo/" を再作成します
   --quiet, -q                        進捗情報等の表示を抑制します
   --verbose, -v                      詳細を表示します
   --debug                            デバッグ情報を表示します
   --version                          バージョン番号を表示します
   --help, -h                         show help
```

### 入力音声ファイル

`<音声ファイル>` は `.wav` `.aiff` `.mp3` 等のフォーマットに対応しています。
詳細は afconvert（macOS標準付属）のヘルプを `afconvert -hf` にてお読みください。

### テキストファイル

デフォルトでは、`<音声ファイル>` の拡張子を `.txt` に置換した名前のテキストファイルを認識します。

例: 音声ファイル `hello.wav` に対応するテキストファイルは、同ディレクトリの `hello.txt`

- このテキストファイルが存在しないときは、音声ファイルから自動的に発話内容を認識し、同名のテキストファイルに保存します。
- このテキストファイルが存在するときは、その内容を「読み」として採用します。
- 自動認識された読みが期待通りでない場合は、テキストファイルの内容を変更してから再実行してください。

このテキストファイルは以下のように記述する必要があります。

- 文字エンコーディングは Unicode です
- 基本的には、読みをひらがなで記述しますが、子音のみの箇所は半角英字で発音記号を記述する必要があります
- 間隔が開く箇所には ` sp ` と記述します（左右に半角スペースが必要です）
- 助詞の「は」「へ」は `わ` `え` と記述する必要があります（`は` と記述すると `h a` と読まれてしまいます）

### 出力

デフォルトでは、`<音声ファイル>` の拡張子を `.vsqx` に置換した名前で生成シーケンスを保存します。

## 使用例

[examples/](./examples) を参考にしてください。

## 注意事項

- 音声が長すぎるとエラーになる場合があります。その場合は、細切れにして別々に処理してください。
- 出力されるVSQXに設定されたシンガーは、現バージョンでは「結月ゆかり 穏」固定です。Vocaloidエディタで読み込み後、目的のシンガーに変更してください。
- 入力音声ファイルは、内部的に以下のスペックのwavファイルに変換されます。これを超えるスペックのファイルを用意しても、高周波成分やパンは考慮されません。
  - サンプリング周波数 16,000 Hz
  - 量子化ビット数 16 bit
  - モノラル
- 入力音声ファイルを更新して再実行した際、キャッシュに残っている古い音声ファイルが参照されてしまう場合があります（入力音声ファイルのタイムスタンプが更新されていないと発生します）。このような場合は `-r` オプションを付けるか、または事前にキャッシュディレクトリ `音声ファイル.tlo/` を削除してから再実行してください。
- キャッシュディレクトリ `音声ファイル.tlo/` は、同一音声ファイルに対する再実行時の処理を軽減するために作成されています。出力VSQXファイルの内容を確認して問題がなければ、このキャッシュディレクトリは削除しても構いません。

## ライセンス

- 本ソフトウェアは三条項BSDライセンスです。[LICENSE](./LICENSE) をお読みください。
- イントネーションの抽出に [音声分析変換合成システム WORLD](https://github.com/mmorise/World) を使用しています。WORLD の著作権およびライセンスについては [LICENSE-world.txt](./LICENSE-world.txt) をお読みください。
- 発音タイミングの抽出に [大語彙連続音声認識エンジン Julius](https://github.com/julius-speech/julius) および [segmentation-kit](https://github.com/julius-speech/segmentation-kit) に含まれる音響モデルを使用しています。Julius の著作権およびライセンスについては [LICENSE-julius.txt](./LICENSE-julius.txt) をお読みください。
- 本ソフトウェアを利用して制作したソフトウェアのソースコード および バイナリに付属するドキュメントには、本ソフトウェアのライセンス表記に加え、上記 WORLD, Julius の各ライセンス表記を併せて含める必要があります。

## TODO

- やりたい
  - ディクテーションキットの自動ダウンロード
  - アプリケーションリンクへのDnDで変換
  - 無音部分のf0補間方法を改良
- やるかも
  - 抑揚を強調
  - 音量からDYNを生成
  - 非周期成分の比率からBREを生成
  - Windows対応
  - GUI
