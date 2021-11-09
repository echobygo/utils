package MiaUtils
import (
	"MiaGame/Library/MiaLog"
	"os"
	"strconv"
	"time"

)
func GetEnvVarDuration(n string) (time.Duration, bool) {
	str, ok := os.LookupEnv(n)
	if !ok {
		return 0, false
	}

	du, err := time.ParseDuration(str)
	if err != nil {
		MiaLog.CErrorf("can not parse env var %q as time.Duration, incorrect format  %s", str,err.Error())
		return 0, false
	}

	return du, true
}

func GetEnvVarInt(n string) (int, bool) {
	str, ok := os.LookupEnv(n)
	if !ok {
		return 0, false
	}

	num, err := strconv.Atoi(str)
	if err != nil {
		MiaLog.CErrorf("can not parse env var %q as int, incorrect format %s", str,err.Error())
		return 0, false
	}

	return num, true
}

func GetEnvVarBool(n string) (bool, bool) {
	str, ok := os.LookupEnv(n)
	if !ok {
		return false, false
	}

	num, err := strconv.ParseBool(str)
	if err != nil {
		MiaLog.CErrorf("can not parse env var %q as bool, incorrect format  %s", str,err.Error())
		return false, false
	}

	return num, true
}
