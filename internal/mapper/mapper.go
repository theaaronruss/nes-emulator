package mapper

type Mapper interface {
	TranslateProgramDataAddress(int, uint16) uint16
	TranslateCharacterDataAddress(int, uint16) uint16
}

var Mappers = map[int]func() Mapper{
	0: NewMapper000,
}
