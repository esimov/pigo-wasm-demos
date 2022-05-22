package pixels

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"syscall/js"
	"time"
)

// ImgToPix converts an image to an 1D uint8 pixel array.
// In order to preserve the color information per pixel we need to reconstruct the array to a specific format.
func ImgToPix(src *image.NRGBA) []uint8 {
	size := src.Bounds().Size()
	width, height := size.X, size.Y
	pixels := make([][][3]uint8, height)

	for y := 0; y < height; y++ {
		row := make([][3]uint8, width)
		for x := 0; x < width; x++ {
			idx := (y*width + x) * 4
			pix := src.Pix[idx : idx+4]
			row[x] = [3]uint8{uint8(pix[0]), uint8(pix[1]), uint8(pix[2])}
		}
		pixels[y] = row
	}

	// Flatten the 3d array to 1D.
	flattened := []uint8{}
	for x := 0; x < len(pixels); x++ {
		for y := 0; y < len(pixels[x]); y++ {
			r := pixels[x][y][0]
			g := pixels[x][y][1]
			b := pixels[x][y][2]
			flattened = append(flattened, r, g, b, 255)
		}
	}
	return flattened
}

// PixToImage converts the pixel data to an image.
func PixToImage(pixels []uint8, rect image.Rectangle) image.Image {
	img := image.NewNRGBA(rect)
	bounds := img.Bounds()
	dx, dy := bounds.Max.X, bounds.Max.Y
	col := color.NRGBA{}

	for y := bounds.Min.Y; y < dy; y++ {
		for x := bounds.Min.X; x < dx*4; x += 4 {
			col.R = pixels[x+y*dx*4]
			col.G = pixels[x+y*dx*4+1]
			col.B = pixels[x+y*dx*4+2]
			col.A = pixels[x+y*dx*4+3]

			img.SetNRGBA(int(x/4), y, col)
		}
	}
	return img
}

// RotateImg rotates the image per pixel level to a certain degree and returns an image data.
func RotateImg(img *image.NRGBA, angle float64) []uint8 {
	bounds := img.Bounds()
	dx, dy := bounds.Max.X, bounds.Max.Y
	col := color.NRGBA{}

	for x := bounds.Min.X; x < dx; x++ {
		for y := bounds.Min.Y; y < dy; y++ {
			x0, y0 := dx/2, dy/2
			xoff, yoff := x-x0, y-y0
			rad := angle * math.Pi / 180
			newX := int(math.Cos(rad)*float64(xoff) + math.Sin(rad)*float64(yoff) + float64(x0))
			newY := int(math.Cos(rad)*float64(yoff) - math.Sin(rad)*float64(xoff) + float64(y0))

			pos := x + (y * dx)
			col.R = img.Pix[pos*4]
			col.G = img.Pix[pos*4+1]
			col.B = img.Pix[pos*4+2]
			col.A = img.Pix[pos*4+3]

			img.SetNRGBA(newX, newY, col)
		}
	}
	return ImgToPix(img)
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

// LoadImage load the source image and encodes it to base64 format.
func LoadImage(path string) (string, error) {
	href := js.Global().Get("location").Get("href")
	u, err := url.Parse(href.String())
	if err != nil {
		return "", err
	}

	u.Path = path
	u.RawQuery = fmt.Sprint(time.Now().UnixNano())

	resp, err := http.Get(u.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}
