package html

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/imthaghost/goclone/pkg/crawler"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// TODO: figure out what was done here at 4am
func arrange(projectDir string) error {

	indexfile := projectDir + "/index.html"
	crawler.StdHTMLFile(indexfile)
	input, err := ioutil.ReadFile(indexfile)
	if err != nil {
		return err
	}
	lines := strings.Split(string(input), "\n")

	for index, line := range lines {

		lines[index] = crawler.DownloadHTMLUrl(line, crawler.DomainString, projectDir)

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
				s.SetAttr("src", "js/"+file)
				if data, _ := s.Attr("src"); data != "" {

					// 判断文件名是否带后缀
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
				fileName := crawler.GetHTMLName(data)
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

var reSrc = regexp.MustCompile(`src\s*=\s*"(.+?)"|src\s*=\s*'(.+?)'`)
var cssSrc = regexp.MustCompile(`href\s*=\s*"(.+?)"|href\s*=\s*'(.+?)'`)
