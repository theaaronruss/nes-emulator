package main

import (
	"time"

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

	cartridge, err := nes.NewCartridge("roms/nestest.nes")
	if err != nil {
		panic(err.Error())
	}
	system := nes.NewSystem(window, cartridge)

	canvas := opengl.NewCanvas(pixel.R(0, 0, nes.FrameWidth, nes.FrameHeight))

	for !window.Closed() {
		start := time.Now()
		system.ClockFrame()
		elapsed := time.Since(start)
		sleepTime := (time.Second / 60) - elapsed
		if sleepTime > 0 {
			time.Sleep(sleepTime)
		}

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
