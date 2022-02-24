package WxService

import (
	"encoding/xml"
	"io"
	"crypto/sha1"
	"encoding/hex"
	"encoding/base64"
	"crypto/aes"
	"crypto/cipher"
)

const (
	Oauth2Url                    = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	Code2SessionUrl              = "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code"
	UnifiedOrderUrl              = "https://api.mch.weixin.qq.com/pay/unifiedorder"
	SandboxUnifiedOrderUrl       = "https://api.mch.weixin.qq.com/sandboxnew/pay/unifiedorder"
	RefundUrl                    = "https://api.mch.weixin.qq.com/secapi/pay/refund"
	SandboxRefundUrl             = "https://api.mch.weixin.qq.com/sandboxnew/secapi/pay/refund"
	CloseOrderUrl                = "https://api.mch.weixin.qq.com/pay/closeorder"
	SandboxCloseOrderUrl         = "https://api.mch.weixin.qq.com/sandboxnew/pay/closeorder"
	GetAccessTokenUrl            = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
	GetDailyRetainUrl            = "https://api.weixin.qq.com/datacube/getweanalysisappiddailyretaininfo?access_token=%s"
	GetMonthlyRetainUrl          = "https://api.weixin.qq.com/datacube/getweanalysisappidmonthlyretaininfo?access_token=%s"
	GetWeeklyRetainUrl           = "https://api.weixin.qq.com/datacube/getweanalysisappidweeklyretaininfo?access_token=%s"
	GetDailySummaryUrl           = "https://api.weixin.qq.com/datacube/getweanalysisappiddailysummarytrend?access_token=%s"
	GetDailyVisitTrendUrl        = "https://api.weixin.qq.com/datacube/getweanalysisappiddailyvisittrend?access_token=%s"
	GetWeeklyVisitTrendUrl       = "https://api.weixin.qq.com/datacube/getweanalysisappidweeklyvisittrend?access_token=%s"
	GetMonthlyVisitTrendUrl      = "https://api.weixin.qq.com/datacube/getweanalysisappidmonthlyvisittrend?access_token=%s"
	GetDailyUserPortraitUrl      = "https://api.weixin.qq.com/datacube/getweanalysisappiduserportrait?access_token=%s"
	GetDailyVisitDistributionUrl = "https://api.weixin.qq.com/datacube/getweanalysisappidvisitdistribution?access_token=%s"
	GetDailyVisitPageUrl         = "https://api.weixin.qq.com/datacube/getweanalysisappidvisitpage?access_token=%s"
	GetSandboxSignKeyUrl         = "https://api.mch.weixin.qq.com/sandboxnew/pay/getsignkey"
)
type UserSession struct {
	OpenId     string `json:"openid"`      // 用户唯一标识
	SessionKey string `json:"session_key"` // 会话密钥
	UnionId    string `json:"unionid"`     // 用户在开放平台的唯一标识符，在满足 UnionID 下发条件的情况下会返回，详见 UnionID 机制说明。
	Errcode    int    `json:"errcode"`     // 错误码
	Errmsg     string `json:"errmsg"`      // 错误信息
}
type Xml map[string]interface{}

type xmlMapEntry struct {
	XMLName xml.Name
	Value   interface{} `xml:",chardata"`
}

type xmlMapStrEntry struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

func (x Xml) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(x) == 0 {
		return nil
	}

	err := e.EncodeToken(start)
	if err != nil {
		return err
	}

	for k, v := range x {
		if err := e.Encode(xmlMapEntry{XMLName: xml.Name{Local: k}, Value: v}); err != nil {
			return err
		}
	}

	return e.EncodeToken(start.End())
}

func (x *Xml) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*x = Xml{}
	for {
		var e xmlMapStrEntry

		err := d.Decode(&e)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		(*x)[e.XMLName.Local] = e.Value
	}

	return nil
}
func Sha1s(s string) string {
	r := sha1.Sum([]byte(s))
	return hex.EncodeToString(r[:])
}
func AesCBCDncrypt(encryptData, key, iv []byte) ([]byte, error) {
	var aesBlockDecrypter cipher.Block
	aesBlockDecrypter, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decrypted := make([]byte, len(encryptData))
	aesDecrypter := cipher.NewCBCDecrypter(aesBlockDecrypter, iv)
	aesDecrypter.CryptBlocks(decrypted, encryptData)

	return decrypted, nil
}
func DecryptData(encryptedData, key, iv string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}
	iKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}
	iIv, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return nil, err
	}
	dnData, err := AesCBCDncrypt(data, iKey, iIv)
	if err != nil {
		return nil, err
	}
	return dnData, nil
}