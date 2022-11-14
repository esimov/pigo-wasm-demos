//go:build js && wasm

package main

import (
	"fmt"

	"github.com/esimov/pigo-wasm-demos/wasm"
)

func main() {
	c := wasm.NewCanvas()
	webcam, err := c.StartWebcam()
	if err != nil {
		c.Alert("Webcam not detected!")
	} else {
		err := webcam.Render()
		if err != nil {
			c.Log(fmt.Sprint(err))
		}
	}
}
