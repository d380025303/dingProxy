package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type ServiceConfig struct {
	Listen       int               `json:"listen"`
	ServerName   string            `json:"serverName"`
	AuthByDing   bool              `json:"authByDing"`
	Redirects    []RedirectService `json:"redirects"`
	WhiteIp      string            `json:"whiteIp"`
	Ssl          bool              `json:"ssl"`
	RedirectHold bool              `json:"redirectHold"`
	RedirectUrl  string            `json:"redirectUrl"`
}

type RedirectService struct {
	Path   string `json:"path"`
	Url    string `json:"url"`
	Ignore bool   `json:"ignore"`
}

type Service struct {
	Listen      int
	redirectMap map[string]*App
}

type App struct {
	Listen          int
	ServerName      string
	AuthByDing      bool
	DingAppKey      string
	DingAppSecret   string
	WhiteIpMap      map[string]Empty
	RedirectsMap    map[string]*RedirectService
	IgnoreRedirects []*RedirectService
	Ssl             bool
	RedirectHold    bool
	RedirectUrl     string
}

func startServer(service *Service) {
	addr := fmt.Sprintf(":%v", service.Listen)
	err := http.ListenAndServe(addr, service)
	LogIfFatal(err, "ListenAndServe error: ")
}

func ifInWhite(r *http.Request, app *App) bool {
	ip, _ := GetIP(r)
	whiteIpMap := app.WhiteIpMap
	_, haveWhite := whiteIpMap[ip]
	return haveWhite
}

func ifIgnorePath(r *http.Request, app *App) bool {
	path := r.URL.Path
	redirects := app.IgnoreRedirects
	if len(redirects) > 0 {
		for _, v := range redirects {
			if strings.HasPrefix(path, v.Path) {
				return true
			}
		}
	}
	return false
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	redirectMap := s.redirectMap
	host := r.Host
	app, ok := redirectMap[host]
	if ok {
		// 白名单
		nowUrl := r.URL
		log.Println(nowUrl.RequestURI())
		if ifInWhite(r, app) {
			toProxy(w, r, app.RedirectUrl)
		} else if ifIgnorePath(r, app) {
			toProxy(w, r, app.RedirectUrl)
		} else if app.AuthByDing {
			proxy4Ding(w, r, app)
		} else {
			toProxy(w, r, app.RedirectUrl)
		}
	} else {
		log.Println("host is not support : ", host)
	}

}

func getProtocol(ssl bool) string {
	if ssl {
		return "https://"
	} else {
		return "http://"
	}
}

func proxy4Ding(w http.ResponseWriter, r *http.Request, app *App) {
	nowUrl := r.URL
	host := app.ServerName
	appKey := app.DingAppKey
	appSecret := app.DingAppSecret
	authCode := GetUrlParam(nowUrl, "authCode")
	if authCode != "" && !ttlMap.Have(authCode) {
		// 授权登录逻辑
		if CheckAuthCode(authCode, appKey, appSecret) {
			var cookieValue = host + authCode
			ttlMap.Put(cookieValue, "")
			cookie := &http.Cookie{
				Name:   "dingA",
				Value:  cookieValue,
				Path:   "/",
				Domain: host,
				MaxAge: maxTtl,
			}
			http.SetCookie(w, cookie)
			query := nowUrl.RawQuery
			split := strings.Split(query, "&")
			var filterUrlSplit = make([]string, 0)
			for _, v := range split {
				if !strings.HasPrefix(v, "authCode=") {
					filterUrlSplit = append(filterUrlSplit, v)
				}
			}
			var p = ""
			if len(filterUrlSplit) > 0 {
				p = "?" + strings.Join(filterUrlSplit, "&")
			}
			redirect := getProtocol(app.Ssl) + host + nowUrl.Path + p
			log.Println("check auth code success: ", redirect)
			http.Redirect(w, r, redirect, http.StatusSeeOther)
		} else {
			log.Println("check auth code fail: ")
			http.Redirect(w, r, GetDingAuthUrl(app, nowUrl), http.StatusSeeOther)
		}
	} else {
		cookie, err := r.Cookie("dingA")
		if err != nil {
			// 没找到
			log.Println("named cookie not present")
			http.Redirect(w, r, GetDingAuthUrl(app, nowUrl), http.StatusSeeOther)
		} else {
			value := cookie.Value
			_, ok := ttlMap.Get(value)
			if ok {
				toProxy(w, r, app.RedirectUrl)
			} else {
				http.Redirect(w, r, GetDingAuthUrl(app, nowUrl), http.StatusSeeOther)
			}
		}
	}
}

func toProxy(w http.ResponseWriter, r *http.Request, innerUrl string) {
	var target *url.URL
	target, _ = url.Parse(innerUrl)
	targetQuery := target.RawQuery

	proxy := httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
			req.Host = target.Host
			if targetQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = targetQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
			}
		},
	}
	proxy.ServeHTTP(w, r)
}
