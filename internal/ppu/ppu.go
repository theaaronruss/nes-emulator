package ppu

import "math/rand"

const (
	FrameWidth  int = 256
	FrameHeight int = 240
)

var (
	FrameBuffer   []uint8 = make([]uint8, 4*FrameWidth*FrameHeight)
	FrameComplete bool
	pixelOffset   int
)

func Clock() {
	var color uint8
	if rand.Intn(2) == 1 {
		color = 0xFF
	} else {
		color = 0x00
	}
	FrameBuffer[pixelOffset*4] = color
	FrameBuffer[pixelOffset*4+1] = color
	FrameBuffer[pixelOffset*4+2] = color
	FrameBuffer[pixelOffset*4+3] = 0xFF
	pixelOffset++
	if pixelOffset >= FrameWidth*FrameHeight {
		pixelOffset = 0
		FrameComplete = true
	}
}
