package cpu

type addressMode int

const (
	addrModeZeroPageX addressMode = iota
	addrModeZeroPageY
	addrModeAbsoluteX
	addrModeAbsoluteY
	addrModeIndexIndirX
	addrModeIndirIndexY
	addrModeImplicit
	addrModeAccumulator
	addrModeImmediate
	addrModeZeroPage
	addrModeAbsolute
	addrModeRelative
	addrModeIndirect
)

type instruction struct {
	mnemonic string
	addrMode addressMode
	bytes    int
	cycles   int
}

var opcodes = [255]instruction{}
