package util

import (
	"unicode"
	"unicode/utf8"
	"unsafe"
)

const (
	overflow64 uint64 = 0x8080808080808080
	overflow32 uint32 = 0x80808080
)

type Chars struct {
	slice           []byte // or []rune
	inBytes         bool
	trimLengthKnown bool
	trimLength      uint16

	// XXX Piggybacking item index here is a horrible idea. But I'm trying to
	// minimize the memory footprint by not wasting padded spaces.
	Index int32
}

func checkAscii(bytes []byte) (bool, int) {
	i := 0
	for ; i < len(bytes)-8; i += 8 {
		if (overflow64 & *(*uint64)(unsafe.Pointer(&bytes[i]))) > 0 {
			return false, i
		}
	}
	for ; i < len(bytes)-4; i += 4 {
		if (overflow32 & *(*uint32)(unsafe.Pointer(&bytes[i]))) > 0 {
			return false, i
		}
	}
	for ; i < len(bytes); i++ {
		if bytes[i] >= utf8.RuneSelf {
			return false, i
		}
	}
	return true, 0
}

// ToChars converts byte array into rune array
func ToChars(bytes []byte) Chars {
	inBytes, bytesUntil := checkAscii(bytes)
	if inBytes {
		return Chars{slice: bytes, inBytes: inBytes}
	}

	runes := make([]rune, bytesUntil, len(bytes))
	for i := 0; i < bytesUntil; i++ {
		runes[i] = rune(bytes[i])
	}
	for i := bytesUntil; i < len(bytes); {
		r, sz := utf8.DecodeRune(bytes[i:])
		i += sz
		runes = append(runes, r)
	}
	return RunesToChars(runes)
}

func RunesToChars(runes []rune) Chars {
	return Chars{slice: *(*[]byte)(unsafe.Pointer(&runes)), inBytes: false}
}

func (chars *Chars) optionalRunes() []rune {
	if chars.inBytes {
		return nil
	}
	return *(*[]rune)(unsafe.Pointer(&chars.slice))
}

func (chars *Chars) Get(i int) rune {
	if runes := chars.optionalRunes(); runes != nil {
		return runes[i]
	}
	return rune(chars.slice[i])
}

func (chars *Chars) Length() int {
	if runes := chars.optionalRunes(); runes != nil {
		return len(runes)
	}
	return len(chars.slice)
}

// TrimLength returns the length after trimming leading and trailing whitespaces
func (chars *Chars) TrimLength() uint16 {
	if chars.trimLengthKnown {
		return chars.trimLength
	}
	chars.trimLengthKnown = true
	var i int
	len := chars.Length()
	for i = len - 1; i >= 0; i-- {
		char := chars.Get(i)
		if !unicode.IsSpace(char) {
			break
		}
	}
	// Completely empty
	if i < 0 {
		return 0
	}

	var j int
	for j = 0; j < len; j++ {
		char := chars.Get(j)
		if !unicode.IsSpace(char) {
			break
		}
	}
	chars.trimLength = AsUint16(i - j + 1)
	return chars.trimLength
}

func (chars *Chars) TrailingWhitespaces() int {
	whitespaces := 0
	for i := chars.Length() - 1; i >= 0; i-- {
		char := chars.Get(i)
		if !unicode.IsSpace(char) {
			break
		}
		whitespaces++
	}
	return whitespaces
}

func (chars *Chars) ToString() string {
	if runes := chars.optionalRunes(); runes != nil {
		return string(runes)
	}
	return string(chars.slice)
}

func (chars *Chars) ToRunes() []rune {
	if runes := chars.optionalRunes(); runes != nil {
		return runes
	}
	bytes := chars.slice
	runes := make([]rune, len(bytes))
	for idx, b := range bytes {
		runes[idx] = rune(b)
	}
	return runes
}

func (chars *Chars) CopyRunes(dest []rune) {
	if runes := chars.optionalRunes(); runes != nil {
		copy(dest, runes)
		return
	}
	for idx, b := range chars.slice {
		dest[idx] = rune(b)
	}
	return
}
