package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/theaaronruss/nes-emulator/internal/cartridge"
	"github.com/theaaronruss/nes-emulator/internal/cpu"
	"github.com/theaaronruss/nes-emulator/internal/ppu"
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
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: emulator romFile")
		return
	}

	err := cartridge.LoadCartridge(os.Args[1])
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
	canvas := opengl.NewCanvas(pixel.R(0, 0,
		float64(ppu.FrameWidth), float64(ppu.FrameHeight)))
	for !win.Closed() {
		startTime := time.Now()
		for !ppu.IsFrameComplete {
			ppu.Clock()
			if ppu.HandleVblank {
				cpu.Nmi()
				ppu.HandleVblank = false
			}
			clockCycle++
			if clockCycle >= 3 {
				cpu.Clock()
				clockCycle = 0
			}
		}
		canvas.SetPixels(ppu.FrameBuffer)
		ppu.IsFrameComplete = false

		transformationMatrix := pixel.IM.ScaledXY(
			pixel.Vec{},
			pixel.Vec{X: 2, Y: -2},
		)
		transformationMatrix = transformationMatrix.Moved(win.Bounds().Center())
		canvas.Draw(win, transformationMatrix)
		win.Update()
		elapsedTime := time.Since(startTime).Milliseconds()
		sleepTime := frameTime - elapsedTime
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	}
}
