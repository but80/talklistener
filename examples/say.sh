#!/bin/bash

cd $( dirname "$0" )

if ! ( which talklistener >/dev/null ); then
  echo 'talklistener をインストールしてください' >&2
  exit 1
fi

say -v Kyoko -o say 'これはテストです' || exit $?
afconvert -f WAVE -d I16@16000 -c 1 --mix -o say.wav say.aiff || exit $?
echo 'これわてすとです' > say.txt
talklistener say.wav say.txt say.vsqx
