package draw

import (
	"image"
	"image/color"
)

// ellipse defines the struct components required to apply the ellipse's formula.
type ellipse struct {
	cx int // center x
	cy int // center y
	rx int // semi-major axis x
	ry int // semi-minor axis y
}

func NewEllipse(cx, cy, rx, ry int) *ellipse {
	return &ellipse{cx, cy, rx, ry}
}

func (e *ellipse) ColorModel() color.Model {
	return color.AlphaModel
}

func (e *ellipse) Bounds() image.Rectangle {
	min := image.Point{
		X: e.cx - e.rx,
		Y: e.cy - e.ry,
	}
	max := image.Point{
		X: e.cx + e.rx,
		Y: e.cy + e.ry,
	}
	return image.Rectangle{Min: min, Max: max} // size of just mask
}

func (e *ellipse) At(x, y int) color.Color {
	// Equation of ellipse
	p1 := float64((x-e.cx)*(x-e.cx)) / float64(e.rx*e.rx)
	p2 := float64((y-e.cy)*(y-e.cy)) / float64(e.ry*e.ry)
	eqn := p1 + p2

	if eqn <= 1 {
		// rMin := math.Min(float64(e.rx), float64(e.ry))
		// rMax := math.Max(float64(e.rx), float64(e.ry))

		// grad := NewRadialGradient(float64(x), float64(y), rMin*0.7, float64(x), float64(y), rMax-(rMin/4))
		// grad.AddColorStop(0, color.RGBA{255, 255, 255, 255})
		// grad.AddColorStop(0.6, color.RGBA{127, 127, 127, 127})
		// grad.AddColorStop(1, color.RGBA{0, 0, 0, 0})

		// return grad.ColorAt(e.cx, e.cy)
		return color.Alpha{255}
	}
	return color.Alpha{0}
}
