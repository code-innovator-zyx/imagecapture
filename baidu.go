package imagecapture

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// BaiduCapture 实现百度图片搜索引擎
type BaiduCapture struct {
	Downloader
	client   *http.Client
	headers  map[string]string
	baseUrl  string
	q        query
	routines int
}

// NewBaiduCapture 初始化百度图片搜索引擎 传入最大支持并发数量，建议不超过6个
func NewBaiduCapture(routineSize int) Capture {
	headers := map[string]string{
		"Accept":           "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"Proxy-Connection": "keep-alive",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
			"Chrome/84.0.4147.125 Safari/537.36",
		"Accept-Encoding": "gzip, deflate, sdch",
		"Referer":         "https://image.baidu.com/",
	}
	// 最少一个并发量
	if routineSize == 0 {
		routineSize = 6
	}
	bc := &BaiduCapture{
		client: &http.Client{
			Transport: &http.Transport{
				MaxConnsPerHost:     10,
				MaxIdleConns:        5,
				MaxIdleConnsPerHost: 5,
			},
			Timeout: 5 * time.Second,
		},
		routines: routineSize,
		headers:  headers,
		q:        newQuery(),
		baseUrl:  "https://image.baidu.com/search/flip",
	}
	bc.Downloader = newDownloader(bc.client, bc.headers)
	return bc.init()
}

func (bc *BaiduCapture) init() Capture {
	bc.q.Set("tn", "baiduimage")
	bc.q.Set("ipn", "rj")
	bc.q.Set("ct", "201326592")
	bc.q.Set("lm", "-1") // 动图   -1 正常    6- 动图
	bc.q.Set("fp", "result")
	bc.q.Set("ie", "utf-8")
	bc.q.Set("oe", "utf-8")
	bc.q.Set("st", "-1")
	bc.q.Set("pn", "0")    // 当前页
	bc.q.Set("rn", "60")   // 分页大小
	bc.q.Set("hd", "")     // 是否高清图  1. 高清
	bc.q.Set("latest", "") // 1 最新图片
	bc.q.Set("z", "")      // 尺寸大小 1-小  2-中  3-大  9-特大
	bc.q.Set("face", "")
	bc.q.Set("copyright", "") // 版权问题
	return bc
}

func (bc *BaiduCapture) SearchImages(keyword string, maxNumber int, opts ...Option) ([]string, error) {
	pool, err := ants.NewPool(bc.routines)
	if err != nil {
		return nil, err
	}
	defer pool.Release()
	q := bc.q
	q.Set("word", keyword)
	for _, option := range opts {
		option(&q)
	}
	batchSize := 60
	var collector = make(chan string, Min(maxNumber, batchSize))
	// 设置单个任务的基础超时（例如 3 秒）
	baseTimeout := 3 * time.Second
	timeout := calculateTimeout(maxNumber, batchSize, bc.routines, baseTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	var wg sync.WaitGroup
	defer cancel()
	for i := 0; i < maxNumber; i += batchSize {
		q.Set("pn", strconv.Itoa(i))
		queryURL := fmt.Sprintf("%s?%s", bc.baseUrl, q.Encode())
		wg.Add(1)
		err = pool.Submit(func() {
			bc.searchBaidu(ctx, queryURL, collector)
			wg.Done()
		})
		if err != nil {
			return nil, err
		}
	}
	var imageUrls = make(map[string]struct{}, maxNumber)
	go func() {
		wg.Wait()
		close(collector)
	}()
SELECT:
	for {
	Next:
		select {
		case url, ok := <-collector:
			if !ok {
				// 所有goroutine 执行完了，但是数量不够，任然要返回的
				break SELECT
			}
			for _, source := range unSupportSource {
				if strings.Contains(strings.ToLower(url), source) {
					fmt.Println("包含抖音的url", url)
					break Next
				}
			}
			imageUrls[url] = struct{}{}
			if len(imageUrls) >= maxNumber {
				break SELECT
			}
		case <-ctx.Done():
			// 超时了,但是爬到的数据还是要给你的
			break SELECT
		}
	}
	var urls = make([]string, 0, len(imageUrls))
	for url := range imageUrls {
		urls = append(urls, url)
	}
	return urls, nil
}

// 获取图片
func (bc *BaiduCapture) searchBaidu(ctx context.Context, url string, collector chan<- string) {
	select {
	case <-ctx.Done():
		return
	default:
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return
		}
		for k, v := range bc.headers {
			req.Header.Set(k, v)
		}
		try := 0
		var resp *http.Response
		for {
			if try >= 3 {
				return
			}
			resp, err = bc.client.Do(req)
			if err == nil {
				break
			}
			try += 1
		}
		// Handling GZIP compressed responses
		var reader = resp.Body
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gzipReader, err := gzip.NewReader(resp.Body)
			if err != nil {
				return
			}
			reader = gzipReader
			err = gzipReader.Close()
			if err != nil {
				return
			}
		}
		var data bytes.Buffer
		_, err = io.Copy(&data, reader)
		if err != nil {
			return
		}
		err = reader.Close()
		if err != nil {
			fmt.Println("failed to close reader:", err)
		}
		err = resp.Body.Close()
		if err != nil {
			fmt.Println("failed to close body")
		}
		pattern := `"objURL":"(.*?)",`
		re := regexp.MustCompile(pattern)
		for _, data := range re.FindAllStringSubmatch(data.String(), -1) {
			if len(data) > 1 {
				collector <- data[1]
			}
		}
		return
	}
}
