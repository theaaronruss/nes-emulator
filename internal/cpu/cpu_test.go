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
