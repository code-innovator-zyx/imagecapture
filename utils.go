package imagecapture

import (
	"crypto/rand"
	"fmt"
	"time"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2024/11/6 下午3:13
* @Package:
 */

type Element interface {
	~int | ~string | ~float64 | ~float32
}

func Min[T Element](x, y T) T {
	if x < y {
		return x
	}
	return y
}

func Max[T Element](x, y T) T {
	if x > y {
		return x
	}
	return y
}

// 计算超时时间 根据并发数，合理缩短整体超时时间
func calculateTimeout(imageNum, batchSize, routines int, baseTimeout time.Duration) time.Duration {
	times := (imageNum + batchSize - 1) / batchSize
	if imageNum > 100 {
		baseTimeout = baseTimeout * 2
	}
	timeoutFactor := float64(times) / float64(routines)
	if timeoutFactor < 1 {
		timeoutFactor = 1
	}
	timeout := time.Duration(float64(baseTimeout) * timeoutFactor)
	if timeout < baseTimeout {
		return baseTimeout
	}
	return timeout
}

// GenerateUUID 生成uuid
func GenerateUUID() (string, error) {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return "", err
	}

	// 设置UUID版本和变种
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // UUID version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // UUID variant

	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
