package pixelate

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"syscall/js"

	"github.com/esimov/pigo-wasm-demos/detector"
)

// Canvas struct holds the Javascript objects needed for the Canvas creation
type Canvas struct {
	done   chan struct{}
	succCh chan struct{}
	errCh  chan error

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

	// Canvas interaction related variables
	showPupil bool
	showFrame bool
	useNoise  bool

	// Quantizer related variables
	numOfColors int
	cellSize    int

	frame *image.NRGBA
}

// SubImager is a wrapper implementing the SubImage method from the image package.
type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

const (
	minColors   = 2
	maxColors   = 32
	minCellSize = 8
	maxCellSize = 30
)

var (
	det   *detector.Detector
	quant Quant
)

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
	c.showPupil = false
	c.showFrame = false
	c.useNoise = false

	c.numOfColors = 8
	c.cellSize = 10

	det = detector.NewDetector()
	quant = Quant{}

	return &c
}

// Render calls the `requestAnimationFrame` Javascript function in asynchronous mode.
func (c *Canvas) Render() error {
	var data = make([]byte, c.windowSize.width*c.windowSize.height*4)
	c.done = make(chan struct{})

	err := det.UnpackCascades()
	if err != nil {
		return err
	}
	c.renderer = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			c.window.Get("stats").Call("begin")

			width, height := c.windowSize.width, c.windowSize.height
			c.reqID = c.window.Call("requestAnimationFrame", c.renderer)
			// Draw the webcam frame to the canvas element
			c.ctx.Call("drawImage", c.video, 0, 0)
			rgba := c.ctx.Call("getImageData", 0, 0, width, height).Get("data")

			// Convert the rgba value of type Uint8ClampedArray to Uint8Array in order to
			// be able to transfer it from Javascript to Go via the js.CopyBytesToGo function.
			uint8Arr := js.Global().Get("Uint8Array").New(rgba)
			js.CopyBytesToGo(data, uint8Arr)
			gray := c.rgbaToGrayscale(data)
			res := det.DetectFaces(gray, height, width)
			c.drawDetection(data, res)

			c.window.Get("stats").Call("end")
		}()
		return nil
	})
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
	col := color.NRGBA{
		R: uint8(0),
		G: uint8(0),
		B: uint8(0),
		A: uint8(255),
	}

	for y := bounds.Min.Y; y < dy; y++ {
		for x := bounds.Min.X; x < dx*4; x += 4 {
			col.R = uint8(pixels[x+y*dx*4])
			col.G = uint8(pixels[x+y*dx*4+1])
			col.B = uint8(pixels[x+y*dx*4+2])
			col.A = uint8(pixels[x+y*dx*4+3])

			c.frame.SetNRGBA(y, int(x/4), col)
		}
	}
	return c.frame
}

// imgToPix converts an image to an array buffer
func (c *Canvas) imgToPix(img image.Image) []uint8 {
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

// pixelateDetectedRegion pixelates the detected face region
func (c *Canvas) pixelateDetectedRegion(data []uint8, dets []int, useNoise bool) []uint8 {
	// Converts the array buffer to an image
	img := c.pixToImage(data, int(float64(dets[2])*0.75))

	// Quantize the substracted image in order to reduce the number of colors.
	// This will results in a pixelated subtype image.
	cell := quant.Draw(img, c.numOfColors, c.cellSize, useNoise)
	return c.imgToPix(cell)
}

// drawDetection draws the detected faces and eyes.
func (c *Canvas) drawDetection(data []uint8, dets [][]int) {
	for i := 0; i < len(dets); i++ {
		if dets[i][3] > 50 {
			c.ctx.Call("beginPath")
			c.ctx.Set("lineWidth", 2)
			c.ctx.Set("strokeStyle", "rgba(255, 0, 0, 0.5)")

			row, col, scale := dets[i][1], dets[i][0], dets[i][2]
			col = col + int(float64(col)*0.115)
			scale = int(float64(scale) * 0.75)

			if c.showFrame {
				c.ctx.Call("rect", row-scale/2, col-scale/2, scale, scale)
			}
			// Substract the image under the detected face region.
			imgData := make([]byte, scale*scale*4)
			subimg := c.ctx.Call("getImageData", row-scale/2, col-scale/2, scale, scale).Get("data")
			uint8Arr := js.Global().Get("Uint8Array").New(subimg)
			js.CopyBytesToGo(imgData, uint8Arr)

			buffer := c.pixelateDetectedRegion(imgData, dets[i], c.useNoise)
			uint8Arr = js.Global().Get("Uint8Array").New(scale * scale * 4)
			js.CopyBytesToJS(uint8Arr, buffer)

			uint8Clamped := js.Global().Get("Uint8ClampedArray").New(uint8Arr)
			rawData := js.Global().Get("ImageData").New(uint8Clamped, scale)

			// Replace the underlying face region byte array with the quantized values.
			c.ctx.Call("putImageData", rawData, row-scale/2, col-scale/2)
			c.ctx.Call("stroke")

			if c.showPupil {
				leftPupil := det.DetectLeftPupil(dets[i])
				if leftPupil != nil {
					col, row, scale := leftPupil.Col, leftPupil.Row, leftPupil.Scale/8
					c.ctx.Call("moveTo", col+int(scale), row)
					c.ctx.Call("arc", col, row, scale, 0, 2*math.Pi, true)
				}

				rightPupil := det.DetectRightPupil(dets[i])
				if rightPupil != nil {
					col, row, scale := rightPupil.Col, rightPupil.Row, rightPupil.Scale/8
					c.ctx.Call("moveTo", col+int(scale), row)
					c.ctx.Call("arc", col, row, scale, 0, 2*math.Pi, true)
				}
				c.ctx.Call("stroke")
			}
		}
	}
}

// detectKeyPress listen for the keypress event and retrieves the key code.
func (c *Canvas) detectKeyPress() {
	keyEventHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		keyCode := args[0].Get("key")
		switch {
		case keyCode.String() == "s":
			c.showPupil = !c.showPupil
		case keyCode.String() == "f":
			c.showFrame = !c.showFrame
		case keyCode.String() == "n":
			c.useNoise = !c.useNoise
		case keyCode.String() == "=":
			if c.numOfColors <= maxColors {
				c.numOfColors++
			}
		case keyCode.String() == "-":
			if c.numOfColors > minColors {
				c.numOfColors--
			}
		case keyCode.String() == "]":
			if c.cellSize <= maxCellSize {
				c.cellSize++
			}
		case keyCode.String() == "[":
			if c.cellSize > minCellSize {
				c.cellSize--
			}
		default:
			c.showFrame = false
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
