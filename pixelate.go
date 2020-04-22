// +build js,wasm

package main

import "github.com/esimov/pigo-wasm-demos/pixelate"

func main() {
	c := pixelate.NewCanvas()
	webcam, err := c.StartWebcam()
	if err != nil {
		c.Alert("Webcam not detected!")
	} else {
		webcam.Render()
	}
}
