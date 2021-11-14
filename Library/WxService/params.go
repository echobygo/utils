package WxService

type Params struct {
	value map[string]interface{}
}

func NewParams() *Params {
	return &Params{
		value: make(map[string]interface{}),
	}
}

func (p *Params) Values() map[string]interface{} {
	return p.value
}

func (p *Params) SetString(key string, val string) *Params {
	p.value[key] = val
	return p
}

func (p *Params) SetInt(key string, val int) *Params {
	p.value[key] = val
	return p
}

func (p *Params) SetInt32(key string, val int32) *Params {
	p.value[key] = val
	return p
}

func (p *Params) SetInt64(key string, val int64) *Params {
	p.value[key] = val
	return p
}

func (p *Params) GetString(key string) (val string) {
	val, _ = p.value[key].(string)
	return val
}

func (p *Params) Get(key string) interface{} {
	return p.value[key]
}

func (p *Params) GetInt(key string) (val int) {
	val, _ = p.value[key].(int)
	return val
}

func (p *Params) GetInt32(key string) (val int32) {
	val, _ = p.value[key].(int32)
	return val
}

func (p *Params) GetInt64(key string) (val int64) {
	val, _ = p.value[key].(int64)
	return val
}
