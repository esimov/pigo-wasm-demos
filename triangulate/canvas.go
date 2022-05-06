package triangulate

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"sync"
	"syscall/js"

	"github.com/esimov/pigo-wasm-demos/detector"
	ellipse "github.com/esimov/pigo-wasm-demos/draw"
	triangle "github.com/esimov/triangle/v2"
	"golang.org/x/sync/errgroup"
)

// Canvas struct holds the Javascript objects needed for the Canvas creation
type Canvas struct {
	done   chan struct{}
	succCh chan struct{}
	errCh  chan error
	lock   sync.Mutex
	g      *errgroup.Group

	// DOM elements
	window     js.Value
	doc        js.Value
	body       js.Value
	windowSize struct{ width, height int }

	// Canvas properties
	canvas   js.Value
	ctx      js.Value
	reqID    js.Value
	renderer js.Func

	// Webcam properties
	navigator js.Value
	video     js.Value

	// Delaunay triangulation related variables
	triangle  *triangle.Image
	processor *triangle.Processor
	frame     *image.NRGBA

	// Canvas interaction related variables
	showFrame       bool
	isSolid         bool
	isGrayScaled    bool
	wireframe       int
	trianglePoints  int
	pointsThreshold int
	pointRate       float64
	strokeWidth     float64
}

const (
	minTrianglePoints = 150
	maxTrianglePoints = 750

	minPointsThreshold = 2
	maxPointsThreshold = 25

	minPointRate = 0.010
	maxPointRate = 0.095

	minStrokeWidth = 0
	maxStrokeWidth = 4
)

var det *detector.Detector

// NewCanvas creates and initializes the new Canvas element
func NewCanvas() *Canvas {
	var c Canvas
	c.window = js.Global()
	c.doc = c.window.Get("document")
	c.body = c.doc.Get("body")

	c.windowSize.width = 640
	c.windowSize.height = 480

	c.canvas = c.doc.Call("createElement", "canvas")
	c.canvas.Set("width", c.windowSize.width)
	c.canvas.Set("height", c.windowSize.height)
	c.canvas.Set("id", "canvas")
	c.body.Call("appendChild", c.canvas)

	c.ctx = c.canvas.Call("getContext", "2d")
	c.showFrame = false
	c.isSolid = false
	c.isGrayScaled = false

	c.wireframe = triangle.WithoutWireframe
	c.strokeWidth = 0
	c.trianglePoints = 450
	c.pointsThreshold = 10
	c.pointRate = 0.075

	det = detector.NewDetector()

	c.processor = &triangle.Processor{
		BlurRadius:      2,
		Noise:           0,
		BlurFactor:      2,
		EdgeFactor:      4,
		PointRate:       c.pointRate,
		MaxPoints:       c.trianglePoints,
		PointsThreshold: c.pointsThreshold,
		Wireframe:       c.wireframe,
		StrokeWidth:     c.strokeWidth,
		IsStrokeSolid:   c.isSolid,
		Grayscale:       c.isGrayScaled,
		BgColor:         "#ffffff00",
	}
	c.lock = sync.Mutex{}
	c.g = &errgroup.Group{}

	c.triangle = &triangle.Image{*c.processor}

	return &c
}

// Render calls the `requestAnimationFrame` Javascript function in asynchronous mode.
func (c *Canvas) Render() error {
	width, height := c.windowSize.width, c.windowSize.height
	var data = make([]byte, width*height*4)
	c.done = make(chan struct{})

	err := det.UnpackCascades()
	if err != nil {
		return err
	}

	c.renderer = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() error {
			c.window.Get("stats").Call("begin")

			c.reqID = c.window.Call("requestAnimationFrame", c.renderer)
			// Draw the webcam frame to the canvas element
			c.ctx.Call("drawImage", c.video, 0, 0)
			rgba := c.ctx.Call("getImageData", 0, 0, width, height).Get("data")

			// Convert the rgba value of type Uint8ClampedArray to Uint8Array in order to
			// be able to transfer it from Javascript to Go via the js.CopyBytesToGo function.
			uint8Arr := js.Global().Get("Uint8Array").New(rgba)
			js.CopyBytesToGo(data, uint8Arr)
			gray := c.rgbaToGrayscale(data)

			// Reset the data slice to its default values to avoid unnecessary memory allocation.
			// Otherwise, the GC won't clean up the memory address allocated by this slice
			// and the memory will keep up increasing by each iteration.
			data = make([]byte, len(data))

			res := det.DetectFaces(gray, height, width)
			if err := c.drawDetection(res); err != nil {
				return err
			}
			c.window.Get("stats").Call("end")

			return nil
		}()
		return nil
	})
	// Release renderer to free up resources.
	defer c.renderer.Release()

	c.window.Call("requestAnimationFrame", c.renderer)
	c.detectKeyPress()
	<-c.done

	return nil
}

// Stop stops the rendering.
func (c *Canvas) Stop() {
	c.window.Call("cancelAnimationFrame", c.reqID)
	c.done <- struct{}{}
	close(c.done)
}

// StartWebcam reads the webcam data and feeds it into the canvas element.
// It returns an empty struct in case of success and error in case of failure.
func (c *Canvas) StartWebcam() (*Canvas, error) {
	var err error
	c.succCh = make(chan struct{})
	c.errCh = make(chan error)

	c.video = c.doc.Call("createElement", "video")

	// If we don't do this, the stream will not be played.
	c.video.Set("autoplay", 1)
	c.video.Set("playsinline", 1) // important for iPhones

	// The video should fill out all of the canvas
	c.video.Set("width", 0)
	c.video.Set("height", 0)

	c.body.Call("appendChild", c.video)

	success := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			c.video.Set("srcObject", args[0])
			c.video.Call("play")
			c.succCh <- struct{}{}
		}()
		return nil
	})

	failure := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			err = fmt.Errorf("failed initialising the camera: %s", args[0].String())
			c.errCh <- err
		}()
		return nil
	})

	opts := js.Global().Get("Object").New()

	videoSize := js.Global().Get("Object").New()
	videoSize.Set("width", c.windowSize.width)
	videoSize.Set("height", c.windowSize.height)
	videoSize.Set("aspectRatio", 1.777777778)

	opts.Set("video", videoSize)
	opts.Set("audio", false)

	promise := c.window.Get("navigator").Get("mediaDevices").Call("getUserMedia", opts)
	promise.Call("then", success, failure)

	select {
	case <-c.succCh:
		return c, nil
	case err := <-c.errCh:
		return nil, err
	}
}

// rgbaToGrayscale converts the rgb pixel values to grayscale
func (c *Canvas) rgbaToGrayscale(data []uint8) []uint8 {
	rows, cols := c.windowSize.width, c.windowSize.height
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			// gray = 0.2*red + 0.7*green + 0.1*blue
			data[r*cols+c] = uint8(math.Round(
				0.2126*float64(data[r*4*cols+4*c+0]) +
					0.7152*float64(data[r*4*cols+4*c+1]) +
					0.0722*float64(data[r*4*cols+4*c+2])))
		}
	}
	return data
}

// pixToImage converts an array buffer to an image.
func (c *Canvas) pixToImage(pixels []uint8, dim int) image.Image {
	c.frame = image.NewNRGBA(image.Rect(0, 0, dim, dim))
	bounds := c.frame.Bounds()
	dx, dy := bounds.Max.X, bounds.Max.Y
	col := color.NRGBA{}

	for y := bounds.Min.Y; y < dy; y++ {
		for x := bounds.Min.X; x < dx*4; x += 4 {
			col.R = pixels[x+y*dx*4]
			col.G = pixels[x+y*dx*4+1]
			col.B = pixels[x+y*dx*4+2]
			col.A = pixels[x+y*dx*4+3]

			c.frame.SetNRGBA(y, int(x/4), col)
		}
	}
	return c.frame
}

// imgToPix converts an image to a buffer array
func (c *Canvas) imgToPix(img image.Image) []uint8 {
	bounds := img.Bounds()
	pixels := make([]uint8, 0, bounds.Max.X*bounds.Max.Y*4)

	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
			r, g, b, a := img.At(i, j).RGBA()
			pixels = append(pixels, uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8))
		}
	}
	return pixels
}

// drawDetection draws the detected faces and eyes.
func (c *Canvas) drawDetection(dets [][]int) error {
	c.processor.MaxPoints = c.trianglePoints
	c.processor.Grayscale = c.isGrayScaled
	c.processor.StrokeWidth = c.strokeWidth
	c.processor.PointsThreshold = c.pointsThreshold
	c.processor.Wireframe = c.wireframe

	c.triangle = &triangle.Image{*c.processor}

	for _, det := range dets {
		det := det
		c.g.Go(func() error {
			if det[3] > 50 {
				c.ctx.Call("beginPath")
				c.ctx.Set("lineWidth", 2)
				c.ctx.Set("strokeStyle", "rgba(255, 0, 0, 0.5)")

				row, col, scale := det[1], det[0], det[2]
				// Substract the image under the detected face region.
				imgData := make([]byte, scale*scale*4)
				subimg := c.ctx.Call("getImageData", row-scale/2, col-scale/2, scale, scale).Get("data")
				uint8Arr := js.Global().Get("Uint8Array").New(subimg)
				js.CopyBytesToGo(imgData, uint8Arr)

				unionMask := image.NewNRGBA(image.Rect(0, 0, scale, scale))
				// Add to union mask
				ellipse := &ellipse.Ellipse{
					Cx: row,
					Cy: col,
					Ry: int(float64(scale) * 0.8 / 2),
					Rx: int(float64(scale) * 0.8 / 1.6),
				}
				// Draw the ellipse mask.
				c.lock.Lock()
				draw.Draw(unionMask, unionMask.Bounds(), ellipse, image.Point{X: row - scale/2, Y: col - scale/2}, draw.Over)
				c.lock.Unlock()

				// Triangulate the detected face region.
				buffer, err := c.triangulate(unionMask, imgData, scale)
				if err != nil {
					return err
				}
				uint8Arr = js.Global().Get("Uint8Array").New(scale * scale * 4)
				js.CopyBytesToJS(uint8Arr, buffer)

				uint8Clamped := js.Global().Get("Uint8ClampedArray").New(uint8Arr)
				rawData := js.Global().Get("ImageData").New(uint8Clamped, scale)

				// Replace the underlying face region with the triangulated image.
				c.ctx.Call("putImageData", rawData, row-scale/2, col-scale/2)

				if c.showFrame {
					c.ctx.Call("rect", row-scale/2, col-scale/2, scale, scale)
					c.ctx.Call("stroke")
				}
			}
			return nil
		})
	}
	if err := c.g.Wait(); err != nil {
		return err
	}
	return nil
}

// triangulate triangulates the detected face region
func (c *Canvas) triangulate(unionMask *image.NRGBA, data []uint8, scale int) ([]uint8, error) {
	faceTemplate := image.NewNRGBA(image.Rect(0, 0, scale, scale))
	// Converts the buffer array to an image.
	img := c.pixToImage(data, scale)

	// Create a new image and draw the webcam frame captures into it.
	newImg := image.NewNRGBA(image.Rect(0, 0, scale, scale))
	draw.Draw(newImg, newImg.Bounds(), img, newImg.Bounds().Min, draw.Over)

	// Call the face triangulation algorithm.
	triangled, _, _, err := c.triangle.Draw(newImg, *c.processor, func() {})
	if err != nil {
		return nil, err
	}

	// Paste triangled image into the face template.
	c.lock.Lock()
	draw.Draw(faceTemplate, img.Bounds(), triangled, image.Point{}, draw.Over)
	c.lock.Unlock()

	// Draw the triangled image through the facemask and on top of the source.
	draw.DrawMask(img.(draw.Image), img.Bounds(), faceTemplate, image.Point{}, unionMask, image.Point{}, draw.Over)

	return c.imgToPix(img), nil
}

// detectKeyPress listen for the keypress event and retrieves the key code.
func (c *Canvas) detectKeyPress() {
	keyEventHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		keyCode := args[0].Get("key")
		switch {
		case keyCode.String() == "f":
			c.showFrame = !c.showFrame
		case keyCode.String() == "g":
			c.isGrayScaled = !c.isGrayScaled
		case keyCode.String() == "-":
			if c.trianglePoints > minTrianglePoints {
				c.trianglePoints -= 20
			}
		case keyCode.String() == "=":
			if c.trianglePoints <= maxTrianglePoints {
				c.trianglePoints += 20
			}
		case keyCode.String() == "[":
			if c.pointsThreshold > minPointsThreshold {
				c.pointsThreshold -= 2
			}
		case keyCode.String() == "]":
			if c.pointsThreshold <= maxPointsThreshold {
				c.pointsThreshold += 2
			}
		case keyCode.String() == "0":
			if c.pointRate <= maxPointRate {
				c.pointRate += 0.005
			}
		case keyCode.String() == "9":
			if c.pointRate > minPointRate {
				c.pointRate -= 0.005
			}
		case keyCode.String() == "1":
			if c.strokeWidth > minStrokeWidth {
				c.strokeWidth--
			}
			if c.strokeWidth == minStrokeWidth {
				c.wireframe = triangle.WithoutWireframe
			}
		case keyCode.String() == "2":
			c.wireframe = triangle.WithWireframe
			if c.strokeWidth <= maxStrokeWidth {
				c.strokeWidth++
			}
		}
		return nil
	})
	c.doc.Call("addEventListener", "keypress", keyEventHandler)
}

// Log calls the `console.log` Javascript function
func (c *Canvas) Log(args ...interface{}) {
	c.window.Get("console").Call("log", args...)
}

// Alert calls the `alert` Javascript function
func (c *Canvas) Alert(args ...interface{}) {
	alert := c.window.Get("alert")
	alert.Invoke(args...)
}
