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
)

type instruction struct {
	mnemonic string
	addrMode addressMode
	bytes    int
	cycles   int
	fn       func(*Cpu, *instruction)
}

var opcodes = [256]instruction{
	0x00: {brk, addrModeImplied, 2, 7, (*Cpu).brk},
}
