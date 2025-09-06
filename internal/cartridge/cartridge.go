package cartridge

import (
	"errors"
	"io"
	"os"
)

var headerStart = [...]byte{0x4E, 0x45, 0x53, 0x1A}

const (
	headerSize     int = 16
	trainerSize    int = 512
	prgRomBaseSize int = 16384
)

var (
	programData   []byte
	characterData []byte

	prgSize    int
	chrSize    int
	hasTrainer bool
)

func LoadCartridge(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	err = parseHeader(file)
	if err != nil {
		return err
	}

	if hasTrainer {
		file.Seek(int64(trainerSize), io.SeekCurrent)
	}

	err = readProgramData(file)
	if err != nil {
		return err
	}

	err = readCharacterData(file)
	if err != nil {
		return err
	}

	return nil
}

func ReadProgramData(address uint16) uint8 {
	return programData[address%uint16(prgRomBaseSize)]
}

func parseHeader(file *os.File) error {
	buffer := make([]byte, headerSize)
	n, err := file.Read(buffer)
	if err != nil {
		return err
	}
	if n < headerSize {
		return errors.New("unexpected end of rom file")
	}

	for i, b := range headerStart {
		if buffer[i] != b {
			return errors.New("invalid rom file")
		}
	}

	prgSize = int(buffer[4])
	chrSize = int(buffer[5])

	flags := buffer[6]
	hasTrainer = flags&0x04 > 0

	return nil
}

func readProgramData(file *os.File) error {
	programData = make([]byte, prgRomBaseSize)
	n, err := file.Read(programData)
	if err != nil {
		return err
	}
	if n < prgRomBaseSize {
		return errors.New("unexpected end of rom file")
	}
	return nil
}

func readCharacterData(file *os.File) error {
	characterData = make([]byte, chrSize)
	n, err := file.Read(characterData)
	if err != nil {
		return err
	}
	if n < chrSize {
		return errors.New("unexpected end of rom file")
	}
	return nil
}
