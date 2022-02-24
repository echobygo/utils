package MiaUtils

/**
 * @Description
 * @Author 岁岁年年码不同
 * @Date 2021/11/27 3:35 上午
 **/

import (
	"io"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
)

func MarshalToString(v interface{}) (string, error) {
	return jsoniter.ConfigCompatibleWithStandardLibrary.MarshalToString(v)
}

func Marshal(v interface{}) ([]byte, error) {
	return jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(v)
}

func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return jsoniter.ConfigCompatibleWithStandardLibrary.MarshalIndent(v, prefix, indent)
}

func  UnmarshalFromString(str string, v interface{}) error {
	return jsoniter.ConfigCompatibleWithStandardLibrary.UnmarshalFromString(str, v)
}

func  Unmarshal(data []byte, v interface{}) error {
	return jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(data, v)
}

func Get(data []byte, path ...interface{}) jsoniter.Any {
	return jsoniter.ConfigCompatibleWithStandardLibrary.Get(data, path)
}

func NewEncoder(writer io.Writer) *jsoniter.Encoder {
	return jsoniter.ConfigCompatibleWithStandardLibrary.NewEncoder(writer)
}

func NewDecoder(reader io.Reader) *jsoniter.Decoder {
	return jsoniter.ConfigCompatibleWithStandardLibrary.NewDecoder(reader)
}

func Valid(data []byte) bool {
	return jsoniter.ConfigCompatibleWithStandardLibrary.Valid(data)
}

func RegisterExtension(extension jsoniter.Extension) {
	jsoniter.ConfigCompatibleWithStandardLibrary.RegisterExtension(extension)
}

func DecoderOf(typ reflect2.Type) jsoniter.ValDecoder {
	return jsoniter.ConfigCompatibleWithStandardLibrary.DecoderOf(typ)
}

func EncoderOf(typ reflect2.Type) jsoniter.ValEncoder {
	return jsoniter.ConfigCompatibleWithStandardLibrary.EncoderOf(typ)
}
