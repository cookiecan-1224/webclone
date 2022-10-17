package crawler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/imthaghost/goclone/pkg/parser"
)

// file extension map for directing files to their proper directory in O(1) time
var (
	extensionDir = map[string]string{
		".css":  "css",
		".js":   "js",
		".jpg":  "imgs",
		".jpeg": "imgs",
		".gif":  "imgs",
		".png":  "imgs",
		".svg":  "imgs",
		".ico":  "imgs",
	}
)

// Extractor visits a link determines if its a page or sublink
// downloads the contents to a correct directory in project folder
// TODO add functionality for determining if page or sublink
func Extractor(link string, projectPath string, stdFileName string) {
	fmt.Println("Extracting --> ", link)
	isStatic, filename := changFileName(link)
	if isStatic {
		link = filename
	}
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			return
		}
	}()

	// get the html body
	resp, err := http.Get(link)
	if err != nil {
		panic(err)
	}

	// Closure
	defer resp.Body.Close()
	// file base
	base := parser.URLFilename(link)
	// store the old ext, in special cases the ext is weird ".css?a134fv"
	oldExt := filepath.Ext(base)
	// new file extension
	ext := parser.URLExtension(link)

	// checks if there was a valid extension
	if ext != "" {
		// checks if that extension has a directory path name associated with it
		// from the extensionDir map
		dirPath := extensionDir[ext]
		if dirPath != "" {
			// If extension and path are valid pass to writeFileToPath
			writeFileToPath(projectPath, base, oldExt, ext, dirPath, resp, link, stdFileName)
		}
	}
}

func writeFileToPath(projectPath, base, oldFileExt, newFileExt, fileDir string, resp *http.Response, link string, stdFileName string) {
	var name = base[0 : len(base)-len(oldFileExt)]
	var filename string
	document := name + newFileExt

	if stdFileName == "" {
		filename = projectPath + "/" + fileDir + "/" + document
	} else {
		filename = projectPath + "/" + fileDir + "/" + stdFileName
	}

	// get the project name and path we use the path to
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	htmlData, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}
	f.Write(htmlData)

	// 若为css文件
	if strings.Contains(document, `.css`) {
		replaceCSSFile(filename, link[:strings.Index(link, document)], projectPath)
	}
}

// changFileName 将不合法命名的静态资源文件改名
func changFileName(fileName string) (bool, string) {
	reg := regexp.MustCompile(`.*.\.js|.*.\.css`)
	if reg == nil {
		return false, "regexp err"
	}
	result := reg.FindAllStringSubmatch(fileName, -1)
	if len(result) != 0 {
		return true, result[0][0]
	}
	return false, ""
}

// replaceCSSFile 替换css文件中的图片路径
func replaceCSSFile(cssName string, link string, projectPath string) error {
	input, err := ioutil.ReadFile(cssName)
	if err != nil {
		return err
	}
	lines := strings.Split(string(input), "\n")
	for index, line := range lines {
		if cssArr := cssRegexp.FindAllStringSubmatch(line, -1); len(cssArr) != 0 {
			cssImgPath := cssArr[0][0]
			cssImgPath = cssImgPath[4 : len(cssImgPath)-1]
			downLink := link + cssImgPath
			imgName := imgRegex.FindAllStringSubmatch(lines[index], -1)[0][0]
			lines[index] = cssRegexp.ReplaceAllString(lines[index], fmt.Sprintf(`url(%s/imgs/%s)`, projectPath, filepath.Base(imgName)))
			DownImg(projectPath, downLink, filepath.Base(imgName))
		}
	}
	output := strings.Join(lines, "\n")
	return ioutil.WriteFile(cssName, []byte(output), 0777)
}

var imgRegex = regexp.MustCompile(`/.*(\.png|\.jpg|\.jpeg|\.gif|\.svg|\.ico)`)

// downImg 下载css文件中的图片链接
func DownImg(projectPath string, link string, imgName string) {

	resp, err := http.Get(link)
	if err != nil {
		panic(err)
	}

	filename := projectPath + "/imgs" + "/" + imgName
	// get the project name and path we use the path to
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0777)
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
