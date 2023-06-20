package geometry

import (
	"fmt"
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

func (p Point) String() string {
	if len(p) < 2 {
		return ""
	}

	return fmt.Sprintf("(%f,%f)", p.X(), p.Y())
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
