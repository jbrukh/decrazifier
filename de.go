package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"os"

	_ "image/jpeg"
)

var (
	expWidth  = 240
	expHeight = 240
	expSquare = 60
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("usage: de [file]")
		os.Exit(1)
	}
	imgFile := args[0]

	fmt.Printf("opening %v...\n", imgFile)
	file, err := os.Open(imgFile)
	if err != nil {
		log.Fatalf("could not open file %v (%v)", imgFile, err)
	}
	defer file.Close()

	m, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	verifyBounds(m)
}

func verifyBounds(m image.Image) {
	b := m.Bounds()
	if b.Max.Y != expHeight || b.Max.X != expWidth || b.Min.X != 0 || b.Min.Y != 0 {
		log.Fatal("unexpected bounds")
	}
}
