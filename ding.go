package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var dingBaseUrl = "https://api.dingtalk.com"
var dingLoginApi = "https://login.dingtalk.com/oauth2/auth?redirect_uri=%v&response_type=code&client_id=%v&scope=openid&state=&prompt=consent"

type DingConfig struct {
	AppKey    string `json:"appKey"`
	AppSecret string `json:"appSecret"`
}

type Token struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpireIn     string `json:"expireIn"`
	CorpId       string `json:"corpId"`
}

func GetAccessToken(authCode string, appKey string, appSecret string) *Token {
	contentType := "application/json"
	data := fmt.Sprintf(`{"clientId":"%v","clientSecret":"%v", "code": "%v", "grantType": "authorization_code"}`, appKey, appSecret, authCode)
	resp, _ := http.Post(fmt.Sprintf("%v/v1.0/oauth2/userAccessToken", dingBaseUrl), contentType, strings.NewReader(data))
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	token := &Token{}
	_ = json.Unmarshal(body, token)
	fmt.Println("获取token: ", string(body))
	return token
}

func CheckAuthCode(authCode string, appKey string, appSecret string) bool {
	token := GetAccessToken(authCode, appKey, appSecret)
	if token.AccessToken == "" {
		return false
	}
	return true
}

func GetDingAuthUrl(app *App, nowUrl *url.URL) string {
	host := app.ServerName
	appKey := app.DingAppKey
	if app.RedirectHold {
		return fmt.Sprintf(dingLoginApi, url.QueryEscape(getProtocol(app.Ssl)+host+nowUrl.RequestURI()), appKey)
	} else {
		return fmt.Sprintf(dingLoginApi, url.QueryEscape(getProtocol(app.Ssl)+host), appKey)
	}
}
