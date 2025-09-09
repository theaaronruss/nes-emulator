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
	actual := cpu.getZeroPageAddress()
	if actual != 0x69 {
		t.Errorf("expected address 0x69, got 0x%X", actual)
	}
}

func TestGetAbsoluteAddress(t *testing.T) {
	bus := newFakeSysBus()
	bus.data[0xFFFC] = 0x00
	bus.data[0xFFFD] = 0x06
	bus.data[0x0601] = 0x25
	bus.data[0x0602] = 0xC3
	cpu := NewCpu(bus)
	actual := cpu.getAbsoluteAddress()
	if actual != 0xC325 {
		t.Errorf("expected address 0xC325, got 0x%X", actual)
	}
}

func TestGetRelativeAddress(t *testing.T) {
	tests := []struct {
		name        string
		pc          uint16
		input       uint8
		expected    uint16
		pageCrossed bool
	}{
		{
			name: "basic jump forward",
			pc:   0x0630, input: 0x25, expected: 0x0657, pageCrossed: false,
		},
		{
			name: "basic jump backward",
			pc:   0x0660, input: 0xAC, expected: 0x060E, pageCrossed: false,
		},
		{
			name: "jump forward wrap",
			pc:   0xFFF0, input: 0x3C, expected: 0x002E, pageCrossed: true,
		},
		{
			name: "jump backward wrap",
			pc:   0x000F, input: 0xC8, expected: 0xFFD9, pageCrossed: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bus := newFakeSysBus()
			bus.data[0xFFFC] = uint8(test.pc & 0x00FF)
			bus.data[0xFFFD] = uint8(test.pc & 0xFF00 >> 8)
			bus.data[test.pc+1] = test.input
			cpu := NewCpu(bus)
			actual, actualPageCrossed := cpu.getRelativeAddress()
			if actual != test.expected {
				t.Errorf("%s: expected address 0x%X, got 0x%X", test.name,
					test.expected, actual)
			}
			if actualPageCrossed != test.pageCrossed {
				t.Errorf("%s: expected page crossed to be %t, got %t", test.name,
					test.pageCrossed, actualPageCrossed)
			}
		})
	}
}
