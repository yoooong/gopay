package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/yoooong/gopay/common"
	"github.com/yoooong/gopay/util"
)

var defaultWechatWebH5Client *WechatWebH5Client

func InitWxWebH5Client(c *WechatWebH5Client) {
	defaultWechatWebH5Client = c
}

func DefaultWechatWebH5Client() *WechatWebH5Client {
	return defaultWechatWebH5Client
}

// WechatWebClient 微信公众号支付
type WechatWebH5Client struct {
	AppID       string       // 公众账号ID
	MchID       string       // 商户号ID
	Key         string       // 密钥
	PrivateKey  []byte       // 私钥文件内容
	PublicKey   []byte       // 公钥文件内容
	httpsClient *HTTPSClient // 双向证书链接
}

// Pay 支付
func (this *WechatWebH5Client) Pay(charge *common.Charge) (map[string]string, error) {
	var m = make(map[string]string)
	m["appid"] = this.AppID
	m["mch_id"] = this.MchID
	m["nonce_str"] = util.RandomStr()
	m["body"] = TruncatedText(charge.Describe, 32)
	m["out_trade_no"] = charge.TradeNum
	m["total_fee"] = WechatMoneyFeeToString(charge.MoneyFee)
	m["spbill_create_ip"] = util.LocalIP()
	m["notify_url"] = charge.CallbackURL
	m["trade_type"] = "MWEB"
	m["openid"] = charge.OpenID
	m["sign_type"] = "MD5"
	m["scene_info"] = charge.SceneInfo

	sign, err := WechatGenSign(this.Key, m)
	if err != nil {
		return map[string]string{}, err
	}
	m["sign"] = sign

	// 转出xml结构
	xmlRe, err := PostWechat("https://api.mch.weixin.qq.com/pay/unifiedorder", m, nil)
	if err != nil {
		return map[string]string{}, err
	}

	var c = make(map[string]string)
	c["appId"] = this.AppID
	c["timeStamp"] = fmt.Sprintf("%d", time.Now().Unix())
	c["nonceStr"] = util.RandomStr()
	c["package"] = fmt.Sprintf("prepay_id=%s", xmlRe.PrepayID)
	c["signType"] = "MD5"
	sign2, err := WechatGenSign(this.Key, c)
	if err != nil {
		return map[string]string{}, errors.New("WechatWeb: " + err.Error())
	}
	c["paySign"] = sign2

	return c, nil
}

// 关闭订单
func (this *WechatWebH5Client) CloseOrder(outTradeNo string) (common.WeChatQueryResult, error) {
	return WachatCloseOrder(this.AppID, this.MchID, this.Key, outTradeNo)
}

// 支付到用户的微信账号
func (this *WechatWebH5Client) PayToClient(charge *common.Charge) (map[string]string, error) {
	return WachatCompanyChange(this.AppID, this.MchID, this.Key, this.httpsClient, charge)
}

// QueryOrder 查询订单
func (this *WechatWebH5Client) QueryOrder(tradeNum string) (common.WeChatQueryResult, error) {
	return WachatQueryOrder(this.AppID, this.MchID, this.Key, tradeNum)
}
