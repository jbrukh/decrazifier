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
	"strings"
	"time"
)

var (
	expWidth  = 240 // expected width of the image
	expHeight = 240 // expected height of the image
	expSquare = 60  // expected side of the subsquares of the image
	xCoords   = []int{0, 60, 120, 180}
	yCoords   = []int{0, 60, 120, 180}
	L         = len(yCoords) * len(xCoords)
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

}

func toSquare(x, y int) int {
	return x + len(xCoords)*y
}

func fromSquare(s int) (x, y int) {
	return (s % 4) * expSquare, (s / len(xCoords)) * expSquare
}

func fromWide(s int) (x, y int) {
	return 0, s * expSquare
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

	table := make(map[int]int, L)
	for s := 0; s < L; s++ {
		best := computeBest(s, m)
		table[s] = best
	}

	W := len(xCoords)
	out := image.NewRGBA(image.Rect(0, 0, W*expSquare, L*expSquare))
	for v := 0; v < L; v++ {
		thisOne := v
		for h := 0; h < len(xCoords); h++ {
			r := image.Rect(h*expSquare, v*expSquare, (h+1)*expSquare, (v+1)*expSquare)
			if h > 0 {
				thisOne = table[thisOne]
			}
			bx, by := fromSquare(thisOne)
			draw.Draw(out, r, m, image.Pt(bx, by), draw.Src)
		}
	}

	toFile(out, "poopie.jpg")

	table2 := make(map[int]int, L)
	for s := 0; s < L; s++ {
		best := computeBestWide(s, out)
		log.Printf("wide: %v -> %v\n", s, best)
		table2[s] = best
	}

	out2 := image.NewRGBA(image.Rect(0, 0, W*L*expSquare, len(yCoords)*expSquare))
	for h := 0; h < L; h++ {
		thisOne := h
		for v := 0; v < W; v++ {
			r := image.Rect(h*expSquare*W, v*expSquare, (h+1)*expSquare*W, (v+1)*expSquare)
			if v > 0 {
				thisOne = table2[thisOne]
			}
			bx, by := fromWide(thisOne)
			draw.Draw(out2, r, out, image.Pt(bx, by), draw.Src)
		}
	}
	name := strings.Split(imgFile, ".")[0]
	outFile := fmt.Sprintf("%v-descrambled.jpg", name)
	log.Println(outFile)
	toFile(out2, outFile)
}

func computeBestWide(s int, m image.Image) int {
	sy := (s+1)*expSquare - 1 // bottom of selected
	var min float64
	var best int = -1
	for i := 0; i < L; i++ {
		if i != s {
			// compare the two
			y := i * expSquare // top of comparison rect
			//log.Printf("comparing %v (%v,%v) and %v (%v, %v)... ", s, sx, sy, i, x, y)
			v := make([]float64, expSquare*len(xCoords))
			for i, _ := range v {
				d := colorDist(m.At(i, sy), m.At(i, y))
				v[i] = d
			}

			// now, see the distance
			D := dist(v)
			//log.Printf("%v ", D)
			if D < min || min == 0 {
				min = D
				best = i
				//log.Printf("MIN!\n")
			}
		}
	}
	return best
}

func computeBest(s int, m image.Image) int {
	sx, sy := fromSquare(s)
	var min float64
	var best int = -1
	for i := 0; i < len(xCoords)*len(yCoords); i++ {
		if i != s {
			// compare the two
			x, y := fromSquare(i)
			//log.Printf("comparing %v (%v,%v) and %v (%v, %v)... ", s, sx, sy, i, x, y)
			v := make([]float64, expSquare)
			for i, _ := range v {
				d := colorDist(m.At(sx+expSquare-1, sy+i), m.At(x, y+i))
				v[i] = d
			}

			// now, see the distance
			D := dist(v)
			//log.Printf("%v ", D)
			if D < min || min == 0 {
				min = D
				best = i
				//log.Printf("MIN!\n")
			}
		}
	}
	return best
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
