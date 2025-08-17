package cpu

import (
	"testing"
)

func TestForceBreak(t *testing.T) {
	c := NewCpu()
	c.pc = 0x0800
	c.status = 0b10000011
	forceBreak(c)
	if c.pc != 0xFFFE {
		t.Error("incorrect program counter")
	}
	if c.status != 0b10000111 {
		t.Error("incorrect status flags")
	}
	if c.cycleDelay != 7 {
		t.Error("incorrect cycle delay")
	}
	if c.stackPop() != 0b10110011 || c.stackPop() != 0x02 || c.stackPop() != 0x08 {
		t.Error("incorrect stack contents")
	}
}

func TestBitwiseOrImmediate(t *testing.T) {
	c := NewCpu()
	c.a = 0b01010111
	c.memory[0x00] = 0b01000010
	c.memory[c.pc+1] = 0x00
	bitwiseOrImmediate(c)
	if c.a != 0b01010111 {
		t.Error("incorrect accumulator value")
	}
	if c.testZero() {
		t.Error("incorrect zero flag")
	}
	if c.testNegative() {
		t.Error("incorrect negative flag")
	}
}

func TestBitwiseOrImmediateZero(t *testing.T) {
	c := NewCpu()
	c.a = 0b00000000
	c.memory[0x00] = 0b00000000
	c.memory[c.pc+1] = 0x00
	bitwiseOrImmediate(c)
	if c.a != 0b00000000 {
		t.Error("incorrect accumulator value")
	}
	if !c.testZero() {
		t.Error("incorrect zero flag")
	}
	if c.testNegative() {
		t.Error("incorrect negative flag")
	}
}

func TestBitwiseOrImmediateNegative(t *testing.T) {
	c := NewCpu()
	c.a = 0b11010111
	c.memory[0x00] = 0b01000010
	c.memory[c.pc+1] = 0x00
	bitwiseOrImmediate(c)
	if c.a != 0b11010111 {
		t.Error("incorrect accumulator value")
	}
	if c.testZero() {
		t.Error("incorrect zero flag")
	}
	if !c.testNegative() {
		t.Error("incorrect negative flag")
	}
}

func TestBitwiseAndImmediate(t *testing.T) {
	c := NewCpu()
	c.a = 0b11010101
	c.memory[c.pc+1] = 0b01011101
	bitwiseAndImmediate(c)
	if c.a != 0b01010101 {
		t.Error("invalid accumulator value")
	}
	if c.testZero() {
		t.Error("invalid zero flag")
	}
	if c.testNegative() {
		t.Error("invalid negative flag")
	}
}

func TestBitwiseAndImmediateZero(t *testing.T) {
	c := NewCpu()
	c.a = 0b10101010
	c.memory[c.pc+1] = 0b01010101
	bitwiseAndImmediate(c)
	if c.a != 0b00000000 {
		t.Error("invalid accumulator value")
	}
	if !c.testZero() {
		t.Error("invalid zero flag")
	}
	if c.testNegative() {
		t.Error("invalid negative flag")
	}
}

func TestBitwiseAndImmediateNegative(t *testing.T) {
	c := NewCpu()
	c.a = 0b11010101
	c.memory[c.pc+1] = 0b11011101
	bitwiseAndImmediate(c)
	if c.a != 0b11010101 {
		t.Error("invalid accumulator value")
	}
	if c.testZero() {
		t.Error("invalid zero flag")
	}
	if !c.testNegative() {
		t.Error("invalid negative flag")
	}
}
