package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	services := handleConfig()
	log.Println(services)
	for _, v := range services {
		if v.Listen != 0 {
			go startServer(&v)
		}
	}
	select {}
}

var maxTtl = 86400
var ttlMap = NewTTLMap(10, maxTtl)

type Empty struct{}

type GlobalConfig struct {
	Ding     DingConfig      `json:"ding"`
	Services []ServiceConfig `json:"service"`
}

func handleConfig() []Service {
	config := &GlobalConfig{}
	file, _ := os.Open("proxy.json")
	defer file.Close()
	bytes, err2 := ioutil.ReadAll(file)
	LogIfFatal(err2, "can not read config file named proxy.json")
	err2 = json.Unmarshal(bytes, config)
	LogIfFatal(err2, "can not read config file named proxy.json")
	services := config.Services
	if len(services) == 0 {
		LogIfFatal(errors.New("no services are set"))
	}

	returnServices := make([]Service, 0)
	for i := 0; i < len(services); i++ {
		service := &services[i]

		returnService := getService(returnServices, service.Listen)
		if returnService == nil {
			appMap := make(map[string]*App)
			putAppMap(appMap, service, config)
			returnService = &Service{
				Listen:      service.Listen,
				redirectMap: appMap,
			}
			returnServices = append(returnServices, *returnService)
		} else {
			appMap := returnService.redirectMap
			putAppMap(appMap, service, config)
		}

	}
	return returnServices
}

func putAppMap(appMap map[string]*App, service *ServiceConfig, config *GlobalConfig) {
	ding := config.Ding
	serverName := service.ServerName
	_, ok := appMap[serverName]
	if !ok {
		whiteIp := service.WhiteIp
		whiteSet := make(map[string]Empty)
		if whiteIp != "" {
			split := strings.Split(whiteIp, ",")
			for _, v := range split {
				whiteSet[v] = Empty{}
			}
			service.WhiteIp = ""
		}

		redirectServiceMap := make(map[string]*RedirectService)
		ignoreRedirects := make([]*RedirectService, 0)
		for _, v := range service.Redirects {
			if v.Ignore {
				ignoreRedirects = append(ignoreRedirects, &v)
			} else {
				redirectServiceMap[v.Path] = &v
			}
		}

		appMap[serverName] = &App{
			Listen:          service.Listen,
			ServerName:      service.ServerName,
			AuthByDing:      service.AuthByDing,
			DingAppKey:      ding.AppKey,
			DingAppSecret:   ding.AppSecret,
			WhiteIpMap:      whiteSet,
			RedirectsMap:    redirectServiceMap,
			Ssl:             service.Ssl,
			RedirectHold:    service.RedirectHold,
			IgnoreRedirects: ignoreRedirects,
			RedirectUrl:     service.RedirectUrl,
		}
	}

}

func getService(returnServices []Service, listen int) *Service {
	for _, v := range returnServices {
		if v.Listen == listen {
			return &v
		}
	}
	return nil
}
