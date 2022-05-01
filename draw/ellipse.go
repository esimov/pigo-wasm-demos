package draw

import (
	"image"
	"image/color"
)

// Ellipse defines the struct components required to apply the ellipse's formula.
type Ellipse struct {
	Cx int // center x
	Cy int // center y
	Rx int // semi-major axis x
	Ry int // semi-minor axis y
}

func (e *Ellipse) ColorModel() color.Model {
	return color.AlphaModel
}

func (e *Ellipse) Bounds() image.Rectangle {
	min := image.Point{
		X: e.Cx - e.Rx,
		Y: e.Cy - e.Ry,
	}
	max := image.Point{
		X: e.Cx + e.Rx,
		Y: e.Cy + e.Ry,
	}
	return image.Rectangle{Min: min, Max: max} // size of just mask
}

func (e *Ellipse) At(x, y int) color.Color {
	// Equation of ellipse
	p1 := float64((x-e.Cx)*(x-e.Cx)) / float64(e.Rx*e.Rx)
	p2 := float64((y-e.Cy)*(y-e.Cy)) / float64(e.Ry*e.Ry)
	eqn := p1 + p2

	if eqn <= 1 {
		return color.Alpha{255}
	}
	return color.Alpha{0}
}
