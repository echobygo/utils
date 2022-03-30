package MiaUtils
import (
	"fmt"
	"github.com/google/cel-go/common/types/ref"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"time"
	"encoding/json"
	"crypto/md5"
	"encoding/base64"
	"net/url"
	"unsafe"
	"encoding/hex"
)
const (
	intType                  = "int"
	int8Type                 = "int8"
	int16Type                = "int16"
	int32Type                = "int32"
	int64Type                = "int64"
	uintType                 = "uint"
	uint8Type                = "uint8"
	uint16Type               = "uint16"
	uint32Type               = "uint32"
	uint64Type               = "uint64"
	bigIntType               = "bigInt"
	bigFloatType             = "bigFloat"
	float64Type              = "float64"
	stringType               = "string"
	boolType                 = "bool"
	listType                 = "list"
	invalidTypeMsg           = "invalid variable, expect %v got %v"
)
var (
	sourceIsEmpty                  = fmt.Errorf("source is empty")
	invalidExpression              = fmt.Errorf("invalid expression")
	invalidMethodFormat            = fmt.Errorf("invalid method format")
	abiNotFound                    = fmt.Errorf("abi is not found")
	methodNotFound                 = fmt.Errorf("method is not found")
	paramsArgumentsNotMatch        = fmt.Errorf("params and arguments are not matched")
	paramValueNotCorrect           = fmt.Errorf("param's value is not correct")
	unsupportedType                = fmt.Errorf("unsupported type")
	invalidIfParams                = fmt.Errorf("not enough arguments for If function")
	invalidIfStatement             = fmt.Errorf("invalid if statement")
	incorrectReturnedValueInIFFunc = fmt.Errorf("IF func must returns only 1 bool value")
	invalidSignal                  = fmt.Errorf("invalid signal")
	stopSignal                     = fmt.Errorf("signal stop has been applied")
	invalidVariables               = fmt.Errorf("invalid variables")
	variableNotFound               = fmt.Errorf("variable not found")
	invalidForEachParam            = fmt.Errorf("invalid for each param")
	invalidForEachStatement        = fmt.Errorf("invalid for each statement")
	notEnoughArgsForSplit          = fmt.Errorf("not enough arguments for split function")
	notEnoughArgsForFunc           = fmt.Errorf("not enough arguments for create/call Func function")
	invalidSplitArgs               = fmt.Errorf("invalid split arguments")
	invalidDefineFunc              = fmt.Errorf("invalid define function")
)
var (
	SupportedTypes = map[string]func(val interface{}) (interface{}, error){
		intType: func(val interface{}) (interface{}, error) {
			kind := reflect.TypeOf(val).Kind()
			if kind != reflect.String {
				if kind == reflect.Int {
					return val.(int), nil
				}
				return nil, fmt.Errorf(invalidTypeMsg, intType, kind.String())
			}
			v, err := strconv.ParseInt(val.(string), 10, 32)
			if err != nil {
				return nil, err
			}
			return int(v), nil
		},
		int8Type: func(val interface{}) (interface{}, error) {
			kind := reflect.TypeOf(val).Kind()
			if kind != reflect.String {
				if kind == reflect.Int8 {
					return val.(int8), nil
				}
				return nil, fmt.Errorf(invalidTypeMsg, int8Type, kind.String())
			}
			v, err := strconv.ParseInt(val.(string), 10, 8)
			if err != nil {
				return nil, err
			}
			return int8(v), nil
		},
		int16Type: func(val interface{}) (interface{}, error) {
			kind := reflect.TypeOf(val).Kind()
			if kind != reflect.String {
				if kind == reflect.Int16 {
					return val.(int16), nil
				}
				return nil, fmt.Errorf(invalidTypeMsg, int16Type, kind.String())
			}
			v, err := strconv.ParseInt(val.(string), 10, 16)
			if err != nil {
				return nil, err
			}
			return int16(v), nil
		},
		int32Type: func(val interface{}) (interface{}, error) {
			kind := reflect.TypeOf(val).Kind()
			if kind != reflect.String {
				if kind == reflect.Int32 {
					return val.(int32), nil
				}
				return nil, fmt.Errorf(invalidTypeMsg, int32Type, kind.String())
			}
			v, err := strconv.ParseInt(val.(string), 10, 32)
			if err != nil {
				return nil, err
			}
			return int32(v), nil
		},
		int64Type: func(val interface{}) (interface{}, error) {
			kind := reflect.TypeOf(val).Kind()
			if kind != reflect.String {
				if kind == reflect.Int64 {
					return val.(int64), nil
				} else if kind.String() == "*big.Int" {
					return val.(*big.Int).Int64(), nil
				}
				return nil, fmt.Errorf(invalidTypeMsg, int64Type, kind.String())
			}
			return strconv.ParseInt(val.(string), 10, 64)
		},
		uintType: func(val interface{}) (interface{}, error) {
			kind := reflect.TypeOf(val).Kind()
			if kind != reflect.String {
				if kind == reflect.Uint {
					return val.(uint), nil
				}
				return nil, fmt.Errorf(invalidTypeMsg, uintType, kind.String())
			}
			v, err := strconv.ParseUint(val.(string), 10, 32)
			if err != nil {
				return nil, err
			}
			return uint(v), nil
		},
		uint8Type: func(val interface{}) (interface{}, error) {
			kind := reflect.TypeOf(val).Kind()
			if kind != reflect.String {
				if kind == reflect.Uint8 {
					return val.(uint8), nil
				}
				return nil, fmt.Errorf(invalidTypeMsg, uint8Type, kind.String())
			}
			v, err := strconv.ParseUint(val.(string), 10, 8)
			if err != nil {
				return nil, err
			}
			return uint8(v), nil
		},
		uint16Type: func(val interface{}) (interface{}, error) {
			kind := reflect.TypeOf(val).Kind()
			if kind != reflect.String {
				if kind == reflect.Uint16 {
					return val.(uint16), nil
				}
				return nil, fmt.Errorf(invalidTypeMsg, uint16Type, kind.String())
			}
			v, err := strconv.ParseUint(val.(string), 10, 16)
			if err != nil {
				return nil, err
			}
			return uint16(v), nil
		},
		uint32Type: func(val interface{}) (interface{}, error) {
			kind := reflect.TypeOf(val).Kind()
			if kind != reflect.String {
				if kind == reflect.Uint32 {
					return val.(uint32), nil
				}
				return nil, fmt.Errorf(invalidTypeMsg, uint32Type, kind.String())
			}
			v, err := strconv.ParseUint(val.(string), 10, 32)
			if err != nil {
				return nil, err
			}
			return uint32(v), nil
		},
		uint64Type: func(val interface{}) (interface{}, error) {
			kind := reflect.TypeOf(val).Kind()
			if kind != reflect.String {
				if kind == reflect.Uint64 {
					return val.(uint64), nil
				} else if kind == reflect.Int64 { // by default CEL convert number to int64
					return uint64(val.(int64)), nil
				} else if reflect.ValueOf(val).Type().String() == "*big.Int" {
					return val.(*big.Int).Uint64(), nil
				}
				return nil, fmt.Errorf(invalidTypeMsg, uint64Type, kind.String())
			}
			return strconv.ParseUint(val.(string), 10, 64)
		},
		bigIntType: func(val interface{}) (interface{}, error) {
			kind := reflect.TypeOf(val).Kind()
			if kind != reflect.String {
				if kind == reflect.Int64 {
					return big.NewInt(val.(int64)), nil
				} else if kind == reflect.Uint64 {
					return big.NewInt(int64(val.(uint64))), nil
				} else if reflect.ValueOf(val).Type().String() == "*big.Int" {
					return val.(*big.Int), nil
				}
				return nil, fmt.Errorf(invalidTypeMsg, bigIntType, reflect.ValueOf(val).Type().String())
			}
			v, _ := big.NewInt(0).SetString(val.(string), 10)
			return v, nil
		},
		bigFloatType: func(val interface{}) (interface{}, error) {
			kind := reflect.TypeOf(val).Kind()
			if kind != reflect.String {
				if kind == reflect.Float64 {
					return big.NewFloat(val.(float64)), nil
				} else if reflect.ValueOf(val).Type().String() == "*big.Float" {
					return val.(*big.Float), nil
				}
				return nil, fmt.Errorf(invalidTypeMsg, bigFloatType, kind.String())
			}
			v, _ := big.NewFloat(0).SetString(val.(string))
			return v, nil
		},
		float64Type: func(val interface{}) (interface{}, error) {
			kind := reflect.TypeOf(val).Kind()
			if kind != reflect.String {
				if kind == reflect.Float64 {
					return val.(float64), nil
				}
				return nil, fmt.Errorf(invalidTypeMsg, uint64Type, kind.String())
			}
			return strconv.ParseFloat(val.(string), 64)
		},
		stringType: func(val interface{}) (interface{}, error) {
			kind := reflect.TypeOf(val).Kind()
			if kind != reflect.String {
				return InterfaceToString(val)
			}
			return val.(string), nil
		},
		boolType: func(val interface{}) (interface{}, error) {
			if reflect.ValueOf(val).Type().Kind() == reflect.Bool {
				return reflect.ValueOf(val).Bool(), nil
			}
			return strconv.ParseBool(val.(string))
		},
		listType: func(val interface{}) (interface{}, error) {
			kind := reflect.TypeOf(val).Kind()
			if kind != reflect.Array && kind != reflect.Slice {
				return nil, fmt.Errorf(invalidTypeMsg, listType, kind.String())
			}
			return interfaceToSlice(val)
		},
	}
)
func isType(t string, vals ...reflect.Value) bool {
	for _, val := range vals {
		if !strings.Contains(val.Type().String(), t) {
			return false
		}
	}
	return true
}


func InterfaceToString(val interface{}) (string, error) {
	v := reflect.ValueOf(val)
	if isType("ref.Val", v) {
		return InterfaceToString(val.(ref.Val).Value())
	} else if isType("big.Int", v) {
		return val.(*big.Int).String(), nil
	} else if isType("big.Float", v) {
		return val.(*big.Float).String(), nil
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), []byte("f")[0], 8, 32), nil
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), []byte("f")[0], 8, 64), nil
	case reflect.String:
		return v.String(), nil
	}
	return "", unsupportedType
}

func convertToNative(val reflect.Value) (interface{}, error) {
	if val.Type().String() == "*big.Int" {
		return val.Interface().(*big.Int), nil
	} else if val.Type().String() == "*big.Float" {
		return val.Interface().(*big.Float), nil
	}
	kind := val.Kind()
	switch kind {
	case reflect.String:
		return val.String(), nil
	case reflect.Bool:
		return val.Bool(), nil
	case reflect.Uint, reflect.Uintptr:
		v, _ := big.NewInt(0).SetString(strconv.FormatUint(val.Uint(), 10), 10)
		return v, nil
	case reflect.Uint8:
		return uint8(val.Uint()), nil
	case reflect.Uint16:
		return uint16(val.Uint()), nil
	case reflect.Uint32:
		return uint32(val.Uint()), nil
	case reflect.Uint64:
		return val.Uint(), nil
	case reflect.Int:
		v, _ := big.NewInt(0).SetString(strconv.FormatInt(val.Int(), 10), 10)
		return v, nil
	case reflect.Int8:
		return int8(val.Int()), nil
	case reflect.Int16:
		return int16(val.Int()), nil
	case reflect.Int32:
		return int32(val.Int()), nil
	case reflect.Int64:
		return val.Int(), nil
	case reflect.Float32, reflect.Float64:
		return val.Float(), nil
	}
	return "", fmt.Errorf("unsupported value type %v", val.Type().String())
}

func interfaceToSlice(val interface{}) ([]interface{}, error) {
	if reflect.TypeOf(val).Kind() != reflect.Slice && reflect.TypeOf(val).Kind() != reflect.Array {
		return nil, fmt.Errorf("invalid list type, expect slice or array, got %v", reflect.TypeOf(val).Kind().String())
	}
	results := make([]interface{}, 0)
	if reflect.TypeOf(val).Elem().String() == "ref.Val" {
		for _, v := range val.([]ref.Val) {
			results = append(results, v.Value())
		}
		return results, nil
	}

	switch reflect.TypeOf(val).Elem().Kind() {
	case reflect.String:
		for _, v := range val.([]string) {
			results = append(results, v)
		}
	case reflect.Bool:
		for _, v := range val.([]bool) {
			results = append(results, v)
		}
	case reflect.Int:
		for _, v := range val.([]int) {
			results = append(results, v)
		}
	case reflect.Int8:
		for _, v := range val.([]int8) {
			results = append(results, v)
		}
	case reflect.Int16:
		for _, v := range val.([]int16) {
			results = append(results, v)
		}
	case reflect.Int32:
		for _, v := range val.([]int32) {
			results = append(results, v)
		}
	case reflect.Int64:
		for _, v := range val.([]int64) {
			results = append(results, v)
		}
	case reflect.Uint:
		for _, v := range val.([]uint) {
			results = append(results, v)
		}
	case reflect.Uint8:
		for _, v := range val.([]uint8) {
			results = append(results, v)
		}
	case reflect.Uint16:
		for _, v := range val.([]uint16) {
			results = append(results, v)
		}
	case reflect.Uint32:
		for _, v := range val.([]uint32) {
			results = append(results, v)
		}
	case reflect.Uint64:
		for _, v := range val.([]uint64) {
			results = append(results, v)
		}
	case reflect.Uintptr:
		for _, v := range val.([]uintptr) {
			results = append(results, v)
		}
	case reflect.Float32:
		for _, v := range val.([]float32) {
			results = append(results, v)
		}
	case reflect.Float64:
		for _, v := range val.([]float64) {
			results = append(results, v)
		}
	case reflect.Interface:
		return val.([]interface{}), nil
	default:
		return nil, unsupportedType
	}
	return results, nil
}


func F2i(f float64) int {
	i, _ := strconv.Atoi(fmt.Sprintf("%1.0f", f))
	return i
}

func Timeunixtostring(e string) ( timestr string, err error) {
	data, err := strconv.ParseInt(e, 10, 64)
	datatime := time.Unix(data/1000, 0)
	timestr   =  datatime.Format("2006-01-02 15:04:05")
	return
}


//value转化为Interface
func value2Interface(fieldValue reflect.Value) interface{} {
	fieldType := fieldValue.Type()
	k := fieldType.Kind()
	switch k {
	case reflect.Bool:
		return fieldValue.Bool()
		//Int()返回的是int64,而不是int
	case reflect.Int:
		return int(fieldValue.Int())
	case reflect.Int8:
		return int8(fieldValue.Int())
	case reflect.Int16:
		return int16(fieldValue.Int())
	case reflect.Int32:
		return int32(fieldValue.Int())
	case reflect.Int64:
		return int64(fieldValue.Int())
		//Uint()返回的是uint64,而不是uint
	case reflect.Uint:
		return uint(fieldValue.Uint())
	case reflect.Uint8:
		return uint8(fieldValue.Uint())
	case reflect.Uint16:
		return uint16(fieldValue.Uint())
	case reflect.Uint32:
		return uint32(fieldValue.Uint())
	case reflect.Uint64:
		return uint64(fieldValue.Uint())
		//Float()返回的是float64
	case reflect.Float32:
		return float32(fieldValue.Float())
	case reflect.Float64:
		return float64(fieldValue.Float())
		//Complex()返回的是complex128
	case reflect.Complex64:
		return complex64(fieldValue.Complex())
	case reflect.Complex128:
		return complex128(fieldValue.Complex())
		//case reflect.Array:
		// case reflect.Chan:
		// case reflect.Func:
		// case reflect.Interface:
		// case reflect.Map:
	case reflect.Ptr:
		return fieldValue.Pointer()
	// case reflect.Slice:
	case reflect.String:
		return fieldValue.String()
	// case reflect.Struct:
	default:
		return fieldValue.Interface()
	}

}

//把结构体中的数据转到切片中
func StructToSlice(fromStructName interface{}) (fieldNames []string, slicedata []interface{}, err error) {
	//确保fromStructName是结构体
	refValue := reflect.ValueOf(fromStructName)
	if refValue.Kind() != reflect.Struct {
		err = fmt.Errorf("the argument fromStructName must be a Struct.")
		return
	}
	//反射获得所有列名及值
	fieldNum := refValue.NumField()
	slicedata = make([]interface{}, 0, fieldNum)
	fieldNames = make([]string, 0, fieldNum)
	for i := 0; i < fieldNum; i++ {
		fieldNames = append(fieldNames, refValue.Type().Field(i).Name)
		slicedata = append(slicedata, value2Interface(refValue.Field(i)))
	}

	return
}

func StringSliceContains(slice []string, needle string) bool {
	for _, s := range slice {
		if s == needle {
			return true
		}
	}

	return false
}
func String2float64(s string) (f64 float64) {

	i, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return i
}

func Struct2JsonStr(str interface{}) string {
	if str==nil {
		return ""
	}
	result,err:=json.Marshal(str)
	if err!=nil {
		fmt.Println(err.Error())
	}
	return string(result)

}
func String2BigInt(s string)(st int64) {
	st,_=strconv.ParseInt(s,10,64)
	return
}
func Uint2String(value uint64 )string{
	return strconv.FormatUint(value,10)
}
func Int2String(value int64 )string{
	return strconv.FormatInt(value,10)
}
func String2Int(s string)(st int) {
	st,_=strconv.Atoi(s)
	return
}
func StringMd5(str string) string {
	hash := md5.New()
	hash.Write([]byte(str))

	return hex.EncodeToString(hash.Sum(nil))
}
// String2Base64 返回指定str的base64编码
func String2Base64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}
func Base642Byte(base64Str string) []byte {
	b, _ := base64.StdEncoding.DecodeString(base64Str)
	return b
}

func URLEncode(str string) string {
	strs := strings.Split(str, " ")

	for i, s := range strs {
		strs[i] = url.QueryEscape(s)
	}

	return strings.Join(strs, "%20")
}


func Strings2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

