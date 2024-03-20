package shorturl

import (
	"errors"
)

const (
	// proof against offensive words (removed 'a', 'e', 'i', 'o' and 'u')
	// unambiguous (removed 's', '5', 'I', 'l', '1', 'O' and '0')
	charset = "2346789BCDFGHJKLMNPQRSTVWXYZbcdfghjkmnpqrtvwxyz"

	base = uint64(len(charset))
)

var (
	ReverseCharset = [256]byte{}
	ErrInvalid     = errors.New("invalid code")
)

func init() {
	for i := 0; i < len(ReverseCharset); i++ {
		ReverseCharset[i] = 0xFF
	}
	for i := 0; i < len(charset); i++ {
		ReverseCharset[charset[i]] = byte(i)
	}
}

func Encode(i uint64) string {
	var tmp [11]byte
	result := tmp[:0]
	for i > 0 {
		result = append(result, charset[i%base])
		i /= base
	}
	l := len(result)
	for i := 0; i < l/2; i++ {
		result[i], result[l-1-i] = result[l-1-i], result[i]
	}
	return string(result)
}

func Decode(s string) (uint64, error) {
	var result uint64
	for i := 0; i < len(s); i++ {
		r := ReverseCharset[s[i]]
		if r == 0xFF {
			return 0, ErrInvalid
		}
		result = result*base + uint64(r)
	}
	return result, nil
}

func Validate(s string) bool {
	for i := 0; i < len(s); i++ {
		if ReverseCharset[s[i]] == 0xFF {
			return false
		}
	}
	return true
}
