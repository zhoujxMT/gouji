package frame

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	jpg "image/jpeg"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	. "kelei.com/utils/common"
	"kelei.com/utils/logger"
)

var (
	token = ""
)

var (
	AppID               = "wx3d2d58c41fa2ed72"
	Secret              = "a1b741695124f50523b8d5acf0863686"
	AccessTokenFetchUrl = "https://api.weixin.qq.com/cgi-bin/token"
)

type AccessTokenResponse struct {
	AccessToken string  `json:"access_token"`
	ExpiresIn   float64 `json:"expires_in"`
}

type AccessTokenErrorResponse struct {
	Errcode float64
	Errmsg  string
}

func initSDK() {
	initToken()
}

func initToken() {
	var err error
	token, err = getAccessToken()
	logger.CheckError(err)
}

//获取wx_AccessToken 拼接get请求 解析返回json结果 返回 AccessToken和err
func getAccessToken() (string, error) {
	requestLine := strings.Join([]string{AccessTokenFetchUrl,
		"?grant_type=client_credential",
		"&appid=",
		AppID,
		"&secret=",
		Secret}, "")

	resp, err := http.Get(requestLine)
	if err != nil || resp.StatusCode != http.StatusOK {
		logger.Errorf("发送get请求获取 atoken 错误 : %s", err.Error())
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("发送get请求获取 atoken 读取返回body错误 : %s", err.Error())
		return "", err
	}

	if bytes.Contains(body, []byte("access_token")) {
		atr := AccessTokenResponse{}
		err = json.Unmarshal(body, &atr)
		if err != nil {
			logger.Errorf("发送get请求获取 atoken 返回数据json解析错误 : %s", err.Error())
			return "", err
		}
		return atr.AccessToken, nil
	} else {
		ater := AccessTokenErrorResponse{}
		err = json.Unmarshal(body, &ater)
		if err != nil {
			logger.Errorf("发送get请求获取 微信返回 的错误信息 : %s", err.Error())
			return "", err
		}
		return "", fmt.Errorf("%s", ater.Errmsg)
	}
}

type WxArgs struct {
	Scene string `json:"scene"`
	Width int    `json:"width"`
}

//生成二维码图片
func (c *Client) createQRCodeImg(roomid string) (image.Image, error) {
	var img image.Image
	//HTTP POST请求
	my_url := "https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=" + token
	wxArgs := WxArgs{roomid, 100}
	b, _ := json.Marshal(wxArgs)
	str := string(b)
	req, err := http.Post(my_url, "application/json", strings.NewReader(str)) //这里定义链接和post的数据
	if err != nil {
		return img, err
	}
	defer req.Body.Close()
	body, err2 := ioutil.ReadAll(req.Body)
	if err2 != nil {
		return img, err2
	}
	img, err = jpg.Decode(bytes.NewReader(body))
	if err != nil {
		initToken()
		return img, err
	}
	return img, err
}

/*
获取房间二维码
*/
func (c *Client) getRoomQRCode() string {
	res := ""
	content := "0GetMatchingRoomID"
	roomid := c.handleClientData([]byte(content))
	if roomid == "-101" {
		return roomid
	}
	img, err := c.createQRCodeImg(roomid)
	if err != nil {
		img, err = c.createQRCodeImg(roomid)
	}
	if err != nil {
		logger.Errorf(err.Error())
		return Res_Unknown
	}
	qrPath := "/var/www/html/gouji/qr/room/"
	fileName := fmt.Sprintf("%s%s.png", qrPath, roomid)
	_, err2 := os.Stat(fileName)
	if err2 != nil {
		file, err3 := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err3 != nil {
			logger.Errorf(err3.Error())
			return Res_Unknown
		}
		defer file.Close()
		jpg.Encode(file, img, &jpg.Options{80})
	}
	res = "http://118.190.205.223:8801/gouji/qr/room/" + roomid + ".png"
	return res
}
