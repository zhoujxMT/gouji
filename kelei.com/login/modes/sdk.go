package modes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"kelei.com/utils/logger"
)

var (
	AppID  = "wx3d2d58c41fa2ed72"
	Secret = "a1b741695124f50523b8d5acf0863686"
)

type AccessOpenIDResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	Unionid    string `json:"unionid"`
}

type AccessOpenIDErrorResponse struct {
	Errcode float64
	Errmsg  string
}

//获取openid
func (this *Login) getOpenID(jsCode string) (string, string, error) {
	requestLine := strings.Join([]string{"https://api.weixin.qq.com/sns/jscode2session",
		"?grant_type=authorization_code",
		"&appid=",
		AppID,
		"&secret=",
		Secret,
		"&js_code=",
		jsCode}, "")
	resp, err := http.Get(requestLine)
	if err != nil || resp.StatusCode != http.StatusOK {
		logger.Debugf("发送get请求获取 openid 错误:%s", err.Error())
		return "", "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Debugf("发送get请求获取 openid 读取返回body错误:%s", err.Error())
		return "", "", err
	}

	if bytes.Contains(body, []byte("openid")) {
		aor := AccessOpenIDResponse{}
		err = json.Unmarshal(body, &aor)
		if err != nil {
			logger.Debugf("发送get请求获取 openid 返回数据json解析错误:%s", err.Error())
			return "", "", err
		}
		return aor.OpenID, aor.SessionKey, err
	} else {
		aor := AccessOpenIDErrorResponse{}
		err = json.Unmarshal(body, &aor)
		logger.Debugf("发送get请求获取 微信返回 的错误信息 %+v\n", aor)
		if err != nil {
			return "", "", err
		}
		return "", "", fmt.Errorf("%s", aor.Errmsg)
	}

}
