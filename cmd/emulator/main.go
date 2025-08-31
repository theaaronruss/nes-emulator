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
		Bounds: pixel.R(0, 0, 512, 448),
		VSync:  true,
	}
	win, err := opengl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	width := int(win.Canvas().Bounds().W())
	height := int(win.Canvas().Bounds().H())
	pixels := make([]uint8, 4*width*height)
	for i := 0; i < cap(pixels); i += 4 {
		pixels[i] = 0xFF
		pixels[i+1] = 0x00
		pixels[i+2] = 0xFF
		pixels[i+3] = 0xFF
	}
	win.Canvas().SetPixels(pixels)
	for !win.Closed() {
		win.Update()
	}
}
