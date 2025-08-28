package bus

import (
	"errors"
	"fmt"
	"os"
)

const ramSize int = 65536

type Bus struct {
	ram [ramSize]uint8
}

func NewBus() *Bus {
	return &Bus{
		ram: [ramSize]uint8{},
	}
}

func (b *Bus) Read(address uint16) uint8 {
	return b.ram[address]
}

func (b *Bus) Write(address uint16, data uint8) {
	b.ram[address] = data
}

func (b *Bus) LoadRom(filePath string) error {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to open rom file: %w", err)
	}
	newRamContents := []uint8(fileBytes)
	length := len(newRamContents)
	if length > ramSize {
		return errors.New("rom file too large")
	}
	b.ram = [ramSize]uint8{}
	copy(b.ram[:length], newRamContents)
	return nil
}
