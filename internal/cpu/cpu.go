package cpu

import (
	"github.com/theaaronruss/nes-emulator/internal/bus"
)

const (
	resetVector         = 0xFFFC
	irqVector           = 0xFFFE
	stackBase           = 0x0100
	initialStackPointer = 0xFD
)

const (
	flagCarry      uint8 = 1 << iota
	flagZero             = 1 << iota
	flagIntDisable       = 1 << iota
	flagDecimal          = 1 << iota
	flagBreak            = 1 << iota
	flagUnused           = 1 << iota
	flagOverflow         = 1 << iota
	flagNegative         = 1 << iota
)

type Cpu struct {
	a      uint8
	x      uint8
	y      uint8
	sp     uint8
	pc     uint16
	status uint8

	mainBus *bus.Bus
}

func NewCpu(mainBus *bus.Bus) *Cpu {
	cpu := &Cpu{
		mainBus: mainBus,
	}
	cpu.Reset()
	return cpu
}

func (c *Cpu) Reset() {
	c.sp = initialStackPointer
	pcLow := c.mainBus.Read(resetVector)
	pcHigh := c.mainBus.Read(resetVector + 1)
	c.pc = uint16(pcHigh)<<8 | uint16(pcLow)
	c.setFlag(flagIntDisable)
}

func (c *Cpu) ClockCycle() {
	opcode := c.mainBus.Read(c.pc)
	instruction := opcodes[opcode]
	instruction.fn(c, &instruction)
}

func (c *Cpu) setFlag(flag uint8) {
	c.status |= flag
}

func (c *Cpu) clearFlag(flag uint8) {
	c.status &= ^flag
}

func (c *Cpu) stackPush(data uint8) {
	address := stackBase + uint16(c.sp)
	c.mainBus.Write(address, data)
	c.sp--
}

func (c *Cpu) stackPop() uint8 {
	c.sp++
	address := stackBase + uint16(c.sp)
	return c.mainBus.Read(address)
}

func (c *Cpu) getAddress(addrMode addressMode) uint16 {
	switch addrMode {
	case addrModeIndexIndirX:
		zeroPageAddr := c.mainBus.Read(c.pc + 1)
		zeroPageAddr += c.x
		low := c.mainBus.Read(uint16(zeroPageAddr))
		high := c.mainBus.Read(uint16(zeroPageAddr) + 1)
		return uint16(high)<<8 | uint16(low)
	case addrModeZeroPage:
		return uint16(c.mainBus.Read(c.pc + 1))
	}
	return 0x0000
}

func (c *Cpu) forceBreak(instr *instruction) {
	c.pc += 2
	oldPcLow := uint8(c.pc & 0x00FF)
	oldPcHigh := uint8(c.pc & 0xFF00 >> 8)
	c.stackPush(oldPcHigh)
	c.stackPush(oldPcLow)
	c.stackPush(c.status | flagUnused | flagBreak)
	newPcLow := c.mainBus.Read(irqVector)
	newPcHigh := c.mainBus.Read(irqVector + 1)
	newPc := uint16(newPcHigh)<<8 | uint16(newPcLow)
	c.pc = newPc
}

func (c *Cpu) bitwiseOr(instr *instruction) {
	var value uint8
	if instr.addrMode == addrModeImmediate {
		value = c.mainBus.Read(c.pc + 1)
	} else {
		address := c.getAddress(instr.addrMode)
		value = c.mainBus.Read(address)
	}
	c.a |= value
	if c.a == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}
	if c.a&0b10000000 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) arithmeticShiftLeft(instr *instruction) {
	var value uint8
	var address uint16
	if instr.addrMode == addrModeAccumulator {
		value = c.a
	} else {
		address = c.getAddress(instr.addrMode)
		value = c.mainBus.Read(address)
	}
	if value&0b10000000 > 0 {
		c.setFlag(flagCarry)
	} else {
		c.clearFlag(flagCarry)
	}
	value <<= 1
	if value == 0 {
		c.setFlag(flagZero)
	} else {
		c.clearFlag(flagZero)
	}
	if value&0b10000000 > 0 {
		c.setFlag(flagNegative)
	} else {
		c.clearFlag(flagNegative)
	}
	if instr.addrMode == addrModeAccumulator {
		c.a = value
	} else {
		c.mainBus.Write(address, value)
	}
	c.pc += uint16(instr.bytes)
}

func (c *Cpu) pushProcessorStatus(instr *instruction) {
	c.stackPush(c.status | flagUnused | flagBreak)
	c.pc += uint16(instr.bytes)
}
