//go:build js && wasm

package main

import "github.com/esimov/pigo-wasm-demos/masquerade"

func main() {
	c := masquerade.NewCanvas()
	webcam, err := c.StartWebcam()
	if err != nil {
		c.Alert("Webcam not detected!")
	} else {
		webcam.Render()
	}
}
