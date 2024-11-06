package imagecapture

import "io"

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2024/11/5 下午5:59
* @Package:
 */

type GoogleCapture struct{}

func NewGoogleCapture() {

}

func (GoogleCapture) SearchImages(keywords string, maxNumber int, opts ...Option) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (GoogleCapture) Download(url, filename string, writer io.Writer) error {
	//TODO implement me
	panic("implement me")
}

func (GoogleCapture) BatchDownload(urls []string, dir string, useMd5Naming bool) ([]string, error) {
	//TODO implement me
	panic("implement me")
}
