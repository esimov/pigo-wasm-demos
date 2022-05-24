package faceblur

import (
	"fmt"
	"image"
	"math"
	"syscall/js"

	"github.com/esimov/pigo-wasm-demos/detector"
	"github.com/esimov/pigo-wasm-demos/pixels"
	"github.com/esimov/stackblur-go"
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
	canvas    js.Value
	ellipse   js.Value
	offscreen js.Value
	ctx       js.Value
	ctxMask   js.Value
	ctxOffscr js.Value
	reqID     js.Value
	renderer  js.Func

	// Webcam properties
	navigator js.Value
	video     js.Value

	// Canvas interaction related variables
	showPupil  bool
	showFrame  bool
	isBlurred  bool
	blurRadius uint32

	frame *image.NRGBA
}

const (
	minBlurRadius = 5
	maxBlurRadius = 50
)

var pigo *detector.Detector

// NewCanvas creates and initializes the new Canvas element
func NewCanvas() *Canvas {
	var c Canvas
	c.window = js.Global()
	c.doc = c.window.Get("document")
	c.body = c.doc.Get("body")

	c.windowSize.width = 768
	c.windowSize.height = 576

	c.canvas = c.doc.Call("createElement", "canvas")
	c.ellipse = c.doc.Call("createElement", "canvas")
	c.offscreen = c.doc.Call("createElement", "canvas")

	c.canvas.Set("width", c.windowSize.width)
	c.canvas.Set("height", c.windowSize.height)
	c.canvas.Set("id", "canvas")
	c.body.Call("appendChild", c.canvas)

	c.ellipse.Set("width", c.windowSize.width)
	c.ellipse.Set("height", c.windowSize.height)

	c.offscreen.Set("width", c.windowSize.width)
	c.offscreen.Set("height", c.windowSize.height)

	c.ctx = c.canvas.Call("getContext", "2d")
	c.ctxMask = c.ellipse.Call("getContext", "2d")
	c.ctxOffscr = c.offscreen.Call("getContext", "2d")

	c.showPupil = false
	c.showFrame = false
	c.isBlurred = true
	c.blurRadius = 20

	pigo = detector.NewDetector()
	return &c
}

// Render calls the `requestAnimationFrame` Javascript function in asynchronous mode.
func (c *Canvas) Render() error {
	width, height := c.windowSize.width, c.windowSize.height
	var data = make([]byte, width*height*4)
	c.done = make(chan struct{})

	err := pigo.UnpackCascades()
	if err != nil {
		return err
	}
	c.renderer = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() error {
			c.window.Get("stats").Call("begin")

			c.reqID = c.window.Call("requestAnimationFrame", c.renderer)
			// Draw the webcam frame into the canvas element
			c.ctx.Call("drawImage", c.video, 0, 0)
			rgba := c.ctx.Call("getImageData", 0, 0, width, height).Get("data")

			// Convert the rgba value of type Uint8ClampedArray to Uint8Array in order to
			// be able to transfer it from Javascript to Go via the js.CopyBytesToGo function.
			uint8Arr := js.Global().Get("Uint8Array").New(rgba)
			js.CopyBytesToGo(data, uint8Arr)

			gray := pixels.RgbaToGrayscale(data, width, height)

			// Reset the data slice to its default values to avoid unnecessary memory allocation.
			// Otherwise, the GC won't clean up the memory address allocated by this slice
			// and the memory will keep increasing by each iteration.
			data = make([]byte, len(data))

			res := pigo.DetectFaces(gray, height, width)
			if len(res) > 0 {
				if err := c.drawDetection(res); err != nil {
					return err
				}
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

// blurFace blures out the detected face region
func (c *Canvas) blurFace(src image.Image) (*image.NRGBA, error) {
	img, err := stackblur.Process(src, c.blurRadius)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// drawDetection draws the detected faces and eyes.
func (c *Canvas) drawDetection(dets [][]int) error {
	var scaleX, scaleY, invScaleX, invScaleY float64
	var grad js.Value

	for _, det := range dets {
		leftPupil := pigo.DetectLeftPupil(det)
		rightPupil := pigo.DetectRightPupil(det)

		if det[3] > 50 {
			c.ctx.Call("beginPath")
			c.ctx.Set("lineWidth", 2)
			c.ctx.Set("strokeStyle", "rgba(255, 0, 0, 0.5)")

			row, col, scale := det[1], det[0], int(float64(det[2])*1.2)

			if c.isBlurred {
				// Substract the image under the detected face region.
				imgData := make([]byte, scale*scale*4)
				subimg := c.ctx.Call("getImageData", row-scale/2, col-scale/2, scale, scale).Get("data")
				uint8Arr := js.Global().Get("Uint8Array").New(subimg)
				js.CopyBytesToGo(imgData, uint8Arr)

				scx, scy := int(float64(scale)*0.8/1.6), int(float64(scale)*0.8/2.1)
				rx, ry := scx/2, scy/2

				if rx >= ry {
					scaleX, invScaleX = 1, 1
					scaleY = float64(rx) / float64(ry)
					invScaleY = float64(ry) / float64(rx)
					grad = c.ctxMask.Call("createRadialGradient", scale/2, float64(scale/2)*invScaleY, 0, scale/2, float64(scale/2)*invScaleY, scx)
				} else {
					scaleY, invScaleY = 1, 1
					scaleX = float64(ry) / float64(rx)
					invScaleX = float64(rx) / float64(ry)
					grad = c.ctxMask.Call("createRadialGradient", float64(scale/2)*invScaleX, scale/2, 0, float64(scale/2)*invScaleX, scale/2, scy)
				}

				grad.Call("addColorStop", 0.53, "rgba(0, 0, 0, 255)")
				grad.Call("addColorStop", 0.75, "rgba(255, 255, 255, 0)")

				// Clear the canvas on each frame.
				c.ctxMask.Call("clearRect", 0, 0, c.windowSize.width, c.windowSize.height)
				c.ctxMask.Call("setTransform", scaleX, 0, 0, scaleY, 0, 0)

				c.ctxMask.Set("fillStyle", grad)
				c.ctxMask.Call("fillRect", 0, 0, scale, scale)

				// Converts the buffer array to an image.
				rect := image.Rect(0, 0, scale, scale)
				img := pixels.PixToImage(imgData, rect)

				// Blur out the image.
				blurred, err := c.blurFace(img)
				if err != nil {
					return err
				}

				uint8Arr = js.Global().Get("Uint8Array").New(scale * scale * 4)
				js.CopyBytesToJS(uint8Arr, pixels.ImgToPix(blurred))

				uint8Clamped := js.Global().Get("Uint8ClampedArray").New(uint8Arr)
				rawData := js.Global().Get("ImageData").New(uint8Clamped, scale)

				// Clear out the canvas on each frame.
				c.ctxOffscr.Call("clearRect", 0, 0, c.windowSize.width, c.windowSize.height)
				// Replace the underlying face region with the blurred image.
				c.ctxOffscr.Call("putImageData", rawData, 0, 0)

				// Calculate the lean angle between the eyes.
				angle := 1 - (math.Atan2(float64(rightPupil.Col-leftPupil.Col), float64(rightPupil.Row-leftPupil.Row)) * 180 / math.Pi / 90)

				c.ctxOffscr.Call("save")
				c.ctxOffscr.Call("translate", scale/2, scale/2)
				c.ctxOffscr.Call("rotate", js.ValueOf(angle).Float())
				c.ctxOffscr.Call("translate", -scale/2, -scale/2)

				// Apply the ellipse mask over the source image by using composite operation.
				c.ctxOffscr.Set("globalCompositeOperation", "destination-atop")
				c.ctxOffscr.Call("drawImage", c.ellipse, 0, 0)
				c.ctxOffscr.Call("restore")

				c.ctx.Call("drawImage", c.offscreen, row-scale/2, col-scale/2)
			}

			if c.showFrame {
				c.ctx.Call("rect", row-scale/2, col-scale/2, scale, scale)
				c.ctx.Call("stroke")
			}

			if c.showPupil {
				if leftPupil != nil {
					col, row, scale := leftPupil.Col, leftPupil.Row, leftPupil.Scale/8
					c.ctx.Call("moveTo", col+int(scale), row)
					c.ctx.Call("arc", col, row, scale, 0, 2*math.Pi, true)
				}

				if rightPupil != nil {
					col, row, scale := rightPupil.Col, rightPupil.Row, rightPupil.Scale/8
					c.ctx.Call("moveTo", col+int(scale), row)
					c.ctx.Call("arc", col, row, scale, 0, 2*math.Pi, true)
				}
				c.ctx.Call("stroke")
			}
		}
	}
	return nil
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
		case keyCode.String() == "b":
			c.isBlurred = !c.isBlurred
		case keyCode.String() == "]":
			if c.blurRadius <= maxBlurRadius {
				c.blurRadius++
			}
		case keyCode.String() == "[":
			if c.blurRadius > minBlurRadius {
				c.blurRadius--
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
