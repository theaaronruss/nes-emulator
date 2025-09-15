package main

import (
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

const (
	frameWidth  float64 = 256
	frameHeight float64 = 240
)

func run() {
	windowConfig := opengl.WindowConfig{
		Title:  "NES Emulator",
		Bounds: pixel.R(0, 0, frameWidth*2, frameHeight*2),
	}
	window, err := opengl.NewWindow(windowConfig)
	if err != nil {
		panic(err)
	}

	canvas := opengl.NewCanvas(pixel.R(0, 0, frameWidth, frameHeight))
	pixelCount := int(frameWidth) * int(frameHeight)
	frameBuffer := make([]uint8, 4*pixelCount)
	for i := 0; i < 4*pixelCount; i += 4 {
		frameBuffer[i] = 0x00
		frameBuffer[i+1] = 0x00
		frameBuffer[i+2] = 0x00
		frameBuffer[i+3] = 0xFF
	}
	canvas.SetPixels(frameBuffer)

	for !window.Closed() {
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

	// cartridge, err := nes.NewCartridge("nestest.nes")
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, err.Error())
	// 	return
	// }
	// bus := nes.NewSysBus()
	// bus.InsertCartridge(cartridge)
	// cpu := nes.NewCpu(bus)
	// for {
	// 	cpu.Clock()
	// }
}
