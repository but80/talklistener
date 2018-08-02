#!/bin/bash

cd $( dirname "$0" )

if ! ( which talklistener >/dev/null ); then
  echo 'talklistener をインストールしてください' >&2
  exit 1
fi

say -v Kyoko -o say 'これはテストです' || exit $?
# echo 'これわてすとです' > say.txt
talklistener say.aiff
