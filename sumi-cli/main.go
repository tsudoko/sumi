package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tsudoko/go.tesseract"

	"github.com/tsudoko/sumi/ocr"
)

var lang = flag.String("l", "jpn", "language(s) used for OCR")

func main() {
	flag.Parse()

	t, err := tesseract.NewTess("", *lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing tesseract: %s\n", err.Error())
		return
	}

	matches, err := ocr.Detect(t, flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error detecting characters: %s\n", err.Error())
		return
	}

	for _, clist := range matches {
		fmt.Printf("%c", clist[0])
	}
	fmt.Printf("\n")
	for _, clist := range matches {
		for _, c := range clist {
			fmt.Printf("%c", c)
		}
		fmt.Printf("\n")
	}
}
