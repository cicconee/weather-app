package geometry

type Polygon []PointCollection

func (p Polygon) Permiter() PointCollection {
	if len(p) == 0 {
		return nil
	}

	return p[0]
}

func (p Polygon) Holes() []PointCollection {
	if len(p) < 2 {
		return nil
	}

	// make a copy to prevent memory leaks.
	holes := make([]PointCollection, len(p)-1)
	copy(holes, p[1:])

	return holes
}

func (p Polygon) AsMultiPolygon() MultiPolygon {
	return MultiPolygon{p}
}

type MultiPolygon []Polygon
