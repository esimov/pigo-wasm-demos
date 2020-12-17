package facemask

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"syscall/js"
	"time"

	"github.com/esimov/pigo-wasm-demos/detector"
	"github.com/esimov/triangle"
)

// Canvas struct holds the Javascript objects needed for the Canvas creation
type Canvas struct {
	done   chan struct{}
	succCh chan struct{}
	errCh  chan error

	// DOM elements
	window      js.Value
	doc         js.Value
	body        js.Value
	snapshotBtn js.Value
	windowSize  struct{ width, height int }

	// Canvas properties
	webcamCanvas js.Value
	maskCanvas   js.Value
	ctx          js.Value
	ctx2         js.Value
	reqID        js.Value
	renderer     js.Func

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
	strokeWidth     float64
}

const (
	minTrianglePoints = 50
	maxTrianglePoints = 800

	minPointsThreshold = 2
	maxPointsThreshold = 50

	minStrokeWidth = 0
	maxStrokeWidth = 4
)

var (
	det  *detector.Detector
	mask js.Value

	maskWidth  int
	maskHeight int
)

// NewCanvas creates and initializes the new Canvas element
func NewCanvas() *Canvas {
	var c Canvas
	c.window = js.Global()
	c.doc = c.window.Get("document")
	c.body = c.doc.Get("body")

	c.windowSize.width = 640
	c.windowSize.height = 480

	wrapper := c.doc.Call("createElement", "div")
	wrapper.Set("id", "wrapper")
	c.body.Call("appendChild", wrapper)

	c.webcamCanvas = c.doc.Call("createElement", "canvas")
	c.webcamCanvas.Set("width", c.windowSize.width)
	c.webcamCanvas.Set("height", c.windowSize.height)
	c.webcamCanvas.Set("id", "canvas")

	c.maskCanvas = c.doc.Call("createElement", "canvas")
	c.maskCanvas.Set("width", c.windowSize.width)
	c.maskCanvas.Set("height", c.windowSize.height)
	c.maskCanvas.Set("id", "canvas2")

	wrapper.Call("appendChild", c.webcamCanvas)
	wrapper.Call("appendChild", c.maskCanvas)

	c.snapshotBtn = c.doc.Call("createElement", "div")
	c.snapshotBtn.Set("id", "snapshot")
	c.body.Call("appendChild", c.snapshotBtn)

	c.ctx = c.webcamCanvas.Call("getContext", "2d")
	c.ctx2 = c.maskCanvas.Call("getContext", "2d")

	c.showFrame = false
	c.isSolid = false
	c.isGrayScaled = false

	c.wireframe = 0
	c.strokeWidth = 0
	c.trianglePoints = 200
	c.pointsThreshold = 20

	det = detector.NewDetector()

	c.processor = &triangle.Processor{
		BlurRadius:      4,
		SobelThreshold:  10,
		Noise:           0,
		MaxPoints:       c.trianglePoints,
		PointsThreshold: c.pointsThreshold,
		Wireframe:       c.wireframe,
		StrokeWidth:     c.strokeWidth,
		IsSolid:         c.isSolid,
		Grayscale:       c.isGrayScaled,
		BackgroundColor: "#ffffff00",
	}
	c.triangle = &triangle.Image{*c.processor}
	return &c
}

// Render calls the `requestAnimationFrame` Javascript function in asynchronous mode.
func (c *Canvas) Render() error {
	var data = make([]byte, c.windowSize.width*c.windowSize.height*4)
	c.done = make(chan struct{})

	img := c.loadImage("/images/surgical-mask.png")
	mask = js.Global().Call("eval", "new Image()")
	mask.Set("src", "data:image/png;base64,"+img)

	maskWidth = js.ValueOf(mask.Get("naturalWidth")).Int()
	maskHeight = js.ValueOf(mask.Get("naturalHeight")).Int()

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
			c.ctx2.Call("drawImage", c.video, 0, 0)
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
			r, g, b, a := img.At(i, j).RGBA()
			pixels = append(pixels, uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8))
		}
	}
	return pixels
}

// pixelateDetectedRegion pixelates the detected face region
func (c *Canvas) triangulateDetectedRegion(data []uint8, dets []int) []uint8 {
	// Converts the array buffer to an image
	img := c.pixToImage(data, int(float64(dets[2])))

	res, _, _, err := c.triangle.Draw(img, nil, func() {})
	if err != nil {
		log.Fatal(err.Error())
	}
	return c.imgToPix(res)
}

// drawDetection draws the detected faces and eyes.
func (c *Canvas) drawDetection(data []uint8, dets [][]int) {
	c.processor.MaxPoints = c.trianglePoints
	c.processor.Grayscale = c.isGrayScaled
	c.processor.StrokeWidth = c.strokeWidth
	c.processor.PointsThreshold = c.pointsThreshold
	c.processor.Wireframe = c.wireframe

	c.triangle = &triangle.Image{*c.processor}

	var imgScale float64

	for i := 0; i < len(dets); i++ {
		if dets[i][3] > 50 {
			c.ctx.Call("beginPath")
			c.ctx.Set("lineWidth", 2)
			c.ctx.Set("strokeStyle", "rgba(255, 0, 0, 0.5)")

			row, col, scale := dets[i][1], dets[i][0], dets[i][2]
			row = row + int(float64(row)*0.02)
			col = col + int(float64(col)*0.2)

			leftPupil := det.DetectLeftPupil(dets[i])
			rightPupil := det.DetectRightPupil(dets[i])

			if leftPupil != nil && rightPupil != nil {
				points := det.DetectMouthPoints(leftPupil, rightPupil)
				p1, p2 := points[0], points[1]

				// Calculate the lean angle between the two mouth points.
				angle := 1 - (math.Atan2(float64(p2[0]-p1[0]), float64(p2[1]-p1[1])) * 180 / math.Pi / 90)
				if math.Abs(angle) > 0.1 {
					c.snapshotBtn.Set("style", "display:none")
				} else {
					c.snapshotBtn.Set("style", "display:block")
				}

				if scale < maskWidth || scale < maskHeight {
					if maskHeight > maskWidth {
						imgScale = float64(scale) / float64(maskHeight)
					} else {
						imgScale = float64(scale) / float64(maskWidth)
					}
				}
				width, height := float64(maskWidth)*imgScale*0.7, float64(maskHeight)*imgScale*0.7
				tx := row - int(width/2)
				ty := p1[1] + (p1[1]-p2[1])/2 - int(height*0.5)

				c.ctx.Call("save")
				c.ctx.Call("translate", js.ValueOf(tx).Int(), js.ValueOf(ty).Int())
				c.ctx.Call("rotate", js.ValueOf(angle).Float())

				c.ctx.Set("globalCompositeOperation", "destination-atop")
				c.ctx.Call("drawImage", mask,
					js.ValueOf(0).Int(), js.ValueOf(0).Int(),
					js.ValueOf(width).Int(), js.ValueOf(height).Int(),
				)

				// Substract the image under the detected face region.
				imgData := make([]byte, scale*scale*4)
				subimg := c.ctx.Call("getImageData", row-scale/2, col-scale/2, scale, scale).Get("data")
				uint8Arr := js.Global().Get("Uint8Array").New(subimg)
				js.CopyBytesToGo(imgData, uint8Arr)

				buffer := c.triangulateDetectedRegion(imgData, dets[i])
				uint8Arr = js.Global().Get("Uint8Array").New(scale * scale * 4)
				js.CopyBytesToJS(uint8Arr, buffer)

				uint8Clamped := js.Global().Get("Uint8ClampedArray").New(uint8Arr)
				rawData := js.Global().Get("ImageData").New(uint8Clamped, scale)

				// Replace the underlying face region with the triangulated image.
				c.ctx.Call("putImageData", rawData, row-scale/2, col-scale/2)
				c.ctx.Call("restore")
			}

			if c.showFrame {
				c.ctx.Call("rect", row-scale/2, col-scale/2, scale, scale)
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
				c.pointsThreshold -= 5
			}
		case keyCode.String() == "]":
			if c.pointsThreshold <= maxPointsThreshold {
				c.pointsThreshold += 5
			}
		case keyCode.String() == "1":
			if c.strokeWidth > minStrokeWidth {
				c.strokeWidth--
			}
			if c.strokeWidth == minStrokeWidth {
				c.wireframe = 0
			}
		case keyCode.String() == "2":
			c.wireframe = 1
			if c.strokeWidth <= maxStrokeWidth {
				c.strokeWidth++
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

// loadImage load the source image and encodes it to base64 format.
func (c *Canvas) loadImage(path string) string {
	href := js.Global().Get("location").Get("href")
	u, err := url.Parse(href.String())
	if err != nil {
		log.Fatal(err)
	}
	u.Path = path
	u.RawQuery = fmt.Sprint(time.Now().UnixNano())

	resp, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}
