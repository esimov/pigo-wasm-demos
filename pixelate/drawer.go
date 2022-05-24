package pixelate

import (
	"image"
	"image/color"
)

// Draw creates uniform cells with the quantified cell color of the source image.
func (quant *Quant) Draw(src image.Image, numOfColors int, cellSize int, noiseLevel int) image.Image {
	dx, dy := src.Bounds().Dx(), src.Bounds().Dy()
	dst := image.NewNRGBA64(src.Bounds())

	// Calculate the image aspect ratio.
	imgRatio := func(w, h int) float64 {
		var ratio float64
		if w > h {
			ratio = float64((w / h) * w)
		} else {
			ratio = float64((h / w) * h)
		}
		return ratio
	}

	if cellSize == 0 {
		cellSize = int(imgRatio(dx, dy) * 0.015)
	} else {
		cellSize = cellSize
	}

	qimg := quant.Quantize(src, numOfColors)

	for x := 0; x < dx; x += cellSize {
		for y := 0; y < dy; y += cellSize {
			rect := image.Rect(x, y, x+cellSize, y+cellSize)
			rect = rect.Intersect(qimg.Bounds())
			if rect.Empty() {
				rect = image.ZR
			}
			subImg := qimg.(*image.Paletted).SubImage(rect).(*image.Paletted)
			cellColor := getAvgColor(subImg)

			// Fill up the cell with the quantified color.
			for xx := x; xx < x+cellSize; xx++ {
				for yy := y; yy < y+cellSize; yy++ {
					dst.Set(xx, yy, cellColor)
				}
			}
		}
	}
	if noiseLevel > 0 {
		addNoise(dst, noiseLevel)
	}
	return dst
}

// getAvgColor get the average color of a cell
func getAvgColor(img *image.Paletted) color.NRGBA64 {
	var (
		bounds  = img.Bounds()
		r, g, b int
	)

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			cr, cg, cb, _ := img.At(x, y).RGBA()
			r += int(cr)
			g += int(cg)
			b += int(cb)
		}
	}

	return color.NRGBA64{
		R: max(0, min(65535, uint16(r/(bounds.Dx()*bounds.Dy())))),
		G: max(0, min(65535, uint16(g/(bounds.Dx()*bounds.Dy())))),
		B: max(0, min(65535, uint16(b/(bounds.Dx()*bounds.Dy())))),
		A: 255,
	}
}
