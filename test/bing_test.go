package test

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/code-innovator-zyx/imagecapture"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2024/11/5 下午5:36
* @Package:
 */
func Test_binghtml(t *testing.T) {
	// 打开 HTML 文件
	file, err := os.Open("./bing.html")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 定义正则表达式来匹配 objURL
	// 定义正则表达式来匹配 imgData 的完整行
	//rule := regexp.MustCompile(`"murl":"http[^\"]+`)
	doc, err := html.Parse(file)
	if err != nil {
		t.Fatal(err)
		return
	}
	var links []imagecapture.BingFormat
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			// 检查是否有指定的 class 属性
			for _, attr := range n.Attr {
				if attr.Key == "class" && attr.Val == "iusc" {
					// 找到 href 属性
					for _, attr := range n.Attr {
						if attr.Key == "m" {
							var data imagecapture.BingFormat
							json.Unmarshal([]byte(attr.Val), &data)
							fmt.Println(data)
							links = append(links, data)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	fmt.Println(len(links))
	fmt.Println(links[0])
}

func Test_Bing(t *testing.T) {
	capture := imagecapture.NewBingCapture(3)
	t.Run("SearchImages", func(t *testing.T) {
		start := time.Now()
		urls, err := capture.SearchImages("石英", 20)
		if err != nil {
			t.Error(err.Error())
			return
		}
		fmt.Println("search cost", time.Since(start).Milliseconds())
		fmt.Println(urls)
		t.Log(len(urls))
	})
	t.Run("RangeImages", func(t *testing.T) {
		var nums int
		err := capture.RangeImages("老虎", func(urls []string) bool {
			nums += len(urls)
			fmt.Println("current get ", len(urls))
			fmt.Println("total ", nums)
			for i := range urls {
				fmt.Println(urls[i])
			}
			if nums >= 120 {
				return false
			}
			return true
		})
		if err != nil {
			t.Error(err.Error())
			return
		}
	})

	t.Run("Download filename", func(t *testing.T) {
		start := time.Now()
		urls, err := capture.SearchImages("美女", 5)
		if err != nil {
			t.Error(err.Error())
			return
		}
		fmt.Println(time.Since(start).Milliseconds())
		capture.Download(urls[0], "./beautiful", nil)
	})

	t.Run("Download writer", func(t *testing.T) {
		urls, err := capture.SearchImages("美女", 5)
		if err != nil {
			t.Error(err.Error())
			return
		}
		file, err := os.OpenFile("./test", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)

		suffix, err := capture.Download(urls[0], "", file)
		if err != nil {
			t.Error(err.Error())
			return
		}
		fmt.Println(suffix)
		file.Close()
		os.Rename("./test", "./test."+suffix)
		//err = os.WriteFile(fmt.Sprintf("./test.%s", suffix), buf.Bytes(), 666)
		//if err != nil {
		//	t.Log(err.Error())
		//}
	})

	t.Run("BatchDownload", func(t *testing.T) {
		urls, err := capture.SearchImages("美女", 5)
		if err != nil {
			t.Error(err.Error())
			return
		}
		fmt.Println(len(urls))
		paths, _ := capture.BatchDownload(urls, "./beautiful", true)
		for i := range paths {
			fmt.Println(paths[i])
		}
	})
}

func Test_md5(t *testing.T) {
	res, err := http.Get("https://pic.52112.com/180419/180419_178/3346oYPl0F_small.jpg")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(res.StatusCode)
	h := md5.New()
	tr := io.TeeReader(res.Body, h)
	_, err = io.ReadAll(tr)
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(h.Sum(nil)[:]))
}
