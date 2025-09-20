package main

import (
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/theaaronruss/nes-emulator/internal/nes"
)

func run() {
	windowConfig := opengl.WindowConfig{
		Title:  "NES Emulator",
		Bounds: pixel.R(0, 0, nes.FrameWidth*2, nes.FrameHeight*2),
		VSync:  false,
	}
	window, err := opengl.NewWindow(windowConfig)
	if err != nil {
		panic(err)
	}

	cartridge, _ := nes.NewCartridge("nestest.nes")
	system := nes.NewSystem(cartridge)

	for {
		system.Clock()
	}

	canvas := opengl.NewCanvas(pixel.R(0, 0, nes.FrameWidth, nes.FrameHeight))

	for !window.Closed() {
		// system.Clock()

		canvas.SetPixels(system.FrameBuffer())

		transMatrix := pixel.IM
		transMatrix = transMatrix.ScaledXY(
			pixel.Vec{},
			pixel.Vec{X: 2, Y: -2},
		)
		transMatrix = transMatrix.Moved(window.Bounds().Center())
		canvas.Draw(window, transMatrix)

		window.Update()
	}
}

func main() {
	opengl.Run(run)
}
