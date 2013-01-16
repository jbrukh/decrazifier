package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"math"
	"math/rand"
	"os"
	"time"
)

var (
	expWidth  = 240 // expected width of the image
	expHeight = 240 // expected height of the image
	expSquare = 60  // expected side of the subsquares of the image
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

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

	// count := 0
	// for y := 0; y < expWidth; y += expSquare {
	// 	for x := 0; x < expHeight; x += expSquare {
	// 		sp := image.Pt(x, y)
	// 		outFile := fmt.Sprintf("poop-%d.jpg", count)
	// 		log.Printf("generating %v...\n", outFile)
	// 		getSquare(m, sp, outFile)
	// 		count++
	// 	}
	// }

	log.Printf("total: %v\n", total(m))

	out := m
	min := total(m)
	for i := 0; i < 1000000; i++ {
		out = scramble(out)
		d := total(out)
		if d < min {
			min = d
			outFile := fmt.Sprintf("poop-%f.jpg", min)
			toFile(out, outFile)
		}
	}
}

func scramble(m image.Image) image.Image {
	out := image.NewRGBA(image.Rect(0, 0, expWidth, expHeight))

	// first scramble vertically
	v := make([]int, expWidth/expSquare)
	for i, _ := range v {
		v[i] = i * expSquare
	}
	shuffle(v)
	fmt.Printf("%v\n", v)
	for i, x := range v {
		draw.Draw(out, image.Rect(x, 0, x+expSquare, expHeight), m, image.Pt(i*expSquare, 0), draw.Src)
	}
	//toFile(out, "poopie.jpg")

	// then scramble horizontally
	out2 := image.NewRGBA(image.Rect(0, 0, expWidth, expHeight))
	v = make([]int, expHeight/expSquare)
	for i, _ := range v {
		v[i] = i * expSquare
	}
	shuffle(v)
	fmt.Printf("%v\n", v)
	for i, y := range v {
		draw.Draw(out2, image.Rect(0, y, expWidth, y+expSquare), out, image.Pt(0, i*expSquare), draw.Src)
	}

	//toFile(out2, "poopie.jpg")
	return out2
}

func shuffle(v []int) {
	n := len(v)
	for i := 0; i < n; i++ {
		j := rand.Intn(n-i) + i
		v[i], v[j] = v[j], v[i]
	}
}

func total(m image.Image) (result float64) {
	for x := 60; x < expWidth; x += expSquare {
		r := vDist(m, x)
		log.Printf("v: %v", r)
		result += r * r
	}
	for y := 60; y < expHeight; y += expSquare {
		r := hDist(m, y)
		log.Printf("h: %v", r)
		result += r * r
	}
	return math.Sqrt(result)
}

func vDist(m image.Image, x int) float64 {
	if x < expSquare || x >= expWidth {
		log.Fatal("vCoord is wrong")
	}
	squares := make([]float64, 0, expHeight)
	for y := 0; y < expHeight; y++ {
		squares = append(squares, colorDist(m.At(x-1, y), m.At(x, y)))
	}

	return dist(squares)
}

func hDist(m image.Image, y int) float64 {
	if y < expSquare || y >= expHeight {
		log.Fatal("hCoord is wrong")
	}
	squares := make([]float64, 0, expWidth)
	for x := 0; x < expWidth; x++ {
		squares = append(squares, colorDist(m.At(x, y-1), m.At(x, y)))
	}
	return dist(squares)
}

func colorDist(c1, c2 color.Color) float64 {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	sum := float64(uint64((r1-r2)*(r1-r2)) + uint64((g1-g2)*(g1-g2)) + uint64((b1-b2)*(b1-b2)) + uint64((a1-a2)*(a1-a2)))
	return math.Sqrt(sum)
}

func dist(v []float64) (result float64) {
	for _, p := range v {
		result += p * p
	}
	return math.Sqrt(result)
}

func getSquare(m image.Image, sp image.Point, outFile string) {
	out := image.NewRGBA(image.Rect(0, 0, expSquare, expSquare))
	draw.Draw(out, out.Bounds(), m, sp, draw.Src)

	w, err := os.OpenFile(outFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("could not open file for writing: %v\n", outFile)
	}
	defer w.Close()

	opts := &jpeg.Options{100}
	jpeg.Encode(w, out, opts)
}

func toFile(m image.Image, outFile string) {
	w, err := os.OpenFile(outFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("could not open file for writing: %v\n", outFile)
	}
	defer w.Close()
	opts := &jpeg.Options{100}
	jpeg.Encode(w, m, opts)
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
