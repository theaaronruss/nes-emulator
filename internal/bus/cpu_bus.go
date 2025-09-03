package bus

import "github.com/theaaronruss/nes-emulator/internal/cartridge"

const (
	memorySize        uint16 = 2048
	memoryStart       uint16 = 0x0000
	ppuRegistersStart uint16 = 0x2000
	cartRomStart      uint16 = 0x8000
)

type CpuBus struct {
	memory    [memorySize]uint8
	cartridge *cartridge.Cartridge
}

func NewCpuBus() *CpuBus {
	return &CpuBus{}
}

func (b *CpuBus) Read(address uint16) uint8 {
	if address >= memoryStart && address < ppuRegistersStart {
		return b.memory[address%memorySize]
	} else if address >= cartRomStart && b.cartridge != nil {
		address := (address - cartRomStart) % uint16(cartridge.PrgRomBankSize)
		return b.cartridge.Read(0, address)
	}
	return 0x00
}

func (b *CpuBus) Write(address uint16, data uint8) {
	if address >= memoryStart && address < ppuRegistersStart {
		b.memory[address%memorySize] = data
	}
}

func (b *CpuBus) InsertCartridge(cartridge *cartridge.Cartridge) {
	b.cartridge = cartridge
}
