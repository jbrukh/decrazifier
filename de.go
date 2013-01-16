package main

import (
	"flag"
	"image"
	"image/draw"
	"log"
	"os"

	"image/jpeg"
)

var (
	expWidth  = 240 // expected width of the image
	expHeight = 240 // expected height of the image
	expSquare = 60  // expected side of the subsquares of the image
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("usage: de [file]")
	}
	imgFile := args[0]

	log.Printf("opening %v...\n", imgFile)
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
	getFirstSquare(m, "output.jpeg")
}

func getFirstSquare(m image.Image, outFile string) {
	out := image.NewRGBA(image.Rect(0, 0, expSquare, expSquare))
	draw.Draw(out, out.Bounds(), m, image.ZP, draw.Src)

	w, err := os.OpenFile(outFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("could not open file for writing: %v\n", outFile)
	}
	defer w.Close()

	opts := &jpeg.Options{100}
	jpeg.Encode(w, out, opts)
}

// verifyBounds checks to see if the image has the expected
// size and sub-square properties.
func verifyBounds(m image.Image) {
	b := m.Bounds()
	if b.Max.Y != expHeight || b.Max.X != expWidth || b.Min.X != 0 || b.Min.Y != 0 {
		log.Fatal("unexpected bounds")
	}
	if expWidth%expSquare != 0 || expHeight%expSquare != 0 {
		log.Fatal("squares do not exhaust the image")
	}
}
