package WxService

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

// 客户端
type Client interface {
	Err() error
	Client() client                                                                           // 获取客户端实例的拷贝
	Code2Session(code string) (*UserSession, error)                                                 // 获取openid
	Oauth2(code string) (openid, at string)                                                   // 网页登录
	UnifiedOrder(p *Params) (prepayId string)                                                 // 统一下单
	Refund(p *Params, keyPath, certPath string)                                                    // 退款
	CloseOrder(p *Params)                                                                     // 关闭订单
	GetAccessToken() (accessToken string)                                                     // 获取access token
	GetDailyRetain(accessToken, date string) (result map[string]interface{})                  // 获取日留存
	GetMonthlyRetain(accessToken string, year int, month int) (result map[string]interface{}) // 获取周留存
	GetDailySummary(accessToken, date string) (result map[string]interface{})                 // 获取日统计
	GetDailyVisitTrend(accessToken, date string) (result map[string]interface{})              // 获取日访问趋势
	GetWeeklyVisitTrend(accessToken string) (result map[string]interface{})                   // 获取周访问趋势
	GetMonthlyVisitTrend(accessToken string, year, month int) (result map[string]interface{}) // 获取月访问趋势
	GetDailyUserPortrait(accessToken, date string) (result map[string]interface{})            // 获取日用户画像
	GetDailyVisitDistribution(accessToken, date string) (result map[string]interface{})       // 获取日访问分布
	GetDailyVisitPage(accessToken, date string) (result map[string]interface{})               // 获取日访问页面
	GetSandboxSignKey(p *Params) (signKey string)                                             // 获取沙箱环境签名key
}

type client struct {
	AppID     string // app id
	AppSecret string // app密钥
	MchID     string // 商户号
	ApiKey    string // api密钥
	IsSandBox bool   // 是否沙箱环境
	err       error
}

// 配置类
type ClientConfig struct {
	AppID     string // app id
	AppSecret string // app密钥
	MchID     string // 商户号
	ApiKey    string // 支付平台api密钥
	IsSandBox bool   // 是否沙箱环境
}

// 创建新的客户端
func NewClient(cfg *ClientConfig) Client {
	return &client{
		AppID:     cfg.AppID,
		AppSecret: cfg.AppSecret,
		MchID:     cfg.MchID,
		ApiKey:    cfg.ApiKey,
		IsSandBox: cfg.IsSandBox,
	}
}

func (c client) Err() error {
	return c.err
}

func (c client) Client() client {
	return c
}


// 登录api
func (c *client) Code2Session(code string) (*UserSession, error) {
	var (
		res  *http.Response
		body []byte
	)
	res, c.err = http.Get(fmt.Sprintf(Code2SessionUrl, c.AppID, c.AppSecret, code))
	if c.err != nil {
		return nil,c.err
	}

	body, c.err = ioutil.ReadAll(res.Body)
	if c.err != nil {
		return nil,c.err
	}

	reply := &UserSession{}
	err := json.Unmarshal(body, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

// web登录api
func (c *client) Oauth2(code string) (openid, at string) {
	var (
		res  *http.Response
		body []byte
		ok   bool
		data = make(map[string]interface{})
	)
	res, c.err = http.Get(fmt.Sprintf(Oauth2Url, c.AppID, c.AppSecret, code))
	if c.err != nil {
		return
	}

	body, c.err = ioutil.ReadAll(res.Body)
	if c.err != nil {
		return
	}

	c.err = json.Unmarshal(body, &data)
	if c.err != nil {
		return
	}

	openid, ok = data["openid"].(string)
	if !ok {
		c.err = fmt.Errorf(ErrMsgWxRemote, "获取openid失败！")
	}

	at, ok = data["access_token"].(string)
	if !ok {
		c.err = fmt.Errorf(ErrMsgWxRemote, "获取access_token失败！")
	}
	return
}

// 统一下单api
func (c *client) UnifiedOrder(p *Params) (prepayId string) {
	var (
		ok         bool
		err        error
		buf        []byte
		res        *http.Response
		resBody    []byte
		returnCode string
		resultCode string
		url        string
		resData    = make(map[string]interface{})
	)

	// 签名
	if c.IsSandBox {
		param := NewParams()
		param.SetString("mch_id", c.Client().MchID).
			SetString("nonce_str", GeneNonceStr(32))

		// 获取沙箱环境签名key
		signKey := c.GetSandboxSignKey(param)

		c.signParamMD5(p, signKey)
		if c.err != nil {
			return
		}

		url = SandboxUnifiedOrderUrl
	} else {
		c.signParamMD5(p, c.ApiKey)
		url = UnifiedOrderUrl
	}

	buf, err = xml.Marshal(Xml(p.value))
	if err != nil {
		c.err = err
		return
	}

	res, err = http.Post(url, "application/xml", bytes.NewReader(buf))
	if err != nil {
		c.err = err
		return
	}

	resBody, err = ioutil.ReadAll(res.Body)
	if err != nil {
		c.err = err
		return
	}

	err = xml.Unmarshal(resBody, (*Xml)(&resData))
	if err != nil {
		c.err = err
		return
	}

	if resData["return_code"] == nil || resData["return_msg"] == nil {
		c.err = fmt.Errorf(ErrMsgWxRemote, "响应中没有return_code或return_msg！")
		return
	}

	returnCode, ok = resData["return_code"].(string)
	if !ok {
		c.err = fmt.Errorf(ErrMsgWxRemote, "return_code类型错误！")
		return
	}

	if returnCode != "SUCCESS" {
		c.err = fmt.Errorf(ErrMsgWxRemote, resData["return_msg"])
		return
	}

	resultCode, ok = resData["result_code"].(string)
	if !ok {
		c.err = fmt.Errorf(ErrMsgWxRemote, "result_code类型错误！")
		return
	}

	if resultCode != "SUCCESS" {
		c.err = fmt.Errorf(ErrMsgWxRemote, resData["err_code_des"].(string))
		return
	}

	if resData["prepay_id"] == nil {
		c.err = fmt.Errorf(ErrMsgWxRemote, "响应中没有prepay_id！")
		return
	}

	prepayId, ok = resData["prepay_id"].(string)
	if !ok {
		c.err = fmt.Errorf(ErrMsgWxRemote, "prepay_id类型错误！")
		return
	}
	return resData["prepay_id"].(string)
}

// 获取沙箱key
func (c client) GetSandboxSignKey(p *Params) (signKey string) {
	var (
		ok         bool
		err        error
		buf        []byte
		res        *http.Response
		resBody    []byte
		returnCode string
		resData    = make(map[string]interface{})
	)

	c.signParamMD5(p, c.ApiKey)

	buf, err = xml.Marshal(Xml(p.value))
	if err != nil {
		c.err = err
		return
	}

	res, err = http.Post(GetSandboxSignKeyUrl, "application/xml", bytes.NewReader(buf))
	if err != nil {
		c.err = err
		return
	}

	resBody, err = ioutil.ReadAll(res.Body)
	if err != nil {
		c.err = err
		return
	}

	err = xml.Unmarshal(resBody, (*Xml)(&resData))
	if err != nil {
		c.err = err
		return
	}

	if resData["return_code"] == nil || resData["return_msg"] == nil {
		c.err = fmt.Errorf(ErrMsgWxRemote, "响应中没有return_code或return_msg！")
		return
	}

	returnCode, ok = resData["return_code"].(string)
	if !ok {
		c.err = fmt.Errorf(ErrMsgWxRemote, "return_code类型错误！")
		return
	}

	if returnCode != "SUCCESS" {
		c.err = fmt.Errorf(ErrMsgWxRemote, resData["return_msg"])
		return
	}

	signKey, ok = resData["sandbox_signkey"].(string)
	if !ok {
		c.err = fmt.Errorf(ErrMsgWxRemote, "签名密钥类型错误！")
	}

	return
}

// 退款
func (c *client) Refund(p *Params, keyPath, certPath string) {
	var (
		ok         bool
		err        error
		buf        []byte
		res        *http.Response
		resBody    []byte
		cert       tls.Certificate
		returnCode string
		resultCode string
		url        string
		resData    = make(map[string]interface{})
	)

	// 签名
	if c.IsSandBox {
		param := NewParams()
		param.SetString("mch_id", c.Client().MchID).
			SetString("nonce_str", GeneNonceStr(32))

		// 获取沙箱环境签名key
		signKey := c.GetSandboxSignKey(param)

		c.signParamMD5(p, signKey)

		url = SandboxRefundUrl
	} else {
		c.signParamMD5(p, c.ApiKey)
		url = RefundUrl
	}

	buf, err = xml.Marshal(Xml(p.value))
	if err != nil {
		c.err = err
		return
	}

	cert, err = tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		c.err = err
		return
	}

	cl := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	res, err = cl.Post(url, "application/xml", bytes.NewReader(buf))
	if err != nil {
		c.err = err
		return
	}

	resBody, err = ioutil.ReadAll(res.Body)
	if err != nil {
		c.err = err
		return
	}

	err = xml.Unmarshal(resBody, (*Xml)(&resData))
	if err != nil {
		c.err = err
		return
	}

	if resData["return_code"] == nil || resData["return_msg"] == nil {
		c.err = fmt.Errorf(ErrMsgWxRemote, "响应中没有return_code或return_msg！")
		return
	}

	returnCode, ok = resData["return_code"].(string)
	if !ok {
		c.err = fmt.Errorf(ErrMsgWxRemote, "return_code类型错误！")
		return
	}

	if resData["return_code"] == nil || resData["return_msg"] == nil {
		c.err = fmt.Errorf(ErrMsgWxRemote, "响应中没有return_code或return_msg！")
		return
	}

	returnCode, ok = resData["return_code"].(string)
	if !ok {
		c.err = fmt.Errorf(ErrMsgWxRemote, "return_code类型错误！")
		return
	}

	if returnCode != "SUCCESS" {
		c.err = fmt.Errorf(ErrMsgWxRemote, resData["return_msg"])
		return
	}

	if resData["result_code"] == nil {
		c.err = fmt.Errorf(ErrMsgWxRemote, "响应中没有result_code！")
		return
	}

	resultCode, ok = resData["result_code"].(string)
	if !ok {
		c.err = fmt.Errorf(ErrMsgWxRemote, "result_code类型错误！")
		return
	}

	if resultCode != "SUCCESS" {
		c.err = fmt.Errorf(ErrMsgWxRemote, resData["err_code_des"].(string))
		return
	}
}

// 关闭订单
func (c *client) CloseOrder(p *Params) {
	var (
		ok         bool
		err        error
		buf        []byte
		res        *http.Response
		resBody    []byte
		returnCode string
		resultCode string
		url        string
		resData    = make(map[string]interface{})
	)

	// 签名
	if c.IsSandBox {
		param := NewParams()
		param.SetString("mch_id", c.Client().MchID).
			SetString("nonce_str", GeneNonceStr(32))

		// 获取沙箱环境签名key
		signKey := c.GetSandboxSignKey(param)

		c.signParamMD5(p, signKey)

		url = SandboxCloseOrderUrl
	} else {
		c.signParamMD5(p, c.ApiKey)
		url = CloseOrderUrl
	}

	buf, err = xml.Marshal(Xml(p.value))
	if err != nil {
		c.err = err
		return
	}

	res, err = http.Post(url, "application/xml", bytes.NewReader(buf))
	if err != nil {
		c.err = err
		return
	}

	resBody, err = ioutil.ReadAll(res.Body)
	if err != nil {
		c.err = err
		return
	}

	err = xml.Unmarshal(resBody, &resData)
	if err != nil {
		c.err = err
		return
	}

	if resData["return_code"] == nil || resData["return_msg"] == nil {
		c.err = fmt.Errorf(ErrMsgWxRemote, "响应中没有return_code或return_msg！")
		return
	}

	returnCode, ok = resData["return_code"].(string)
	if !ok {
		c.err = fmt.Errorf(ErrMsgWxRemote, "return_code类型错误！")
		return
	}

	if resData["return_code"] == nil || resData["return_msg"] == nil {
		c.err = fmt.Errorf(ErrMsgWxRemote, "响应中没有return_code或return_msg！")
		return
	}

	returnCode, ok = resData["return_code"].(string)
	if !ok {
		c.err = fmt.Errorf(ErrMsgWxRemote, "return_code类型错误！")
		return
	}

	if returnCode != "SUCCESS" {
		c.err = fmt.Errorf(ErrMsgWxRemote, resData["return_msg"])
		return
	}

	if resData["result_code"] == nil {
		c.err = fmt.Errorf(ErrMsgWxRemote, "响应中没有result_code！")
		return
	}

	resultCode, ok = resData["result_code"].(string)
	if !ok {
		c.err = fmt.Errorf(ErrMsgWxRemote, "result_code类型错误！")
		return
	}

	if resultCode != "SUCCESS" {
		c.err = fmt.Errorf(ErrMsgWxRemote, resData["err_code_des"].(string))
		return
	}
}

// 获取接口调用凭证
func (c *client) GetAccessToken() (accessToken string) {
	var (
		ok   bool
		err  error
		resp *http.Response
		body []byte
		data map[string]interface{}
	)

	resp, err = http.Post(
		fmt.Sprintf(GetAccessTokenUrl, c.AppID, c.AppSecret),
		"application/x-www-form-urlencoded", nil)
	if err != nil {
		c.err = err
		return
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		c.err = err
		return
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		c.err = err
		return
	}

	if data["access_token"] == nil {
		c.err = fmt.Errorf(ErrMsgWxRemote, "响应中没有access_token")
		return
	}

	accessToken, ok = data["access_token"].(string)
	if !ok {
		c.err = fmt.Errorf(ErrMsgWxRemote, "access_token类型错误！")
		return
	}

	return
}

// 获取日访问留存
func (c *client) GetDailyRetain(accessToken, date string) (result map[string]interface{}) {
	var (
		err     error
		reqBody []byte
		resBody []byte
		r       *bytes.Reader
		res     *http.Response
		resData map[string]interface{}
		req     = struct {
			BeginDate string `json:"begin_date"`
			EndDate   string `json:"end_date"`
		}{
			date,
			date,
		}
	)

	reqBody, err = json.Marshal(&req)
	if err != nil {
		c.err = err
		return
	}

	r = bytes.NewReader(reqBody)

	res, err = http.Post(
		fmt.Sprintf(GetDailyRetainUrl, accessToken),
		"application/json", r)
	if err != nil {
		c.err = err
		return
	}

	resBody, err = ioutil.ReadAll(res.Body)
	if err != nil {
		c.err = err
		return
	}

	err = json.Unmarshal(resBody, &resData)
	if err != nil {
		c.err = err
		return
	}

	result = resData
	return
}

// 获取月访问留存
func (c *client) GetMonthlyRetain(accessToken string, year int, month int) (result map[string]interface{}) {
	var (
		err     error
		begin   string
		end     string
		reqBody []byte
		r       *bytes.Reader
		res     *http.Response
		resBody []byte
		resData map[string]interface{}

		req = struct {
			BeginDate string `json:"begin_date"`
			EndDate   string `json:"end_date"`
		}{}
	)

	begin, end = GetBeginAndEndByMonth(year, month)
	req.BeginDate = begin
	req.EndDate = end

	reqBody, err = json.Marshal(&req)
	if err != nil {
		c.err = err
		return
	}

	r = bytes.NewReader(reqBody)

	res, err = http.Post(
		fmt.Sprintf(GetMonthlyRetainUrl, accessToken),
		"application/json", r)
	if err != nil {
		c.err = err
		return
	}

	resBody, err = ioutil.ReadAll(res.Body)
	if err != nil {
		c.err = err
		return
	}

	err = json.Unmarshal(resBody, &resData)
	if err != nil {
		c.err = err
		return
	}

	result = resData
	return
}

// 获取周访问留存
func (c *client) GetWeeklyRetain(accessToken string) (result map[string]interface{}) {
	var (
		err     error
		begin   string
		end     string
		reqBody []byte
		r       *bytes.Reader
		res     *http.Response
		resBody []byte
		resData map[string]interface{}
		req     = struct {
			BeginDate string `json:"begin_date"`
			EndDate   string `json:"end_date"`
		}{}
	)

	begin, end = GetBeginAndEndByWeek()

	req.BeginDate = begin
	req.EndDate = end

	reqBody, err = json.Marshal(&req)
	if err != nil {
		c.err = err
		return
	}

	r = bytes.NewReader(reqBody)

	res, err = http.Post(
		fmt.Sprintf(GetWeeklyRetainUrl, accessToken),
		"application/json", r)
	if err != nil {
		c.err = err
		return
	}

	resBody, err = ioutil.ReadAll(res.Body)
	if err != nil {
		c.err = err
		return
	}

	err = json.Unmarshal(resBody, &resData)
	if err != nil {
		c.err = err
		return
	}

	result = resData
	return
}

// 获取日统计
func (c *client) GetDailySummary(accessToken, date string) (result map[string]interface{}) {
	var (
		err     error
		reqBody []byte
		resBody []byte
		r       *bytes.Reader
		res     *http.Response
		resData map[string]interface{}
		req     = struct {
			BeginDate string `json:"begin_date"`
			EndDate   string `json:"end_date"`
		}{
			date,
			date,
		}
	)

	reqBody, err = json.Marshal(&req)
	if err != nil {
		c.err = err
		return
	}

	r = bytes.NewReader(reqBody)

	res, err = http.Post(
		fmt.Sprintf(GetDailySummaryUrl, accessToken),
		"application/json", r)
	if err != nil {
		c.err = err
		return
	}

	resBody, err = ioutil.ReadAll(res.Body)
	if err != nil {
		c.err = err
		return
	}

	err = json.Unmarshal(resBody, &resData)
	if err != nil {
		c.err = err
		return
	}

	result = resData
	return
}

// 获取日趋势
func (c *client) GetDailyVisitTrend(accessToken, date string) (result map[string]interface{}) {
	var (
		err     error
		reqBody []byte
		resBody []byte
		r       *bytes.Reader
		res     *http.Response
		resData map[string]interface{}
		req     = struct {
			BeginDate string `json:"begin_date"`
			EndDate   string `json:"end_date"`
		}{
			date,
			date,
		}
	)

	reqBody, err = json.Marshal(&req)
	if err != nil {
		c.err = err
		return
	}

	r = bytes.NewReader(reqBody)

	res, err = http.Post(
		fmt.Sprintf(GetDailyVisitTrendUrl, accessToken),
		"application/json", r)
	if err != nil {
		c.err = err
		return
	}

	resBody, err = ioutil.ReadAll(res.Body)
	if err != nil {
		c.err = err
		return
	}

	err = json.Unmarshal(resBody, &resData)
	if err != nil {
		c.err = err
		return
	}

	result = resData
	return
}

// 获取周趋势
func (c *client) GetWeeklyVisitTrend(accessToken string) (result map[string]interface{}) {
	var (
		err     error
		begin   string
		end     string
		reqBody []byte
		r       *bytes.Reader
		res     *http.Response
		resBody []byte
		resData map[string]interface{}
		req     = struct {
			BeginDate string `json:"begin_date"`
			EndDate   string `json:"end_date"`
		}{}
	)

	begin, end = GetBeginAndEndByWeek()

	req.BeginDate = begin
	req.EndDate = end

	reqBody, err = json.Marshal(&req)
	if err != nil {
		c.err = err
		return
	}

	r = bytes.NewReader(reqBody)

	res, err = http.Post(
		fmt.Sprintf(GetWeeklyVisitTrendUrl, accessToken),
		"application/json", r)
	if err != nil {
		c.err = err
		return
	}

	resBody, err = ioutil.ReadAll(res.Body)
	if err != nil {
		c.err = err
		return
	}

	err = json.Unmarshal(resBody, &resData)
	if err != nil {
		c.err = err
		return
	}

	result = resData
	return
}

// 获取月趋势
func (c *client) GetMonthlyVisitTrend(accessToken string, year, month int) (result map[string]interface{}) {
	var (
		err     error
		begin   string
		end     string
		reqBody []byte
		r       *bytes.Reader
		res     *http.Response
		resBody []byte
		resData map[string]interface{}
		req     = struct {
			BeginDate string `json:"begin_date"`
			EndDate   string `json:"end_date"`
		}{}
	)

	begin, end = GetBeginAndEndByMonth(year, month)
	req.BeginDate = begin
	req.EndDate = end

	reqBody, err = json.Marshal(&req)
	if err != nil {
		c.err = err
		return
	}

	r = bytes.NewReader(reqBody)

	res, err = http.Post(
		fmt.Sprintf(GetMonthlyVisitTrendUrl, accessToken),
		"application/json", r)
	if err != nil {
		c.err = err
		return
	}

	resBody, err = ioutil.ReadAll(res.Body)
	if err != nil {
		c.err = err
		return
	}

	err = json.Unmarshal(resBody, &resData)
	if err != nil {
		c.err = err
		return
	}

	result = resData
	return
}

// 获取用户画像
func (c *client) GetDailyUserPortrait(accessToken, date string) (result map[string]interface{}) {
	var (
		err     error
		reqBody []byte
		resBody []byte
		r       *bytes.Reader
		res     *http.Response
		resData map[string]interface{}
		req     = struct {
			BeginDate string `json:"begin_date"`
			EndDate   string `json:"end_date"`
		}{
			date,
			date,
		}
	)

	reqBody, err = json.Marshal(&req)
	if err != nil {
		c.err = err
		return
	}

	r = bytes.NewReader(reqBody)

	res, err = http.Post(
		fmt.Sprintf(GetDailyUserPortraitUrl, accessToken),
		"application/json", r)
	if err != nil {
		c.err = err
		return
	}

	resBody, err = ioutil.ReadAll(res.Body)
	if err != nil {
		c.err = err
		return
	}

	err = json.Unmarshal(resBody, &resData)
	if err != nil {
		c.err = err
		return
	}

	result = resData
	return

}

// 获取用户分布
func (c *client) GetDailyVisitDistribution(accessToken, date string) (result map[string]interface{}) {
	var (
		err     error
		reqBody []byte
		resBody []byte
		r       *bytes.Reader
		res     *http.Response
		resData map[string]interface{}
		req     = struct {
			BeginDate string `json:"begin_date"`
			EndDate   string `json:"end_date"`
		}{
			date,
			date,
		}
	)

	reqBody, err = json.Marshal(&req)
	if err != nil {
		c.err = err
		return
	}

	r = bytes.NewReader(reqBody)

	res, err = http.Post(
		fmt.Sprintf(GetDailyVisitDistributionUrl, accessToken),
		"application/json", r)
	if err != nil {
		c.err = err
		return
	}

	resBody, err = ioutil.ReadAll(res.Body)
	if err != nil {
		c.err = err
		return
	}

	err = json.Unmarshal(resBody, &resData)
	if err != nil {
		c.err = err
		return
	}

	result = resData
	return
}

// 获取页面数据
func (c *client) GetDailyVisitPage(accessToken, date string) (result map[string]interface{}) {
	var (
		err     error
		reqBody []byte
		resBody []byte
		r       *bytes.Reader
		res     *http.Response
		resData map[string]interface{}

		req = struct {
			BeginDate string `json:"begin_date"`
			EndDate   string `json:"end_date"`
		}{
			date,
			date,
		}
	)

	reqBody, err = json.Marshal(&req)
	if err != nil {
		c.err = err
		return
	}

	r = bytes.NewReader(reqBody)

	res, err = http.Post(
		fmt.Sprintf(GetDailyVisitPageUrl, accessToken), "application/json", r)
	if err != nil {
		c.err = err
		return
	}

	resBody, err = ioutil.ReadAll(res.Body)
	if err != nil {
		c.err = err
		return
	}

	err = json.Unmarshal(resBody, &resData)
	if err != nil {
		c.err = err
		return
	}

	result = resData
	return
}

func (c *client) signParamMD5(p *Params, key string) {
	p.value["sign"] = GeneSign(p.value, key)
}
