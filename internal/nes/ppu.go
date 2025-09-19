package nes

const (
	FrameWidth  float64 = 256
	FrameHeight float64 = 240
)

type ppu struct {
	frameBuffer []uint8
}

func NewPpu(sys *System) *ppu {
	return &ppu{
		frameBuffer: make([]uint8, int(FrameWidth*FrameHeight*4)),
	}
}

func (p *ppu) WritePpuCtrl(addr uint16, data uint8) {
}

func (p *ppu) WritePpuMask(addr uint16, data uint8) {
}

func (p *ppu) ReadPpuStatus(addr uint16) uint8 {
	return 0
}

func (p *ppu) WriteOamAddr(addr uint16, data uint8) {
}

func (p *ppu) ReadOamData(addr uint16) uint8 {
	return 0
}

func (p *ppu) WriteOamData(addr uint16, data uint8) {
}

func (p *ppu) WritePpuScroll(addr uint16, data uint8) {
}

func (p *ppu) WritePpuAddr(addr uint16, data uint8) {
}

func (p *ppu) ReadPpuData(addr uint16) uint8 {
	return 0
}

func (p *ppu) WritePpuData(addr uint16, data uint8) {
}
