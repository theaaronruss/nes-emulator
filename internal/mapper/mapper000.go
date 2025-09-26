package mapper

type Mapper000 struct {
}

func NewMapper000() Mapper {
	return &Mapper000{}
}

func (mapper *Mapper000) TranslateProgramDataAddress(chunks int, addr uint16) uint16 {
	if chunks == 1 {
		return addr % 16384
	}
	return addr % 32768
}

func (mapper *Mapper000) TranslateCharacterDataAddress(chunks int, addr uint16) uint16 {
	return addr % 8192
}
