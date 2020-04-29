package pixelate

import (
	"image"
	"image/color"
	"math"

	"github.com/fogleman/gg"
)

type context struct {
	*gg.Context
}

// Brightness factor
var bf = 1.0005

// Draw creates uniform cells with the quantified cell color of the source image.
func (quant *Quant) Draw(img image.Image, numOfColors int, csize int, useNoise bool) image.Image {
	var cellSize int

	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	imgRatio := func(w, h int) float64 {
		var ratio float64
		if w > h {
			ratio = float64((w / h) * w)
		} else {
			ratio = float64((h / w) * h)
		}
		return ratio
	}

	if csize == 0 {
		cellSize = int(round(imgRatio(dx, dy) * 0.015))
	} else {
		cellSize = csize
	}
	qimg := quant.Quantize(img, numOfColors)

	ctx := &context{gg.NewContext(dx, dy)}
	ctx.SetRGB(1, 1, 1)
	ctx.Clear()
	ctx.SetRGB(0, 0, 0)
	rgba := ctx.convertToNRGBA64(qimg)

	for x := 0; x < dx; x += cellSize {
		for y := 0; y < dy; y += cellSize {
			rect := image.Rect(x, y, x+cellSize, y+cellSize)
			rect = rect.Intersect(qimg.Bounds())
			if rect.Empty() {
				rect = image.ZR
			}
			subImg := rgba.SubImage(rect).(*image.NRGBA64)
			cellColor := ctx.getAvgColor(subImg)
			ctx.drawCell(float64(x), float64(y), float64(cellSize), cellColor)
		}
	}
	ctxImg := ctx.Image()
	if useNoise {
		return noise(ctxImg, dx, dy, 12)
	}
	return ctxImg
}

// drawCell draws the cell filling up with the quantified color
func (ctx *context) drawCell(x, y, cellSize float64, c color.NRGBA64) {
	ctx.DrawRectangle(x, y, x+cellSize, y+cellSize)
	ctx.SetRGBA(float64(c.R/255^0xff)*bf, float64(c.G/255^0xff)*bf, float64(c.B/255^0xff)*bf, 1)
	ctx.Fill()
}

// getAvgColor get the average color of a cell
func (ctx *context) getAvgColor(img *image.NRGBA64) color.NRGBA64 {
	var (
		bounds  = img.Bounds()
		r, g, b int
	)

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			var c = img.NRGBA64At(x, y)
			r += int(c.R)
			g += int(c.G)
			b += int(c.B)
		}
	}

	return color.NRGBA64{
		R: maxUint16(0, minUint16(65535, uint16(r/(bounds.Dx()*bounds.Dy())))),
		G: maxUint16(0, minUint16(65535, uint16(g/(bounds.Dx()*bounds.Dy())))),
		B: maxUint16(0, minUint16(65535, uint16(b/(bounds.Dx()*bounds.Dy())))),
		A: 255,
	}
}

// convertToNRGBA64 converts an image.Image into an image.NRGBA64.
func (ctx *context) convertToNRGBA64(img image.Image) *image.NRGBA64 {
	var (
		bounds = img.Bounds()
		nrgba  = image.NewNRGBA64(bounds)
	)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			nrgba.Set(x, y, img.At(x, y))
		}
	}
	return nrgba
}

// round number down.
func round(x float64) float64 {
	return math.Floor(x)
}

// minUint16 returns the smallest number between two uint16 numbers.
func minUint16(x, y uint16) uint16 {
	if x < y {
		return x
	}
	return y
}

// maxUint16 returns the biggest number between two uint16 numbers.
func maxUint16(x, y uint16) uint16 {
	if x > y {
		return x
	}
	return y
}
