package pixels

import (
	"image"
	"image/color"
	"math"
)

// ImgToPix converts an image to pixel data.
func ImgToPix(img image.Image) []uint8 {
	bounds := img.Bounds()
	pixels := make([]uint8, 0, bounds.Max.X*bounds.Max.Y*4)

	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
			r, g, b, _ := img.At(i, j).RGBA()
			pixels = append(pixels, uint8(r>>8), uint8(g>>8), uint8(b>>8), 255)
		}
	}
	return pixels
}

// PixToImage converts the pixel data to an image.
func PixToImage(pixels []uint8, dim int) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, dim, dim))
	bounds := img.Bounds()
	dx, dy := bounds.Max.X, bounds.Max.Y
	col := color.NRGBA{}

	for y := bounds.Min.Y; y < dy; y++ {
		for x := bounds.Min.X; x < dx*4; x += 4 {
			col.R = pixels[x+y*dx*4]
			col.G = pixels[x+y*dx*4+1]
			col.B = pixels[x+y*dx*4+2]
			col.A = pixels[x+y*dx*4+3]

			img.SetNRGBA(y, int(x/4), col)
		}
	}
	return img
}

// RgbaToGrayscale converts the pixel data to grayscale mode.
func RgbaToGrayscale(data []uint8, dx, dy int) []uint8 {
	for r := 0; r < dx; r++ {
		for c := 0; c < dy; c++ {
			// gray = 0.2*red + 0.7*green + 0.1*blue
			data[r*dy+c] = uint8(math.Round(
				0.2126*float64(data[r*4*dy+4*c+0]) +
					0.7152*float64(data[r*4*dy+4*c+1]) +
					0.0722*float64(data[r*4*dy+4*c+2])))
		}
	}
	return data
}
