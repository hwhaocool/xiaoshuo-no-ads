xiaoshuo-no-ads
---

## 动机
有时看一些被封的小说，只能去看盗版小说（建议支持正版，我只是在起点没有这本书的时候才去看盗版）
大家都知道盗版小说很多广告，在公司摸鱼的时候都不敢点开小说，免得弹出来奇奇怪怪的让人脸红耳赤的广告，让人社死
所以想着开发一个反向代理的web页面，自动屏蔽广告

[demo地址](http://xiaoshuo.guojiang.ltd)

## 运行
因为是golang写的，直接运行即可

linux
```
./xiaoshuo-amd64-linux
```

windows
```
./xiaoshuo-amd64-windows.exe
```

然后访问 `localhost:8090` 就可以看到效果了


## 使用
运行起来之后，两种用法
1. 复制小说网站地址，粘贴进来，点击转换，再点击新的地址即可
2. 打开小说网站，在地址栏最前面添加 你的域名 即可

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

