package nes

import "testing"

type fakeSysBus struct {
	data map[uint16]uint8
}

func newFakeSysBus() *fakeSysBus {
	return &fakeSysBus{
		data: make(map[uint16]uint8),
	}
}

func (bus *fakeSysBus) Read(address uint16) uint8 {
	return bus.data[address]
}

func (bus *fakeSysBus) Write(address uint16, data uint8) {
	bus.data[address] = data
}

func TestGetZeroPageAddress(t *testing.T) {
	bus := newFakeSysBus()
	bus.data[0xFFFC] = 0x00
	bus.data[0xFFFD] = 0x06
	bus.data[0x0601] = 0x69
	cpu := NewCpu(bus)
	address := cpu.getZeroPageAddress()
	if address != 0x69 {
		t.Errorf("expected address 0x69, got 0x%X", address)
	}
}

func TestGetAbsoluteAddress(t *testing.T) {
	bus := newFakeSysBus()
	bus.data[0xFFFC] = 0x00
	bus.data[0xFFFD] = 0x06
	bus.data[0x0601] = 0x25
	bus.data[0x0602] = 0xC3
	cpu := NewCpu(bus)
	address := cpu.getAbsoluteAddress()
	if address != 0xC325 {
		t.Errorf("expected address 0xC325, got 0x%X", address)
	}
}
