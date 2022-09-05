xiaoshuo-no-ads
---

## 动机
有时看一些被封的小说，只能去看盗版小说（建议支持正版，我只是在起点没有这本书的时候才去看盗版）
大家都知道盗版小说很多广告，在公司摸鱼的时候都不敢点开小说，免得弹出来奇奇怪怪的让人脸红耳赤的广告，让人社死
所以想着开发一个反向代理的web页面，自动屏蔽广告

## 运行
因为是golang写的，直接运行即可

linux
```
./xiaoshuo
```

windows
```
./xiaoshuo.exe
```

## 部署

## 交叉编译

编译为linux
```
GOOS=linux GOARCH=amd64 go build -o bin/xiaoshuo-amd64-linux
```

windows
```
GOOS=windows GOARCH=amd64 go build -o bin/xiaoshuo-amd64-windows.exe
```

mac
```
GOOS=darwin GOARCH=amd64 go build -o bin/xiaoshuo-amd64-darwin
```

## 下载的代理
```
export GOPROXY=https://proxy.golang.com.cn,direct
```

### 日志



