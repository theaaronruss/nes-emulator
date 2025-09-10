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

func TestGetZeroPageOffsetAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    uint8
		offset   uint8
		expected uint16
	}{
		{
			name:  "basic zero page",
			input: 0x3A, offset: 0x18, expected: 0x52,
		},
		{
			name:  "wrap around",
			input: 0xF8, offset: 0x0F, expected: 0x07,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bus := newFakeSysBus()
			bus.data[0xFFFC] = 0x00
			bus.data[0xFFFD] = 0x06
			bus.data[0x0601] = test.input
			cpu := NewCpu(bus)
			actual := cpu.getZeroPageOffsetAddress(test.offset)
			if actual != test.expected {
				t.Errorf("%s: expected address 0x%X, got 0x%X", t.Name(),
					test.expected, actual)
			}
		})
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

func TestGetAbsoluteOffsetAddress(t *testing.T) {
	tests := []struct {
		name        string
		offset      uint8
		input       uint16
		expected    uint16
		pageCrossed bool
	}{
		{
			name:        "no page cross",
			offset:      0x15,
			input:       0x8AB3,
			expected:    0x8AC8,
			pageCrossed: false,
		},
		{
			name:        "page cross",
			offset:      0x5F,
			input:       0x8AE5,
			expected:    0x8B44,
			pageCrossed: true,
		},
	}

	for _, test := range tests {
		bus := newFakeSysBus()
		bus.data[0xFFFC] = 0x00
		bus.data[0xFFFD] = 0x06
		bus.data[0x0601] = uint8(test.input & 0x00FF)
		bus.data[0x0602] = uint8(test.input & 0xFF00 >> 8)
		cpu := NewCpu(bus)
		actual, actualPageCrossed := cpu.getAbsoluteOffsetAddress(test.offset)
		if actual != test.expected {
			t.Errorf("%s: expected address 0x%X, got 0x%X", t.Name(),
				test.expected, actual)
		}
		if actualPageCrossed != test.pageCrossed {
			t.Errorf("%s: expected page crossed to be %t, got %t", t.Name(),
				test.pageCrossed, actualPageCrossed)
		}
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
				t.Errorf("%s: expected address 0x%X, got 0x%X", t.Name(),
					test.expected, actual)
			}
			if actualPageCrossed != test.pageCrossed {
				t.Errorf("%s: expected page crossed to be %t, got %t", t.Name(),
					test.pageCrossed, actualPageCrossed)
			}
		})
	}
}

func TestGetIndirectAddress(t *testing.T) {
	t.Run("basic jump", func(t *testing.T) {
		bus := newFakeSysBus()
		bus.data[0xFFFC] = 0x00
		bus.data[0xFFFD] = 0x06
		bus.data[0x0601] = 0x5A
		bus.data[0x0602] = 0x82
		bus.data[0x825A] = 0x34
		bus.data[0x825B] = 0x12
		cpu := NewCpu(bus)
		actual := cpu.getIndirectAddress()
		if actual != 0x1234 {
			t.Errorf("%s: expected address 0x1234, got 0x%X", t.Name(), actual)
		}
	})

	t.Run("boundary jump", func(t *testing.T) {
		bus := newFakeSysBus()
		bus.data[0xFFFC] = 0x00
		bus.data[0xFFFD] = 0x06
		bus.data[0x0601] = 0xFF
		bus.data[0x0602] = 0x30
		bus.data[0x30FF] = 0x34
		bus.data[0x3000] = 0x12
		cpu := NewCpu(bus)
		actual := cpu.getIndirectAddress()
		if actual != 0x1234 {
			t.Errorf("%s: expected address 0x1234, got 0x%X", t.Name(), actual)
		}
	})
}

func TestGetIndexedIndirectAddress(t *testing.T) {
	t.Run("zero page", func(t *testing.T) {
		bus := newFakeSysBus()
		bus.data[0xFFFC] = 0x00
		bus.data[0xFFFD] = 0x06
		bus.data[0x0601] = 0x80
		bus.data[0x0088] = 0x34
		bus.data[0x0089] = 0x12
		cpu := NewCpu(bus)
		actual := cpu.getIndexedIndirectAddress(0x08)
		if actual != 0x1234 {
			t.Errorf("%s: expected address 0x1234, got 0x%X", t.Name(), actual)
		}
	})

	t.Run("page cross", func(t *testing.T) {
		bus := newFakeSysBus()
		bus.data[0xFFFC] = 0x00
		bus.data[0xFFFD] = 0x06
		bus.data[0x0601] = 0xF0
		bus.data[0x00FF] = 0x34
		bus.data[0x0000] = 0x12
		cpu := NewCpu(bus)
		actual := cpu.getIndexedIndirectAddress(0x0F)
		if actual != 0x1234 {
			t.Errorf("%s: expected address 0x1234, got 0x%X", t.Name(), actual)
		}
	})
}

func TestGetIndirectIndexedAddress(t *testing.T) {
	t.Run("same page", func(t *testing.T) {
		bus := newFakeSysBus()
		bus.data[0xFFFC] = 0x00
		bus.data[0xFFFD] = 0x06
		bus.data[0x0601] = 0x4B
		bus.data[0x4B] = 0x26
		bus.data[0x4C] = 0x12
		cpu := NewCpu(bus)
		actual, pageCrossed := cpu.getIndirectIndexedAddress(0x0E)
		if actual != 0x1234 {
			t.Errorf("%s: expected address 0x1234, got 0x%X", t.Name(), actual)
		}
		if pageCrossed {
			t.Errorf("%s: expected no page cross, got page crossed true", t.Name())
		}
	})

	t.Run("page cross", func(t *testing.T) {
		bus := newFakeSysBus()
		bus.data[0xFFFC] = 0x00
		bus.data[0xFFFD] = 0x06
		bus.data[0x0601] = 0x4B
		bus.data[0x4B] = 0xF0
		bus.data[0x4C] = 0xA2
		cpu := NewCpu(bus)
		actual, pageCrossed := cpu.getIndirectIndexedAddress(0x23)
		if actual != 0xA313 {
			t.Errorf("%s: expected address 0xA313, got 0x%X", t.Name(), actual)
		}
		if !pageCrossed {
			t.Errorf("%s: expected page cross, got page crossed false", t.Name())
		}
	})
}

func TestBrk(t *testing.T) {
	bus := newFakeSysBus()
	bus.data[0xFFFC] = 0x00
	bus.data[0xFFFD] = 0x06
	bus.data[0xFFFE] = 0x34
	bus.data[0xFFFF] = 0x12
	cpu := NewCpu(bus)
	oldPc := cpu.pc + 2
	cpu.brk(&opcodes[0x00])
	if cpu.pc != 0x1234 {
		t.Errorf("incorrect irq vector")
	}
	cpu.stackPop() // status flags
	low := cpu.stackPop()
	high := cpu.stackPop()
	if uint16(low) != oldPc&0x00FF || uint16(high) != oldPc&0xFF00>>8 {
		t.Errorf("incorrect program counter pushed to stack")
	}
}
