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
	}
	window, err := opengl.NewWindow(windowConfig)
	if err != nil {
		panic(err)
	}

	bus := nes.NewSysBus()
	cpu := nes.NewCpu(bus)
	bus.SetCpu(cpu)
	ppu := nes.NewPpu(bus)

	canvas := opengl.NewCanvas(pixel.R(0, 0, nes.FrameWidth, nes.FrameHeight))

	for !window.Closed() {
		canvas.SetPixels(ppu.FrameBuffer)

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
