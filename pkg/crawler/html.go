package crawler

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// HTMLExtractor ...
func HTMLExtractor(link string, projectPath string) {
	fmt.Println("Extracting --> ", link)

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("not html file")
		}
	}()

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	// get the html body
	resp, err := http.Get(link)
	if err != nil {
		panic(err)
	}

	// Close the body once everything else is compled
	defer resp.Body.Close()

	// get the project name and path we use the path to
	f, err := os.OpenFile(projectPath+"/"+"index.html", os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	htmlData, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	f.Write(htmlData)

}

// GetHTMLName 判断路由后面是否带html后缀，返回html文件名称
func GetHTMLName(url string) string {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("htmlName err", url)
		}
	}()
	var htmlName string
	nameArry := strings.Split(url, "/")
	//fmt.Println(nameArry)
	if isEndWithHtml := strings.Index(nameArry[len(nameArry)-1], "html"); isEndWithHtml != -1 {
		if nameArry[len(nameArry)-1] != "index.html" {
			htmlName = nameArry[len(nameArry)-1]
		} else {
			htmlName = nameArry[len(nameArry)-2] + ".html"

		}
	} else {
		htmlName = nameArry[len(nameArry)-1] + ".html"

	}
	return htmlName
}

// multiLayerHTMLExtractor 多层路由下载器,下载给定域名下的路由
func multiLayerHTMLExtractor(link string, projectPath string) {
	fmt.Println("Extracting --> ", link)

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("not html file")
		}
	}()

	htmlName := GetHTMLName(link)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	// get the html body
	resp, err := http.Get(link)
	if err != nil {
		panic(err)
	}

	// Close the body once everything else is compled
	defer resp.Body.Close()

	// get the project name and path we use the path to
	f, err := os.OpenFile(projectPath+"/"+htmlName, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	htmlData, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	f.Write(htmlData)
	multiLayerArrange(projectPath, htmlName, link)

}

// stdHTMLFile 标准化html文件
func StdHTMLFile(htmlName string) error {
	input, err := ioutil.ReadFile(htmlName)
	if err != nil {
		return err
	}
	lines := strings.Split(string(input), "\n")
	for index, line := range lines {
		if is1 := strings.LastIndex(line, ">"); is1 != len(line)-1 {
			lines[index] += lines[index+1]
			lines[index+1] = ""
		}
	}

	output := strings.Join(lines, "\n")
	return ioutil.WriteFile(htmlName, []byte(output), 0777)
}

// multiLayerArrange 替换html文件内静态资源的路径
func multiLayerArrange(projectDir string, htmlName string, link string) error {
	indexfile := projectDir + "/" + htmlName
	StdHTMLFile(indexfile)

	input, err := ioutil.ReadFile(indexfile)
	if err != nil {
		return err
	}
	lines := strings.Split(string(input), "\n")

	for index, line := range lines {
		// 判断是否有内嵌url
		lines[index] = DownloadHTMLUrl(line, link, projectDir)

		b := []byte(line)
		r := bytes.NewReader(b)
		doc, err := goquery.NewDocumentFromReader(r)
		if err != nil {
			return err
		}

		// Replace JS links in HTML
		doc.Find("script[src]").Each(func(i int, s *goquery.Selection) {
			data, exists := s.Attr("src")
			if exists {
				file := filepath.Base(data)
				if strings.Contains(lines[index], `.js`) {
					//使用正则替换而不是整行替换
					s.SetAttr("src", "js/"+file)
				} else {
					s.SetAttr("src", "js/"+file+`.js`)
				}
				s.SetAttr("src", "js/"+file)
				if data, _ := s.Attr("src"); data != "" {
					//存在bug,若html书写不规范（不换行），则会出现整行代码全部替换问题
					//lines[index] = fmt.Sprintf(`<script src="%s"></script>`, data)

					if strings.Contains(lines[index], `.js`) {
						//使用正则替换而不是整行替换
						lines[index] = reSrc.ReplaceAllString(lines[index], `src=`+data)
					} else {
						lines[index] = reSrc.ReplaceAllString(lines[index], fmt.Sprintf(`src=/js/%x.js`, md5.Sum([]byte(filepath.Base(data)))))
					}
				}
			}
		})

		// Replace CSS links in HTML
		doc.Find("link[rel='stylesheet']").Each(func(i int, s *goquery.Selection) {
			// For each item found, get the hyperlink reference
			data, exists := s.Attr("href")
			if exists {
				file := filepath.Base(data)
				s.SetAttr("href", "css/"+file)
				if data, _ := s.Attr("href"); data != "" {
					//存在bug,若html书写不规范（不换行），则会出现整行代码全部替换问题
					//lines[index] = fmt.Sprintf(`<link rel="stylesheet" type="text/css" href="%s">`, data)

					//使用正则替换而不是整行替换
					lines[index] = cssSrc.ReplaceAllString(lines[index], `href=`+data)
				}
			}
		})

		// 替换路由链接
		doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
			// For each item found, get the hyperlink reference
			data, exists := s.Attr("href")
			if exists {
				//file := filepath.Base(data)
				fileName := GetHTMLName(data)
				s.SetAttr("href", fileName)
				if data, _ := s.Attr("href"); data != "" {
					//存在bug,若html书写不规范（不换行），则会出现整行代码全部替换问题
					//lines[index] = fmt.Sprintf(`<link rel="stylesheet" type="text/css" href="%s">`, data)

					//使用正则替换而不是整行替换
					lines[index] = cssSrc.ReplaceAllString(lines[index], `href=`+fileName)
				}
			}
		})

		// Replace IMG links in HTML
		// TODO: is the regex necessary here?
		doc.Find("img[src]").Each(func(i int, s *goquery.Selection) {
			data, exists := s.Attr("src")
			if exists {
				original := lines[index]
				file := filepath.Base(data)
				s.SetAttr("src", "imgs/"+file+" ")
				if data, _ := s.Attr("src"); data != "" {
					if is1 := strings.LastIndex(data, "/"); is1 == len(data)-1 {
						data = data[:len(data)-1]
					}
					lines[index] = reSrc.ReplaceAllString(original, `src=`+data+" ")
				}
			}
		})
	}
	output := strings.Join(lines, "\n")
	return ioutil.WriteFile(indexfile, []byte(output), 0777)
}

func DownloadHTMLUrl(line string, link string, projectPath string) string {
	if cssArr := cssRegexp.FindAllStringSubmatch(line, -1); len(cssArr) != 0 {
		cssImgPath := cssArr[0][0]
		ImgPath := htmlUrlImg.FindAllStringSubmatch(cssImgPath, -1)
		downLink := link + ImgPath[0][0]
		fmt.Println("---------download-link-------", downLink)
		imgName := imgRegex.FindAllStringSubmatch(line, -1)[0][0]
		line = cssRegexp.ReplaceAllString(line, fmt.Sprintf(`url(./imgs/%s)`, filepath.Base(imgName)))
		DownImg(projectPath, downLink, filepath.Base(imgName))

		return line
	}
	return line
}

var htmlUrlImg = regexp.MustCompile(`/.*(\.png|\.jpg|\.jpeg|\.gif|\.svg|\.ico)`)
var reSrc = regexp.MustCompile(`src\s*=\s*"(.+?)"|src\s*=\s*'(.+?)'`)
var cssSrc = regexp.MustCompile(`href\s*=\s*"(.+?)"|href\s*=\s*'(.+?)'`)
