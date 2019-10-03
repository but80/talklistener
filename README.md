# TalkListener

話し声を録音した音声ファイルからVocaloid3シーケンスを生成するCLIツールです。

- [デモ](https://twitter.com/bucchigiri/status/1023193037193719808)

## 必須環境

- 以下のいずれかのOS
  - macOS Sierra 以降
  - Linux
  - Windows 10

## インストール手順

### 手作業でのインストール

1. [リリースページ](https://github.com/but80/talklistener/releases) から `talklistener_*.tar.gz` をダウンロード
2. アーカイブを展開
3. 展開されたディレクトリに含まれる `talklistener-cli` を任意の場所に移動

### macOS + Homebrew でのインストール

```bash
brew tap but80/tap
brew install but80/tap/talklistener
```

### Go 1.12 でのインストール（開発者向け）

後述のビルド手順に従ってください。

## 使用方法

```
NAME:
   talklistener - 話し声を録音した音声ファイルからVocaloid3シーケンスを生成します

USAGE:
   talklistener [オプション...] <音声ファイル>

DESCRIPTION:
   - <音声ファイル> は .wav .aiff .flac 等のフォーマットに対応しています。
     詳細はlibsndfileのプロジェクトページ http://www.mega-nerd.com/libsndfile/ をお読みください。
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
   --f0-cutoff value, -f value        基本周波数の変動にかけるLPFのカットオフ周波数 (0.5, 1.0, 1.5, 2.0, 2.5, 3.0) (default: "1.5")
   --f0-delay value, -d value         発音タイミングに対する基本周波数の変動を遅らせます（単位：ミリ秒） (default: 0)
   --dictation-model value, -m value  発話内容の認識に使用するモデル (dictation, ssr, lsr) (default: "ssr")
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

`<音声ファイル>` は `.wav` `.aiff` `.flac` 等のフォーマットに対応しています。
詳細は [libsndfileのプロジェクトページ](http://www.mega-nerd.com/libsndfile/) をお読みください。

細かいオプションを指定する必要がなければ、実行ファイルへのショートカットに対して音声ファイルをドラッグアンドドロップするだけでも処理できます。

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

出力ファイルを **Vocaloid Editor 3 で開くとエラーとなる** 事象が確認されています。
**Piapro Studio でのインポートをおすすめします** 。

## 使用例

[examples/](./examples) を参考にしてください。

## 注意事項

- 音声が長すぎるとエラーになる場合があります。その場合は、細切れにして別々に処理してください。
- 選択可能なシンガーは、本ツールの作成者が compID（ライブラリを特定するためのID）を知り得たもののみを列挙しています。
  - 必要な選択肢がない場合は、Vocaloidエディタで読み込み後、目的のシンガーに変更してください。
  - 追加して欲しいシンガーのトラック（シーケンスは空でもOK）を含むVSQXをお送りいただければ、compID を抽出して選択肢に追加します。
- 入力音声ファイルは、内部的に以下のスペックのwavファイルに変換されます。これを超えるスペックのファイルを用意しても、高周波成分やパンは考慮されません。
  - サンプリング周波数 16,000 Hz
  - 量子化ビット数 16 bit
  - モノラル
- 入力音声ファイルを更新して再実行した際、キャッシュに残っている古い音声ファイルが参照されてしまう場合があります（入力音声ファイルのタイムスタンプが更新されていないと発生します）。このような場合は `-r` オプションを付けるか、または事前にキャッシュディレクトリ `音声ファイル.tlo/` を削除してから再実行してください。
- キャッシュディレクトリ `音声ファイル.tlo/` は、同一音声ファイルに対する再実行時の処理を軽減するために作成されています。出力VSQXファイルの内容を確認して問題がなければ、このキャッシュディレクトリは削除しても構いません。

## ビルド手順

### macOS / Linux

以下のものが必要になりますので、不足している場合は事前にインストールしてください。

- Go 1.12
- libsndfile

```bash
git clone https://github.com/but80/talklistener.git
cd talklistener
go run mage.go buildCli
```

### Windows

1. MSYS2 + MinGW64 環境をセットアップ
2. 環境変数 `MSYSTEM=MINGW64` で MSYS2 起動
3. ```bash
   # Install packages
   pacman -S git make mingw-w64-x86_64-{autoconf,gcc,pkg-config,go,portaudio,libsndfile}
   
   # Setup golang
   mkdir ~/go
   export GOPATH=$HOME/go
   export GOROOT=/mingw64/lib/go
   
   # Build our project
   git clone https://github.com/but80/talklistener.git
   cd talklistener
   (
     cd cmodules/julius/libsent &&
     ./configure --with-mictype=portaudio &&
     make && make install
   )
   (
     cd cmodules/julius/libjulius &&
     ./configure && make && make install
   )
   ( cd cmodules/world && make )
   go build ./cmd/talklistener-cli
   
   # Make distribution package
   mkdir dist
   cp talklistener-cli.exe dist/
   cp /mingw64/bin/{libgomp-1,libportaudio-2,libstdc++-6,zlib1}.dll dist/
   ```

## ライセンス

- 本ソフトウェアは三条項BSDライセンスです。[LICENSE](./LICENSE) をお読みください。
- イントネーションの抽出に [音声分析変換合成システム WORLD](https://github.com/mmorise/World) を使用しています。WORLD の著作権およびライセンスについては [LICENSE-world.txt](./LICENSE-world.txt) をお読みください。
- 発音タイミングの抽出に [大語彙連続音声認識エンジン Julius](https://github.com/julius-speech/julius) および [segmentation-kit](https://github.com/julius-speech/segmentation-kit) に含まれる音響モデルを使用しています。Julius の著作権およびライセンスについては [LICENSE-julius.txt](./LICENSE-julius.txt) をお読みください。
- 本ソフトウェアを利用して制作したソフトウェアのソースコード および バイナリに付属するドキュメントには、本ソフトウェアのライセンス表記に加え、上記 WORLD, Julius の各ライセンス表記を併せて含める必要があります。

## TODO

- やりたい
  - 無音部分のf0補間方法を改良
- やるかも
  - 抑揚を強調
  - 音量からDYNを生成
  - 非周期成分の比率からBREを生成
  - GUI
