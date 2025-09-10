package nes

type addressMode int

const (
	addrModeZeroPageX addressMode = iota
	addrModeZeroPageY
	addrModeAbsoluteX
	addrModeAbsoluteY
	addrModeIndexedIndir
	addrModeIndirIndexed
	addrModeImplied
	addrModeAccumulator
	addrModeImmediate
	addrModeZeroPage
	addrModeAbsolute
	addrModeRelative
	addrModeIndirect
)

// opcode mnemonics
const (
	brk = "BRK"
	ora = "ORA"
	slo = "*SLO"
)

type instruction struct {
	mnemonic string
	addrMode addressMode
	bytes    int
	cycles   int
	fn       func(*Cpu, *instruction, uint16)
}

var opcodes = [256]instruction{
	0x00: {brk, addrModeImplied, 2, 7, (*Cpu).brk},
	0x01: {ora, addrModeIndexedIndir, 2, 6, (*Cpu).ora},
	0x03: {slo, addrModeIndexedIndir, 2, 8, (*Cpu).slo},
	0x05: {ora, addrModeZeroPage, 2, 3, (*Cpu).ora},
	0x07: {slo, addrModeZeroPage, 2, 5, (*Cpu).slo},
	0x09: {ora, addrModeImmediate, 2, 2, (*Cpu).ora},
	0x0D: {ora, addrModeAbsolute, 3, 4, (*Cpu).ora},
	0x0F: {slo, addrModeAbsolute, 3, 6, (*Cpu).slo},
	0x11: {ora, addrModeIndirIndexed, 2, 5, (*Cpu).ora},
	0x13: {slo, addrModeIndirIndexed, 2, 8, (*Cpu).slo},
	0x15: {ora, addrModeZeroPageX, 2, 4, (*Cpu).ora},
	0x17: {slo, addrModeZeroPageX, 2, 6, (*Cpu).slo},
	0x19: {ora, addrModeAbsoluteY, 3, 4, (*Cpu).ora},
	0x1B: {slo, addrModeAbsoluteY, 3, 7, (*Cpu).slo},
	0x1D: {ora, addrModeAbsoluteX, 3, 4, (*Cpu).ora},
	0x1F: {slo, addrModeAbsoluteX, 3, 7, (*Cpu).slo},
}
