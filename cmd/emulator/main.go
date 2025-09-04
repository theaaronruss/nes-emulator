package main

import (
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

func main() {
	opengl.Run(run)
}

func run() {
	cfg := opengl.WindowConfig{
		Title:  "NES Emulator",
		Bounds: pixel.R(0, 0, 512, 480),
	}
	win, err := opengl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	canvas := opengl.NewCanvas(pixel.R(0, 0, 256, 240))
	pixels := make([]uint8, 4*256*240)
	for i := 0; i < cap(pixels); i += 4 {
		pixels[i] = 0xFF
		pixels[i+1] = 0x00
		pixels[i+2] = 0xFF
		pixels[i+3] = 0xFF
	}
	canvas.SetPixels(pixels)
	for !win.Closed() {
		transformationMatrix := pixel.IM.Scaled(pixel.Vec{}, 2)
		transformationMatrix = transformationMatrix.Moved(win.Bounds().Center())
		canvas.Draw(win, transformationMatrix)
		win.Update()
	}
}
