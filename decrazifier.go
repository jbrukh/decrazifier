package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"log"
	"math"
	"os"
)

const (
	Width      = 240 // width, in pixels, of input image
	Height     = 240 // height, in pixels, of input image
	Side       = 60  // side length of subsquare
	Horizontal = Width / Side
	Vertical   = Height / Side
	Total      = Horizontal * Vertical // total sub squares
)

type Edge struct {
	// EdgeFinder takes the rectangle and finds the starting
	// point of the given edge of that rectangle
	EdgeFinder func(r image.Rectangle) image.Point

	// Increment represents the direction to travel
	// along the edge
	Increment image.Point
}

// these variables describe the direction to move along
// an edge in
var (
	Right = &Edge{func(r image.Rectangle) image.Point {
		return image.Pt(r.Max.X-1, r.Min.Y)
	}, image.Pt(0, 1)}

	Bottom = &Edge{func(r image.Rectangle) image.Point {
		return image.Pt(r.Min.X, r.Max.Y-1)
	}, image.Pt(1, 0)}

	Left = &Edge{func(r image.Rectangle) image.Point {
		return image.Pt(r.Min.X, r.Min.Y)
	}, image.Pt(0, 1)}

	Top = &Edge{func(r image.Rectangle) image.Point {
		return image.Pt(r.Min.X, r.Min.Y)
	}, image.Pt(1, 0)}
)

// represents the input image and uses
// it to make calculations
type ScrambledImage struct {
	m image.Image
}

// represents a horizontal strip created
// by the concatenation of a "Horizontal" number
// of squares
type Strip struct {
	seq [Horizontal]int
	d   float64 // going to cache the distance function along the way
}

type StripSet struct {
	seq [Vertical]int
	d   float64 // going to cache the distance function along the way
}

// IterateEdge 
func (s *ScrambledImage) IterateEdge(r image.Rectangle, e *Edge, d int) <-chan color.Color {
	ch := make(chan color.Color, Side)

	// get the starting point of the edge
	pt := e.EdgeFinder(r)

	// go that distance
	for j := 0; j < d; j++ {
		ch <- s.m.At(pt.X, pt.Y)
		pt = pt.Add(e.Increment)
	}

	close(ch)
	return ch
}

func NewScrambledImage(r io.Reader) (*ScrambledImage, error) {
	// get the image data
	m, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	// verify the bounds
	b := m.Bounds()
	if b.Max.Y != Height || b.Max.X != Width || b.Min.X != 0 || b.Min.Y != 0 {
		return nil, fmt.Errorf("Incorrect dimensions: expecting %vx%v", Width, Height)
	}
	if Width%Side != 0 || Height%Side != 0 {
		return nil, fmt.Errorf("Incorrect subsquare length: expecting %v", Side)
	}

	// return
	return &ScrambledImage{m}, nil
}

// returns the rectangle of the n-th tile, where the tile
// number is 0...(Total-1)
func Tile(n int) image.Rectangle {
	if n < 0 || n >= Total {
		panic("Out of bounds")
	}
	x, y := (n%Horizontal)*Side, (n/Horizontal)*Side
	return image.Rect(x, y, x+Side, y+Side)
}

// return the distance between tiles, comparing the aEdge
// of a and the bEdge of b; smaller distance
// corresponds to a smoother transition between tiles
func (s *ScrambledImage) CompareTiles(a int, aEdge *Edge, b int, bEdge *Edge) float64 {
	aCh, bCh := s.IterateEdge(Tile(a), aEdge, Side), s.IterateEdge(Tile(b), bEdge, Side)
	if len(aCh) != len(bCh) {
		panic("Edges of different lengths?")
	}
	var sum float64
	for {
		aPt, ok := <-aCh
		if !ok {
			break
		}
		bPt := <-bCh
		sum += distance(aPt, bPt)
	}
	return sum
}

func (s *ScrambledImage) CompareStrips(a *Strip, aEdge *Edge, b *Strip, bEdge *Edge) float64 {
	var sum float64
	for i := 0; i < Horizontal; i++ {
		sum += s.CompareTiles(a.seq[i], aEdge, b.seq[i], bEdge)
	}
	return sum
}

func (s *ScrambledImage) Descramble(strips []*Strip) *StripSet {
	// for each strip
	var min float64 = math.NaN()
	var minStripSet *StripSet

	for i := 0; i < len(strips); i++ { // should == Total
		// find a strip set and measure it's distance rating
		stripSet := s.StripSet(i, strips)
		d := stripSet.d
		for _, v := range stripSet.seq {
			d += strips[v].d
		}

		if math.IsNaN(min) || d < min {
			min = d
			minStripSet = stripSet
		}
	}

	return minStripSet
}

func (s *ScrambledImage) StripSet(n int, strips []*Strip) *StripSet {
	seen := make(map[int]bool, Vertical)
	seen[n] = true

	stripSet := new(StripSet)
	stripSet.seq[0] = n

	for j := 1; j < Vertical; j++ {
		var min float64 = math.NaN()
		var minStrip int

		for i := 0; i < len(strips); i++ {
			if _, ok := seen[i]; !ok {
				strip := strips[stripSet.seq[j-1]]
				strip2 := strips[i]
				d := s.CompareStrips(strip, Bottom, strip2, Top)
				if math.IsNaN(min) || d < min {
					min = d
					minStrip = i
				}
			}
		}

		stripSet.seq[j] = minStrip
		stripSet.d += min
		seen[minStrip] = true
	}

	return stripSet
}

// returns the best horizontal strip generated from the n-th tile; thus
// there is a 1-to-1 mapping between tiles and strips
func (s *ScrambledImage) Strip(tile int) *Strip {
	seen := make(map[int]bool, Horizontal)
	seen[tile] = true

	strip := new(Strip)
	strip.seq[0] = tile // TODO we don't need this value
	for j := 1; j < Horizontal; j++ {
		var min float64 = math.NaN()
		var minTile int
		// calculate the tile with the minimum distance
		for i := 0; i < Total; i++ {
			// if we haven't used this tile before
			if _, ok := seen[i]; !ok {
				// calculate the distance
				d := s.CompareTiles(strip.seq[j-1], Right, i, Left)
				//log.Printf("%v -> %v : %v", tile, i, d)
				//log.Printf("%v : %v", i, d)
				if math.IsNaN(min) || d < min {
					min = d
					minTile = i
				}
			}
		}
		// remember the smallest tile
		strip.seq[j] = minTile
		strip.d += min // caching
		seen[minTile] = true

	}
	return strip
}

// calculates the Euclidean distance between two colors
func distance(c1, c2 color.Color) float64 {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	sum := float64(uint64((r1-r2)*(r1-r2)) + uint64((g1-g2)*(g1-g2)) + uint64((b1-b2)*(b1-b2)) + uint64((a1-a2)*(a1-a2)))
	return math.Sqrt(sum)
}

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

	process(file)
}

func process(file io.Reader) {
	s, err := NewScrambledImage(file)
	if err != nil {
		log.Fatalf("Could not convert file: %v\n", err.Error())
	}

	m := image.NewRGBA(image.Rect(0, 0, Width, Total*Side))
	strips := make([]*Strip, 0, Total)
	for i := 0; i < Total; i++ {
		strip := s.Strip(i)
		strips = append(strips, strip)
		log.Printf("%v\n", strip)
		for j := 0; j < Horizontal; j++ {
			draw.Draw(m, image.Rect(j*Side, i*Side, (j+1)*Side, (i+1)*Side), s.m, Tile(strip.seq[j]).Min, draw.Src)
		}
	}
	toFile(m, "jake.jpg")

	stripSet := s.Descramble(strips)
	m = image.NewRGBA(image.Rect(0, 0, Width, Height))
	for i, strip := range stripSet.seq {
		for j, tile := range strips[strip].seq {
			draw.Draw(m, Tile(Horizontal*i+j), s.m, Tile(tile).Min, draw.Src)
		}
	}
	toFile(m, "jake2.jpg")
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
