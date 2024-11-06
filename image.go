package imagecapture

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"hash"
	"io"
	"net/http"
	"strings"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2024/11/5 上午11:35
* @Package:
 */
const sniffLen = 512
const base = "image"

func checkImageType(data []byte) string {
	// 获取图片类型
	ty := http.DetectContentType(data[:sniffLen])
	arrs := strings.Split(ty, "/")
	if len(arrs) < 2 || arrs[0] != base {
		return ""
	}
	return strings.ToLower(arrs[1])
}

type ImageReader struct {
	io.Reader
	ty  string
	md5 hash.Hash
}

func NewImageReader(reader io.Reader, needMd5 bool) (*ImageReader, error) {
	buf := make([]byte, sniffLen)
	n, err := reader.Read(buf)
	if err != nil {
		return nil, err
	}
	// 检测图片类型
	ty := checkImageType(buf[:n])
	r := io.MultiReader(bytes.NewReader(buf[:n]), reader)
	ir := &ImageReader{
		Reader: r,
		ty:     ty,
	}
	if needMd5 {
		m5 := md5.New()
		m5.Write(buf[:n])
		// 创建一个多读取器，同时往m5写入
		ir.Reader = io.MultiReader(bytes.NewReader(buf[:n]), io.TeeReader(reader, m5))
		ir.md5 = m5
	}
	return ir, nil
}

func (ir ImageReader) Type() string {
	return ir.ty
}

func (ir ImageReader) Md5() string {
	hashBytes := ir.md5.Sum(nil)
	// 计算最终的 MD5 值
	return hex.EncodeToString(hashBytes) // 转换为字符
}

func (ir ImageReader) Read(p []byte) (n int, err error) {
	return ir.Reader.Read(p)
}
