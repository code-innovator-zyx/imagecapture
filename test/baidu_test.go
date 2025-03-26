package test

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/code-innovator-zyx/imagecapture"
	"io"
	"os"
	"testing"
	"time"
)

/*
* @Author: zouyx
* @Email:	1003941268@qq.com
* @Date:   2024/11/1 下午5:30
* @Package:
 */

func Test_BaiduCapture(t *testing.T) {
	capture := imagecapture.NewBaiduCapture(3)
	t.Run("SearchImages", func(t *testing.T) {
		urls, err := capture.SearchImages("石英", 60, imagecapture.WithImageSize(imagecapture.ImageSize_LARGE))
		if err != nil {
			t.Error(err.Error())
			return
		}
		fmt.Println(len(urls))
	})

	t.Run("RangeImages", func(t *testing.T) {
		var images []string
		err := capture.RangeImages("老虎", func(urls []string) bool {
			images = append(images, urls...)
			if len(images) >= 60 {
				return false
			}
			return true
		})
		if err != nil {
			t.Error(err.Error())
			return
		}
		t.Log(len(images))
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

func Test_TeeReader(t *testing.T) {
	file, err := os.Open("./beautiful/8e57310b3e0fab04c4ce111bb94a9e8f.gif")
	if err != nil {
		panic(err)
	}
	h := md5.New()
	tr := io.TeeReader(file, h)
	_, err = io.ReadAll(tr)
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(h.Sum(nil)[:]))
}
