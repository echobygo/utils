package MiaUtils

import (
	"crypto/sha256"
	"errors"
	"github.com/shengdoushi/base58"
)

var ErrDecodeCheck = errors.New("cannot decode b58 with checksum")

var TronAlphabet = base58.NewAlphabet("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func Encode(input []byte) string {
	return base58.Encode(input, TronAlphabet)
}

func EncodeCheck(input []byte) string {
	h1 := sha256.Sum256(input)
	h2 := sha256.Sum256(h1[:])
	checksum := h2[:4]

	inputCheck := input
	inputCheck = append(inputCheck, checksum...)

	return Encode(inputCheck)
}

func Decode(input string) ([]byte, error) {
	return base58.Decode(input, TronAlphabet)
}

func DecodeCheck(input string) ([]byte, error) {
	decodeCheck, err := Decode(input)
	if err != nil {
		return nil, err
	}

	if len(decodeCheck) < 4 {
		return nil, ErrDecodeCheck
	}

	decodeData := decodeCheck[:len(decodeCheck)-4]

	h0 := sha256.Sum256(decodeData)
	h1 := sha256.Sum256(h0[:])

	if h1[0] == decodeCheck[len(decodeData)] &&
		h1[1] == decodeCheck[len(decodeData)+1] &&
		h1[2] == decodeCheck[len(decodeData)+2] &&
		h1[3] == decodeCheck[len(decodeData)+3] {
		return decodeData, nil
	}
	return nil, ErrDecodeCheck
}
