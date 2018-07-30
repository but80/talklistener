package main

import (
	"fmt"
	"os"

	_ "github.com/theckman/goconstraint/go1.10/gte"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "USAGE: talklistener <input wav> <input text> [<output vsqx>]")
		os.Exit(1)
	}
	wavfile := os.Args[1]
	txtfile := os.Args[2]
	outfile := ""
	if 4 <= len(os.Args) {
		outfile = os.Args[3]
	}
	if err := generate(wavfile, txtfile, outfile); err != nil {
		panic(err)
	}
}
