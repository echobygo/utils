package DB

import (
	"MiaGame/Library/MiaError"
	"MiaGame/Library/MiaLog"
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/mediocregopher/radix/v3"
	"github.com/mediocregopher/radix/v3/resp/resp2"
	gconv "github.com/og/x/conv"
	"reflect"
	"strconv"
	"time"
)

// Config the redis configuration used inside sessions
type RedisConfig struct {
	// Network protocol. Defaults to "tcp".
	Network string
	// Addr of the redis server. Defaults to "127.0.0.1:6379".
	Addr string
	// Password string .If no password then no 'AUTH'. Defaults to "".
	Password string
	// If Database is empty "" then no 'SELECT'. Defaults to "".
	Database string
	// MaxActive. Defaults to 10.
	MaxActive int
	// Timeout for connect, write and read, defaults to 30 seconds, 0 means no timeout.
	Timeout time.Duration
	// Prefix "myprefix-for-this-website". Defaults to "".
	Prefix string
	// Delim the delimeter for the keys on the sessiondb. Defaults to "-".
	Delim string
}

// RadixDriver the Redis service based on the radix go client,
// contains the config and the redis pool.
type RadixDriver struct {
	Connected        bool
	Config           RedisConfig
	pool             *radix.Pool
	IsCheckReconnect bool
}

// Connect connects to the redis, called only once
func (r *RadixDriver) ReConnect(c RedisConfig) error {
	if c.Timeout < 0 {
		c.Timeout = time.Duration(30) * time.Second
	}

	if c.Network == "" {
		c.Network = "tcp"
	}

	if c.Addr == "" {
		c.Addr = "127.0.0.1:6379"
	}

	if c.MaxActive == 0 {
		c.MaxActive = 10
	}

	if c.Delim == "" {
		c.Delim = "-"
	}

	customConnFunc := func(network, addr string) (radix.Conn, error) {
		var options []radix.DialOpt

		if c.Password != "" {
			options = append(options, radix.DialAuthPass(c.Password))
		}

		if c.Timeout > 0 {
			options = append(options, radix.DialTimeout(c.Timeout))
		}

		if c.Database != "" {
			dbIndex, err := strconv.Atoi(c.Database)
			if err == nil {
				options = append(options, radix.DialSelectDB(dbIndex))
			}

		}

		return radix.Dial(network, addr, options...)
	}

	pool, err := radix.NewPool(c.Network, c.Addr, c.MaxActive, radix.PoolConnFunc(customConnFunc))
	MiaLog.CInfo(c.Addr, c.Network)
	if err != nil {
		MiaLog.CInfo(err.Error())
		r.IsCheckReconnect = true
		return err
	}

	r.Connected = true
	r.pool = pool
	r.Config = c
	return nil
}

//心跳包
// PingPong sends a ping and receives a pong, if no pong received then returns false and filled error
func (r *RadixDriver) PingPong() (bool, error) {
	var msg string
	err := r.pool.Do(radix.Cmd(&msg, "PING"))
	if err != nil {
		return false, err
	}
	return (msg == "PONG"), nil
}

// CloseConnection closes the redis connection.
func (r *RadixDriver) CloseConnection() error {
	if r.pool != nil {
		return r.pool.Close()
	}
	return errors.New("redis: already closed")
}

// Get returns value, err by its key
// returns nil and a filled error if something bad happened.
func (r *RadixDriver) Get(key string) (redisVal string, err error) {
	mn := radix.MaybeNil{Rcv: &redisVal}
	err = r.pool.Do(radix.Cmd(&mn, "GET", r.Config.Prefix+key))
	if MiaError.CheckError(err) {
        return "",err
    }
	if mn.Nil {
		return "", fmt.Errorf("%s: %w", key, errors.New("key not found"))
	}
	return redisVal, nil
}
func (r *RadixDriver) GetByte(key string) (redisVal []byte, err error) {
	mn := radix.MaybeNil{Rcv: &redisVal}
	err = r.pool.Do(radix.Cmd(&mn, "GET", r.Config.Prefix+key))
	if MiaError.CheckError(err) {
		return nil,err
	}
	if mn.Nil {
		return nil, fmt.Errorf("%s: %w", key, errors.New("key not found"))
	}
	return redisVal, nil
}
func (r *RadixDriver) Set(key string, value interface{}, secondsLifetime int64) error {
	var cmd radix.CmdAction
	if secondsLifetime > 0 {
		cmd = radix.FlatCmd(nil, "SETEX", r.Config.Prefix+key, secondsLifetime, value)
	} else {
		cmd = radix.FlatCmd(nil, "SET", r.Config.Prefix+key, value) // MSET same performance...
	}
	return r.pool.Do(cmd)
}
func (r *RadixDriver) Incr(key string) error {
	err := r.pool.Do(radix.Cmd(nil, "incr", r.Config.Prefix+key))
	return err
}
func (r *RadixDriver) Delete(key string) error {
	err := r.pool.Do(radix.Cmd(nil, "DEL", r.Config.Prefix+key))
	return err
}
func (r *RadixDriver) Exec(action radix.CmdAction)  (error){
    err := r.pool.Do(action)
    MiaError.CheckError(err)
    return err
}

func (r *RadixDriver) DeletePrefix(prefix string) {
	var keyString []string
	err := r.pool.Do(radix.FlatCmd(&keyString, "keys", prefix+"*"))
	if !MiaError.CheckError(err) {
		for _, v := range keyString {
			err = r.pool.Do(radix.FlatCmd(nil, "Del", v))
			MiaError.CheckError(err)
		}
	}
}

func (r *RadixDriver) TTL(key string) (seconds int64, hasExpiration bool, found bool) {
	var redisVal interface{}
	err := r.pool.Do(radix.Cmd(&redisVal, "TTL", r.Config.Prefix+key))
	if err != nil {
		return -2, false, false
	}
	seconds = redisVal.(int64)
	hasExpiration = seconds > -1
	found = seconds != -2
	return
}
func (r *RadixDriver) updateTTLConn(key string, newSecondsLifeTime int64) error {
	var reply int
	err := r.pool.Do(radix.FlatCmd(&reply, "EXPIRE", r.Config.Prefix+key, newSecondsLifeTime))
	if err != nil {
		return err
	}
	if reply == 1 {
		return nil
	} else if reply == 0 {
		return fmt.Errorf("unable to update expiration, the key '%s' was stored without ttl", key)
	} // do not check for -1.

	return nil
}

type scanResult struct {
	cur  string
	keys []string
}

func (s *scanResult) UnmarshalRESP(br *bufio.Reader) error {
	var ah resp2.ArrayHeader
	if err := ah.UnmarshalRESP(br); err != nil {
		return err
	} else if ah.N != 2 {
		return errors.New("not enough parts returned")
	}

	var c resp2.BulkString
	if err := c.UnmarshalRESP(br); err != nil {
		return err
	}

	s.cur = c.S
	s.keys = s.keys[:0]

	return (resp2.Any{I: &s.keys}).UnmarshalRESP(br)
}
//查询出所有相似的key
func (r *RadixDriver) GetKeys(cursor, prefix string) ([]string, error) {
	var res scanResult
	err := r.pool.Do(radix.Cmd(&res, "SCAN", cursor, "MATCH", r.Config.Prefix+prefix+"*", "COUNT", "1000"))
	if err != nil {
		return nil, err
	}

	keys := res.keys[0:]
	if res.cur != "0" {
		moreKeys, err := r.GetKeys(res.cur, prefix)
		if err != nil {
			return nil, err
		}

		keys = append(keys, moreKeys...)
	}

	return keys, nil
}

// UpdateTTLMany like `UpdateTTL` but for all keys starting with that "prefix",
// it is a bit faster operation if you need to update all sessions keys (although it can be even faster if we used hash but this will limit other features),
// look the `sessions/Database#OnUpdateExpiration` for example.
func (r *RadixDriver) UpdateTTLMany(prefix string, newSecondsLifeTime int64) error {
	keys, err := r.GetKeys("0", prefix)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err = r.updateTTLConn(key, newSecondsLifeTime); err != nil { // fail on first error.
			return err
		}
	}

	return err
}

// UpdateTTL will update the ttl of a key.
// Using the "EXPIRE" command.
// Read more at: https://redis.io/commands/expire#refreshing-expires
func (r *RadixDriver) UpdateTTL(key string, newSecondsLifeTime int64) error {
	return r.updateTTLConn(key, newSecondsLifeTime)
}

func (self *RadixDriver) Exists(key string) (exists bool) {
	data := radix.MaybeNil{Rcv: &exists}
	err := self.pool.Do(radix.Cmd(&data, EXISTS, key))
	Check(err)
	return
}
func (r *RadixDriver) IsSetTableExist(key string, existValue string) (IsExist bool) {

	result := r.Exists(key)
	if !result {
		return false
	}
	data := radix.MaybeNil{Rcv: &IsExist}
	err := r.pool.Do(radix.Cmd(&data, "sismember", key, existValue))
	MiaError.CheckError(err)
	return
}
func (r *RadixDriver) SaveSliceToRedisSet(key string, info interface{}) (err error) {
	var list []string
	list = append(list, key)
	if reflect.TypeOf(info).Kind() == reflect.Slice {
		s := reflect.ValueOf(info)
		for i := 0; i < s.Len(); i++ {
			ele := s.Index(i)
			if ele.Type().Kind() == reflect.String {
				list = append(list, ele.Interface().(string))
			} else if ele.Type().Kind() == reflect.Int {
				list = append(list, strconv.Itoa(ele.Interface().(int)))
			}

		}
	}
	return r.pool.Do(radix.Cmd(nil, "sadd", list...))
}

//SaveToRedis 将一个结构保存到redis中
func (r *RadixDriver) SaveToRedis(key string, info interface{}) {
	tableName := key
	dataStruct := reflect.Indirect(reflect.ValueOf(info))
	dataStructType := dataStruct.Type()
	for i := 0; i < dataStructType.NumField(); i++ {
		fieldType := dataStructType.Field(i)
		fieldValue := dataStruct.Field(i)

		switch fieldType.Type.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			str := strconv.FormatInt(fieldValue.Int(), 10)
			err := r.pool.Do(radix.Cmd(nil, "HSet", tableName, fieldType.Name, str))
			MiaError.CheckError(err)
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			str := strconv.FormatUint(fieldValue.Uint(), 10)
			err := r.pool.Do(radix.Cmd(nil, "Hset", tableName, fieldType.Name, str))
			MiaError.CheckError(err)
		case reflect.Float32, reflect.Float64:
			str := strconv.FormatFloat(fieldValue.Float(), 'f', -1, 64)
			err := r.pool.Do(radix.Cmd(nil, "HSet", tableName, fieldType.Name, str))
			MiaError.CheckError(err)

		case reflect.String:
			//client.HSet(tableName, fieldType.Name, fieldValue.String())
			err := r.pool.Do(radix.Cmd(nil, "HSet", tableName, fieldType.Name, fieldValue.String()))
			MiaError.CheckError(err)
		//时间类型
		case reflect.Struct:
			str := strconv.FormatInt(fieldValue.Interface().(time.Time).Unix(), 10)
			// client.HSet(tableName, fieldType.Name, str)
			err := r.pool.Do(radix.Cmd(nil, "HSet", tableName, fieldType.Name, str))
			MiaError.CheckError(err)
		case reflect.Bool:
			if fieldValue.Bool() {
				//  client.HSet(tableName, fieldType.Name, "1")
				err := r.pool.Do(radix.Cmd(nil, "HSet", tableName, fieldType.Name, gconv.IntString(1)))
				MiaError.CheckError(err)
			} else {
				// client.HSet(tableName, fieldType.Name, "0")
				err := r.pool.Do(radix.Cmd(nil, "HSet", tableName, fieldType.Name, gconv.IntString(0)))
				MiaError.CheckError(err)
			}
		case reflect.Slice:
			if fieldType.Type.Elem().Kind() == reflect.Uint8 {
				// client.HSet(tableName, fieldType.Name, string(fieldValue.Interface().([]byte)))
				err := r.pool.Do(radix.Cmd(nil, "HSet", tableName, fieldType.Name, string(fieldValue.Interface().([]byte))))
				MiaError.CheckError(err)
			}
		}
	}
}

//LoadFromRedis 从redis中读取一个结构
func (r *RadixDriver) LoadFromRedis(key string, info interface{}) (err error) {
	result := r.Exists(key)
	if !result {
		return errors.New("key not existst")
	}
	dataStruct := reflect.Indirect(reflect.ValueOf(info))
	dataStructType := dataStruct.Type()
	var resultstringmap map[string]string
	err = r.pool.Do(radix.Cmd(&resultstringmap, "HGETALL", key))
	MiaError.CheckError(err)
	for key, value := range resultstringmap {
		for i := 0; i < dataStructType.NumField(); i++ {
			fieldType := dataStructType.Field(i)
			fieldValue := dataStruct.Field(i)
			if fieldType.Name == key {
				switch fieldType.Type.Kind() {
				case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
					n, _ := strconv.Atoi(value)
					fieldValue.SetInt(int64(n))
				case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					n, _ := strconv.Atoi(value)
					fieldValue.SetUint(uint64(n))
				case reflect.Float32:
					n, _ := strconv.ParseFloat(value, 32)
					fieldValue.SetFloat(n)
				case reflect.Float64:
					n, _ := strconv.ParseFloat(value, 64)
					fieldValue.SetFloat(n)
				case reflect.String:
					fieldValue.SetString(value)
				case reflect.Bool:
					fieldValue.SetBool(value == "1")
				case reflect.Slice:
					fieldValue.SetBytes([]byte(value))
				}
				break
			}
		}
	}
	return nil
}
func (r *RadixDriver) GetFieldFromRedis(key string, info interface{}, field string) (err error) {
	result := r.Exists(key)
	if !result {
		return errors.New("key not existst")
	}
	dataStruct := reflect.Indirect(reflect.ValueOf(info))
	dataStructType := dataStruct.Type()
	for i := 0; i < dataStructType.NumField(); i++ {
		fieldType := dataStructType.Field(i)
		if fieldType.Name == field {
			fieldValue := dataStruct.Field(i)
			var v string
			err = r.pool.Do(radix.Cmd(&r, "HGET", key, field))
			switch fieldType.Type.Kind() {
			case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
				n, _ := strconv.Atoi(v)
				fieldValue.SetInt(int64(n))
			case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				n, _ := strconv.Atoi(v)
				fieldValue.SetUint(uint64(n))
			case reflect.Float32:
				n, _ := strconv.ParseFloat(v, 32)
				fieldValue.SetFloat(n)
			case reflect.Float64:
				n, _ := strconv.ParseFloat(v, 64)
				fieldValue.SetFloat(n)
			case reflect.String:
				fieldValue.SetString(v)
			case reflect.Bool:
				fieldValue.SetBool(v == "1")
			case reflect.Slice:
				fieldValue.SetBytes([]byte(v))
			}
			return nil
		}
	}

	return errors.New("field not existst")
}

//SetFieldFromRedis 从redis中保存一个结构指定的字段
func (r *RadixDriver) SetFieldFromRedis(key string, info interface{}, field string) error {
	result := r.Exists(key)
	if !result {
		return errors.New("key not existst")
	}
	dataStruct := reflect.Indirect(reflect.ValueOf(info))
	dataStructType := dataStruct.Type()
	for i := 0; i < dataStructType.NumField(); i++ {
		fieldType := dataStructType.Field(i)
		if fieldType.Name == field {
			fieldValue := dataStruct.Field(i)
			switch fieldType.Type.Kind() {
			case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
				str := strconv.FormatInt(fieldValue.Int(), 10)
				r.pool.Do(radix.Cmd(nil, "hset", key, fieldType.Name, str))
			case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				str := strconv.FormatUint(fieldValue.Uint(), 10)
				r.pool.Do(radix.Cmd(nil, "hset", key, fieldType.Name, str))
			case reflect.Float32, reflect.Float64:
				str := strconv.FormatFloat(fieldValue.Float(), 'f', -1, 64)
				r.pool.Do(radix.Cmd(nil, "hset", key, fieldType.Name, str))
			case reflect.String:
				//synSender.Do(currentCtx,"hset",key,fieldType.Name, fieldValue.String())
				r.pool.Do(radix.Cmd(nil, "hset", key, fieldType.Name, fieldValue.String()))
			//时间类型
			case reflect.Struct:
				str := strconv.FormatInt(fieldValue.Interface().(time.Time).Unix(), 10)
				//client.HSet(tableName, fieldType.Name, str)
				//  synSender.Do(currentCtx,"hset",key,fieldType.Name, str)
				r.pool.Do(radix.Cmd(nil, "hset", key, fieldType.Name, str))
			case reflect.Bool:
				if fieldValue.Bool() {
					///client.HSet(tableName, fieldType.Name, "1")
					r.pool.Do(radix.Cmd(nil, "hset", key, fieldType.Name, "1"))

				} else {
					r.pool.Do(radix.Cmd(nil, "hset", key, fieldType.Name, "0"))
				}
			case reflect.Slice:
				if fieldType.Type.Elem().Kind() == reflect.Uint8 {
					//client.HSet(tableName, fieldType.Name, string(fieldValue.Interface().([]byte)))
					r.pool.Do(radix.Cmd(nil, "hset", key, fieldType.Name, string(fieldValue.Interface().([]byte))))
				}
			}
			return nil
		}
	}
	return errors.New("field not existst")
}

func (self *RadixDriver) Expire(key string, second int) {
	err := self.pool.Do(radix.Cmd(nil, "EXPIRE", key, gconv.IntString(second)))
	MiaError.CheckError(err)
}
func (self *RadixDriver) ExpireAt(key string, at time.Time) {
	err := self.pool.Do(radix.Cmd(nil, "EXPIREAT", key, gconv.Int64String(at.Unix())))
	MiaError.CheckError(err)
}

func (self *RadixDriver) Pexpire(key string, duration time.Duration) {
	err := self.pool.Do(radix.Cmd(nil, "PEXPIRE", key, gconv.Int64String(duration.Milliseconds())))
	MiaError.CheckError(err)
}

func (self *RadixDriver) PexpireAt(key string, at time.Time) {
	err := self.pool.Do(radix.Cmd(nil, "PEXPIREAT", key, gconv.Int64String(at.UnixNano()/int64(time.Millisecond))))
	MiaError.CheckError(err)
}

func (self *RadixDriver) Randomkey() (key string) {
	data := radix.MaybeNil{Rcv: &key}
	err := self.pool.Do(radix.Cmd(&data, "RANDOMKEY"))
	MiaError.CheckError(err)
	return
}

func (self *RadixDriver) Rename(oldKey string, newKey string) (err error) {
	return self.pool.Do(radix.Cmd(nil, "RENAME", oldKey, newKey))
}
func (self *RadixDriver) RenameNX(oldKey string, newKey string) (done bool, err error) {
	data := radix.MaybeNil{Rcv: &done}
	err = self.pool.Do(radix.Cmd(&data, "RENAMENX", oldKey, newKey))
	return
}
func CreateRedis(config *RedisConfig) (result *RadixDriver) {
	if config != nil {
		result = new(RadixDriver)
		result.IsCheckReconnect = true
		result.ReConnect(*config)
		return
	} else {
		return nil
	}

}
func PingPongRedisServer(result *RadixDriver, ctx context.Context) {
	go func(resulttt *RadixDriver) {
		timer1 := time.NewTicker(time.Duration(30 * int64(time.Second)))
	LOOP:
		for {
			select {
			case <-timer1.C:
				{
					if resulttt.IsCheckReconnect {
						result, err := resulttt.PingPong()
						if err != nil && result != true {
							MiaLog.CError("mysql connect fail,err:", err)
							MiaLog.CInfo("reconnect beginning...")
							resulttt.IsCheckReconnect = false
							resulttt.ReConnect(resulttt.Config)
						}
					}
				}
			case <-ctx.Done():
				{
					break LOOP
				}
			}
		}
	}(result)
}
