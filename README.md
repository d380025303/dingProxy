# dingProxy
集成钉钉的反向代理
不需要侵入内部应用，直接外层代理请求，实现必须钉钉授权才可进入页面

# 配置:
```
{
    "ding": {
        "appKey": "", // 钉钉appKey
        "appSecret": "" // 钉钉appSecret
    },
    "service": [
        {
            "listen": 8087, // 监听端口
            "ssl": true, // 是否是https
            "serverName": "serverName", // 域名
            "redirectUrl": "http://192.168.1.115:8096", // 代理地址
            "authByDing": true, // 是否钉钉授权
            "whiteIp": "182.92.237.124" // 白名单 ,分隔
        },
        {
            "listen": 8087,
            "serverName": "serverName",
            "redirectUrl": "http://192.168.1.115:8082",
            "authByDing": true,
            "redirectHold": true,
            "redirects": [{ // 白名单url
                "path": "/openapi",
                "ignore": true
            }],
            "whiteIp": "182.92.237.124"
        }
    ]
}
```
