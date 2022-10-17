package crawler

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/gocolly/colly/v2"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strings"
	"sync"
)

var recordMap = make(map[string]bool)
var once sync.Once
var hzRegexp, _ = regexp.Compile("([\u4e00-\u9fa5]+)")
var cssRegexp = regexp.MustCompile(`url\(.*(\.png|\.jpg|\.jpeg|\.gif|\.svg|\.ico)\)|url\('.*(\.png|\.jpg|\.jpeg|\.gif|\.svg|\.ico).\)`)
var DomainString string
var imgRegexp, _ = regexp.Compile(`.*(\.png|\.jpg|\.jpeg|\.gif|\.svg|\.ico)`)

// Collector 从给到的链接中搜索img, js, html以及css文件
// TODO improve for better performance
func Collector(ctx context.Context, url string, projectPath string, cookieJar *cookiejar.Jar, proxyString string, userAgent string, recursion int) error {

	// 使用单例模式将根网址保存至变量
	once.Do(func() {
		DomainString = url
	})
	// 出递归
	if recursion < 1 {
		return nil
	} else {
		// create a new collector
		c := colly.NewCollector(colly.Async(true))
		setUpCollector(c, ctx, cookieJar, proxyString, userAgent)

		// 下载css文件
		c.OnHTML("link[rel='stylesheet']", func(e *colly.HTMLElement) {
			fileName := ""
			// hyperlink reference
			link := e.Attr("href")

			// 判断是否需要下载，若以已经在map中则取消下载
			if !isInRecordMap(link) {
				// print css file was found
				fmt.Println("Css found", "-->", link)
				// extraction
				Extractor(e.Request.AbsoluteURL(link), projectPath, fileName)
			}

		})

		// 下载js文件
		c.OnHTML("script[src]", func(e *colly.HTMLElement) {
			fileName := ""
			// src attribute
			link := e.Attr("src")

			// 判断是否需要下载
			if !isInRecordMap(link) {
				// Print link
				fmt.Println("Js found", "-->", link)

				// 若为不带js后缀的js文件
				if !strings.Contains(link, `.js`) {
					fileName = fmt.Sprintf(`%x.js`, md5.Sum([]byte(link[strings.LastIndex(link, `/`)+1:])))
					link += `.js`
				}
				// extraction
				Extractor(e.Request.AbsoluteURL(link), projectPath, fileName)
			}

		})

		//下载所有静态资源
		c.OnHTML("link[href]", func(e *colly.HTMLElement) {
			fileName := ""
			// 匹配href属性
			link := e.Attr("href")

			if !isInRecordMap(link) {
				// 输出下载链接
				fmt.Println("all resource found", "-->", link)
				// 下载静态资源至指定目录
				Extractor(e.Request.AbsoluteURL(link), projectPath, fileName)
			}

		})

		// 下载图片
		c.OnHTML("img[src]", func(e *colly.HTMLElement) {
			fileName := ""
			// src at tribute
			link := e.Attr("src")
			if strings.HasPrefix(link, "data:image") || strings.HasPrefix(link, "blob:") {
				return
			}

			if !isInRecordMap(link) {
				// Print link
				fmt.Println("Img found", "-->", link)
				// extraction

				// 判断是否有中文路径
				if hzRegexp.MatchString(link) {
					// 若存在中文路径则直接传输过去
					Extractor(link, projectPath, fileName)

				} else {
					Extractor(e.Request.AbsoluteURL(link), projectPath, fileName)
				}
			}

		})

		//下载下一层网页资源
		c.OnHTML("a[href]", func(e *colly.HTMLElement) {

			// 匹配href属性
			link := e.Attr("href")
			// 匹配给定域名下的路由
			reg := regexp.MustCompile(DomainString)
			result := reg.FindAllStringSubmatch(link, -1)

			// 匹配HTML文件的路径是否为相对路径
			reg1 := regexp.MustCompile(`^\./`)
			result1 := reg1.FindAllStringSubmatch(link, -1)

			// 判断是否为给定域名下的路由
			if len(result) != 0 {
				fmt.Println("http resource found", "-->", link)
				// 递归下载网页资源
				Collector(ctx, link, projectPath, cookieJar, proxyString, userAgent, recursion-1)

			}

			// 判断静态资源是否为相对路径
			if len(result1) != 0 {
				fmt.Println("http resource found", "-->", link)
				//拼接url
				if is1 := strings.LastIndex(url, "/"); is1 == len(url)-1 {
					link = url + link[2:]
				} else {
					link = url + link[1:]
				}
				// 递归下载网页资源
				Collector(ctx, link, projectPath, cookieJar, proxyString, userAgent, recursion-1)
			}
			//Extractor(e.Request.AbsoluteURL(link), projectPath)

		})

		//Before making a request
		c.OnRequest(func(r *colly.Request) {
			link := r.URL.String()

			if !isInRecordMap(link) {
				if url == DomainString {
					HTMLExtractor(link, projectPath)
				} else {
					multiLayerHTMLExtractor(link, projectPath)
				}
			}
		})

		// Visit each url and wait for stuff to load :)
		if err := c.Visit(url); err != nil {
			return err
		}
		c.Wait()

		return nil
	}

}

type cancelableTransport struct {
	ctx       context.Context
	transport http.RoundTripper
}

func (t cancelableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.ctx.Err(); err != nil {
		return nil, err
	}
	return t.transport.RoundTrip(req.WithContext(t.ctx))
}

func setUpCollector(c *colly.Collector, ctx context.Context, cookieJar *cookiejar.Jar, proxyString, userAgent string) {
	if cookieJar != nil {
		c.SetCookieJar(cookieJar)
	}
	if proxyString != "" {
		c.SetProxy(proxyString)
	} else {
		c.WithTransport(cancelableTransport{ctx: ctx, transport: http.DefaultTransport})
	}
	if userAgent != "" {
		c.UserAgent = userAgent
	}
}

// isInRecordMap 判断资源是否在map中
func isInRecordMap(name string) bool {
	if _, ok := recordMap[name]; ok {
		return true
	} else {
		recordMap[name] = true
		return false
	}
}
