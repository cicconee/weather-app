package geometry

import (
	"fmt"
	"math"
	"strings"
)

type Point []float64

func NewPoint(x, y float64) Point {
	return Point{y, x}
}

func (p Point) X() float64 {
	return p[1]
}

func (p Point) Y() float64 {
	return p[0]
}

func (p Point) Lon() float64 {
	return p.X()
}

func (p Point) Lat() float64 {
	return p.Y()
}

// RoundedLon returns the longitude rounded to the 4th
// decimal place.
func (p Point) RoundedLon() float64 {
	return round(p.Lon(), 4)
}

// RoundedLat returns the latitude rounded to the 4th
// decimal place.
func (p Point) RoundedLat() float64 {
	return round(p.Lat(), 4)
}

func round(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func (p Point) String() string {
	if len(p) < 2 {
		return ""
	}

	return fmt.Sprintf("(%f,%f)", p.X(), p.Y())
}

// RoundedString returns the string representation of this point
// with the longitude and latitude rounded to the 4th decimal place.
func (p Point) RoundedString() string {
	if len(p) < 2 {
		return ""
	}

	return fmt.Sprintf("(%f, %f)", p.RoundedLon(), p.RoundedLat())
}

type PointCollection []Point

func (p PointCollection) String() string {
	if len(p) == 0 {
		return ""
	}

	var ss []string
	for _, pt := range p {
		ss = append(ss, pt.String())
	}

	return fmt.Sprintf("(%s)", strings.Join(ss, ","))
}
