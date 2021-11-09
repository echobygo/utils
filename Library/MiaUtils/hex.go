package MiaUtils
import (
	"encoding/hex"
	"github.com/pkg/errors"
)

var (
	ErrEmptyString = errors.New("empty hex string")
)

// Encode encodes bytes as a hex string.
func Base58Encode(bytes []byte) string {
	encode := make([]byte, len(bytes)*2)
	hex.Encode(encode, bytes)
	return string(encode)
}

// Decode hex string as bytes
func Base58Decode(input string) ([]byte, error) {
	if len(input) == 0 {
		return nil, ErrEmptyString
	}
	return hex.DecodeString(input[:])
}

// Bytes2Hex returns the hexadecimal encoding of d.
func Bytes2Hex(d []byte) string {
	return hex.EncodeToString(d)
}
