package main

import (
	"fmt"
	"time"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/theaaronruss/nes-emulator/internal/cartridge"
	"github.com/theaaronruss/nes-emulator/internal/cpu"
	"github.com/theaaronruss/nes-emulator/internal/ppu"
	"github.com/theaaronruss/nes-emulator/internal/sysbus"
)

const (
	frameWidth  float64 = 256
	frameHeight float64 = 240
	frameTime   int64   = 1000 / 60
)

func main() {
	opengl.Run(run)
}

func run() {
	var err error
	sysbus.GamePak, err = cartridge.LoadCartridge("nestest.nes")
	cpu.Reset()
	if err != nil {
		fmt.Println("Failed to load ROM file")
		return
	}
	clockCycle := 0

	cfg := opengl.WindowConfig{
		Title:  "NES Emulator",
		Bounds: pixel.R(0, 0, frameWidth*2, frameHeight*2),
	}
	win, err := opengl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	canvas := opengl.NewCanvas(pixel.R(0, 0, 256, 240))
	for !win.Closed() {
		startTime := time.Now()
		for !ppu.FrameComplete {
			ppu.Clock()
			clockCycle++
			if clockCycle >= 3 {
				cpu.Clock()
				clockCycle = 0
			}
		}
		canvas.SetPixels(ppu.FrameBuffer)
		ppu.FrameComplete = false

		transformationMatrix := pixel.IM.Scaled(pixel.Vec{}, 2)
		transformationMatrix = transformationMatrix.Moved(win.Bounds().Center())
		canvas.Draw(win, transformationMatrix)
		win.Update()
		elapsedTime := time.Since(startTime).Milliseconds()
		sleepTime := frameTime - elapsedTime
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	}
}
