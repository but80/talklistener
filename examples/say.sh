#!/bin/bash

cd $( dirname "$0" )

if ! [ -e ../talklistener ]; then
  cd ..
  go run mage.go build || exit $?
  cd examples
fi

if ! ( which sox >/dev/null ); then
  echo 'brew install sox にてSoXをインストールしてください' >&2
  exit 1
fi

say -v Kyoko -o say 'これはテストです' || exit $?
sox say.aiff -r 16000 say.16k.wav || exit $?
echo 'これわてすとです' > say.txt
../talklistener say.16k.wav say.txt say.vsqx
