//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"

	"github.com/esimov/pigo-wasm-demos/bgblur"
)

func main() {
	c := bgblur.NewCanvas()
	webcam, err := c.StartWebcam()
	if err != nil {
		c.Alert("Webcam not detected!")
	} else {
		err := webcam.Render()
		if err != nil {
			c.Alert(fmt.Sprint(err))
		}
	}
}
