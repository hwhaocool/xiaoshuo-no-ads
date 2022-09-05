package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"xiaoshuo/myLog"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/brotli"
	"github.com/gookit/color"
	"github.com/imroc/req/v3"
)

var logger *zap.Logger
var client *req.Client

func main() {

	// 通过 -p 得到端口号
	var port int
	flag.IntVar(&port, "p", 8090, "端口号")
	flag.Parse()

	color.Yellow.Sprintf("port is  :%d", port)

	// 1. 初始化日志
	logger = myLog.Init()
	logger.Info(color.FgMagenta.Render("xiaoshuo begin to start"))

	// 2. 初始化 client
	// req.DevMode()

	client = req.C().EnableTraceAll()
	// EnableDebugLog().EnableDumpAllWithoutResponseBody()
	// client = req.C()

	client.OnAfterResponse(func(client *req.Client, resp *req.Response) error {
		if resp.Err != nil { // you can skip if error occurs.
			return nil
		}

		// You can access Client and current Response object to do something
		// as you need

		trace := resp.TraceInfo()
		color.Yellow.Printf("%+v", trace)
		fmt.Println()

		return nil // return nil if it is success
	})

	//初始化连接池
	// httpMgr.Init()

	//打印ip
	logger.Info("local ip", zap.String("ip", getLocalIP()))

	//新建gin 实例
	router := gin.New()

	// 开启gzip 压缩
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// 图标
	router.StaticFile("/favicon.ico", "./html/favicon.png")

	// 健康检查
	router.HEAD("/", nginxHealthCheck)

	router.LoadHTMLGlob("html/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index1.html", "nil")
	})

	router.GET("/read", func(c *gin.Context) {
		c.HTML(200, "read.html", "nil")
	})

	//每个请求都处理一下
	router.Use(ginBodyLogMiddleware())

	// g1 := router.Group("/")

	// g1.Use(ginBodyLogMiddleware())

	router.GET("/31xiaoshuo/*id", xiaoshuo31WithReq) // 31小说

	router.GET("/biquge/*id", biqugeWithReq) // 笔趣阁

	//其它 ->

	router.NoRoute(myNoRoute2)

	//启动 gin 并监听端口
	err := router.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Fatal("proxy start failed,", zap.Error(err))
	}
}

func biqugeWithReq(ctx *gin.Context) {

	// 1. 发出请求，得到 response
	id := ctx.Param("id")

	// resp, err := genRequest(ctx).
	resp, err := client.R().
		// SetHeader("Accept-Encoding", "gzip").
		Get("https://www.ibiquge.net" + id)

	if err != nil {
		logger.Error("ibiquge req error", zap.Error(err))
		ctx.Header("Content-Type", "text/html; charset=utf-8")
		ctx.Writer.WriteString("network error, please try again")
		return
	}

	// fmt.Fprintf(os.Stdout, "%s", resp.String())

	// var respStr string

	// 2. 如果response 是 br的，那么就解压缩

	// 3. 处理response
	ret := processBiquge(resp.Body)

	// 4. 打印Header
	green := color.FgGreen.Render

	logger.Debug(green("biqugeWithReq resp"),
		zap.Any("all_headers", resp.Header),
	)

	// 5. 返回网页内容
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Writer.WriteString(ret)
}

// 自行发送请求
func xiaoshuo31WithReq(ctx *gin.Context) {

	// 请求分为几种
	// 1. html 代码请求
	// 2. js/css  请求 or 屏蔽 or 重定向

	// 其中html 根据情况进行内容处理

	// 1. 发出请求，得到 response
	id := ctx.Param("id")

	resp, err := genRequest(ctx).
		SetHeader("Accept-Encoding", "gzip, deflate, br").
		Get("https://m.31xiaoshuo.com" + id)

	if err != nil {
		logger.Error("error", zap.String("a", "a"))
	}

	var respStr string

	// 2. 如果response 是 br的，那么就解压缩
	if resp.Header.Get("Content-Encoding") == "br" {
		// "Content-Encoding
		// logger.Debug(color.FgYellow.Render("body content is br"))

		respStr = decodeBr(resp.Bytes())
	} else {
		respStr = resp.String()
	}

	// 3. 处理response
	ret := process31Html(respStr)

	// 4. 打印Header
	// green := color.FgGreen.Render

	// logger.Debug(green("xiaoshuo31WithReq resp"),
	// 	zap.Any("all_headers", resp.Header),
	// )

	// 5. 返回网页内容
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Writer.WriteString(ret)

}

// 处理 31小说
func process31Html(respStr string) string {

	return processCommonHtml(strings.NewReader(respStr), "/31xiaoshuo", func(doc *goquery.Document) {

		// 删除 31 自己的 app 广告
		doc.Find("#ljPz").Each(func(i int, s *goquery.Selection) {
			// logger.Debug(color.FgMagenta.Render("find ljPz"))

			s.Remove()
		})

		doc.Find(".RMss1").Remove()

		// 其它小说网站 链接
		doc.Find("#wzsy").Remove()

		// 推荐阅读
		doc.Find(".hot").Remove()

		// 书签
		doc.Find(".am-gotop").Remove()

		// 书架
		doc.Find(".am-header-right").Remove()
		doc.Find(".am-header-left").Remove()

		// am-header-title
		doc.Find(".am-header-title").Each(func(i int, s *goquery.Selection) {

			text := s.Text()
			s.SetText(text + " by:土拨鼠看小说")
		})

	})
}

// 处理 biquge
func processBiquge(respReader io.ReadCloser) string {

	return processCommonHtml(respReader, "/biquge", func(doc *goquery.Document) {

		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			text := s.Text()
			if strings.Contains(text, "章节错误") || strings.Contains(text, "加入书签") {
				s.Remove()
			}
		})

		doc.Find("#content").Each(func(i int, s *goquery.Selection) {
			s.SetAttr("style", "font-size: 1.3rem")
		})

		doc.Find("p").Each(func(i int, s *goquery.Selection) {
			text := s.Text()
			if strings.Contains(text, "天才一秒记住本站地址") {
				s.Remove()
			}
		})

		doc.Find(".ywtop").Remove()
		doc.Find(".header").Remove()
		doc.Find(".clear").Remove()
		doc.Find(".nav").Remove()
		doc.Find(".lm").Remove()
		doc.Find("div[align=center]").Remove()
		doc.Find("#page_set").Remove()
	})
}

// 通用的，比如出去广告什么的
func processCommonHtml(respReader io.Reader, hrefPrefix string, callback func(doc *goquery.Document)) string {

	// Pass resp.Body to goquery.
	// ioReaderData := strings.NewReader(respStr)
	// doc, err := goquery.NewDocumentFromReader(strings.NewReader(respStr))
	doc, err := goquery.NewDocumentFromReader(respReader)
	if err != nil {
		logger.Error("processCommonHtml", zap.Error(err))
		return ""
	}

	// 1. 去除广告
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")

		text := s.Text()

		if strings.Contains(text, "baidu.com") || strings.Contains(text, "bdstatic.com") {
			// 国内广告
			s.Remove()

		} else if strings.HasSuffix(src, "adsbygoogle.js") {
			// 谷歌的广告
			s.Remove()
		} else if strings.HasSuffix(src, ".js") {
			// 干掉所有js
			s.Remove()
		}

	})

	// 2. 删除所有图片
	doc.Find("img").Remove()

	// 3. 修改 href
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exist := s.Attr("href")
		if exist {
			newhref := hrefPrefix + href

			s.SetAttr("href", newhref)
		}
	})

	// 4. 删除layui
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if strings.HasSuffix(href, "layui.css") {

			// s.SetAttr("href", "//unpkg.com/layui@2.6.8/dist/css/layui.css")
			s.Remove()
		}
	})

	// 5. 修改标题
	doc.Find("title").Each(func(i int, s *goquery.Selection) {
		text := s.Text()

		s.SetText("土拨鼠看小说 " + text)
	})

	// 6. 其他乱七八糟的
	doc.Find("footer").Remove()
	doc.Find(".footer").Remove()
	doc.Find("form").Remove()

	// 7. 回调
	callback(doc)

	// 8. 返回
	html, err := doc.Html()

	if err != nil {
		logger.Error("doc.Html() error", zap.Error(err))
		return ""
	}

	return html
}

func decodeBr(bs []byte) string {
	br := brotli.NewReader(bytes.NewReader(bs))
	// defer br.Close()
	respBody, err := ioutil.ReadAll(br)
	if err != nil {
		logger.Info("decodeBr failed", zap.Error(err))
		return ""
	}

	return string(respBody)
}

// 生成请求
func genRequest(ctx *gin.Context) *req.Request {
	return client.R(). // Use R() to create a request.
				SetHeader("Accept", ctx.Request.Header.Get("Accept")).
				SetHeader("User-Agent", ctx.Request.Header.Get("User-Agent")).
				EnableDump()
}

func myNoRoute2(ctx *gin.Context) {

	uri := ctx.Request.RequestURI

	anyweb(ctx, uri)
}

func anyweb(ctx *gin.Context, uri string) {
	logger.Debug(color.FgMagenta.Render("anyweb"),
		zap.String("path", ctx.Request.RequestURI), zap.String("target", uri))

	// 1. 发出请求，得到 response

	resp, err := genRequest(ctx).
		// SetHeader("Accept-Encoding", "gzip, deflate, br").
		Get(uri[1:])

	if err != nil {
		logger.Error("request error", zap.Error(err))

		ctx.Status(200)
		ctx.Header("Content-Type", "text/html; charset=utf-8")
		ctx.Writer.WriteString("网络异常，请刷新重试")
		return
	}

	// 2. 处理 href
	newuri, err := url.Parse(uri)
	if err != nil {
		ctx.Status(200)
		ctx.Header("Content-Type", "text/html; charset=utf-8")
		ctx.Writer.WriteString("网址格式错误: " + uri)

		return
	}

	s := newuri.Scheme
	h := newuri.Host
	hrefPrefix := s + "://" + h
	// logger.Debug(color.FgMagenta.Render("hrefPrefix"), zap.String("hrefPrefix", hrefPrefix))

	var respReader io.Reader

	// 2. 如果response 是 br的，那么就解压缩
	if resp.Header.Get("Content-Encoding") == "br" {
		// "Content-Encoding
		logger.Debug(color.FgYellow.Render("body content is br"))

		respReader = strings.NewReader(decodeBr(resp.Bytes()))
	} else {
		respReader = resp.Body
	}

	// 3. 处理response
	ret := processCommonHtml(respReader, hrefPrefix, func(doc *goquery.Document) {})

	// 4. 打印Header
	// green := color.FgGreen.Render

	// logger.Debug(green("anyweb resp"),
	// 	zap.Any("all_headers", resp.Header),
	// )

	// 5. 返回网页内容
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Writer.WriteString(ret)

}

// slb 健康检查接口使用 head 方法
func nginxHealthCheck(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"type":    "ok",
		"message": "proxy is ok",
	})

	ctx.Abort()
}

// getLocalIP 得到local ip
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func ginBodyLogMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		logger.Debug(color.FgMagenta.Render("ginBodyLogMiddleware"),
			zap.String("path", ctx.Request.RequestURI))

		// 1. 如果是js 全部拦下来
		uri := ctx.Request.RequestURI
		if strings.HasSuffix(uri, ".js") || strings.HasSuffix(uri, ".gif") {
			ctx.Status(404)
			ctx.Writer.WriteString("")

			ctx.Abort()

			return
		}

		// 2. 如果是 css 重定向
		if strings.HasSuffix(uri, ".css") {

			cssUrl := buildCssUrl(uri)

			if cssUrl == "" {
				ctx.Status(404)
				ctx.Writer.WriteString("")
				ctx.Abort()

				return
			}

			ctx.Redirect(http.StatusMovedPermanently, cssUrl)

			ctx.Abort()
			return
		}

		start := time.Now().UnixMilli()

		// 请求
		ctx.Next()

		// 请求后

		end := time.Now().UnixMilli()
		cost := end - start
		host := ctx.Request.Host

		logger.Info(color.FgGreen.Render("process, ")+color.FgYellow.Render(fmt.Sprint(cost))+"ms "+ctx.Request.Method,
			zap.String("url", ctx.Request.RequestURI),
			zap.String("host", host),
			zap.String("ip", ctx.Request.Header.Get("X-Forwarded-For")),
			zap.String("UA", ctx.Request.Header.Get("User-Agent")),
			zap.Int("res-code", ctx.Writer.Status()),
		)

	}
}

func buildCssUrl(uri string) string {
	if strings.HasPrefix(uri, "/31xiaoshuo") {
		return "https://m.31xiaoshuo.com" + strings.Replace(uri, "/31xiaoshuo", "", 1)
	} else if strings.HasPrefix(uri, "/biquge") {
		return "https://m.ibiquge.net" + strings.Replace(uri, "/biquge", "", 1)
	} else {
		return ""
	}
}
