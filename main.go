package main

import (
	"decrazifier"
	"flag"
	"os"
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("usage: de [file]")
	}
	imgFile := args[0]

	file, err := os.Open(imgFile)
	if err != nil {
		log.Fatalf("Could not open file: %v\n", err.Error())
	}

	w, err := os.OpenFile(outFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("could not open file for writing: %v\n", outFile)
	}
	defer w.Close()

	process(file, w)
}
