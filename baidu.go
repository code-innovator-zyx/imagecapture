package imagecapture

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
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

func (bc *BaiduCapture) RangeImages(keyword string, callBack func([]string) bool, opts ...Option) error {
	q := bc.q
	q.Set("word", keyword)
	for _, option := range opts {
		option(&q)
	}
	batchSize := 60
	timeout := 3 * time.Second
	total, err := bc.queryTotalNums(q)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	collector := make(chan string, batchSize)
	defer close(collector)
	for i := 0; i < total; i += batchSize {
		q.Set("pn", strconv.Itoa(i))
		queryURL := fmt.Sprintf("%s?%s", bc.baseUrl, q.Encode())
		// 在 Goroutine 中执行爬取任务
		go func(ctx context.Context, url string) {
			bc.searchBaidu(ctx, url, collector)
		}(ctx, queryURL)
		// 收集当前分页的图片 URL
		urls := make([]string, 0, batchSize)
	WAIT:
		for {
			select {
			case <-ctx.Done():
				// 如果 context 被取消，退出当前分页的处理
				break WAIT
			case url, ok := <-collector:
				if !ok {
					// 通道已关闭，退出
					break WAIT
				}
				urls = append(urls, url)
			}
		}

		// 如果回调函数返回 false，则取消所有任务并退出
		if !callBack(urls) {
			cancel() // 通知所有 Goroutine 停止
			return nil
		}
	}
	return nil
}

// 查询接口能获取的总数量
func (bc *BaiduCapture) queryTotalNums(q query) (total int, err error) {
	q.Set("tn", "resultjson_com")
	queryURL := fmt.Sprintf("https://image.baidu.com/search/acjson?%s", q.Encode())
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return
	}
	for k, v := range bc.headers {
		req.Header.Set(k, v)
	}
	resp, err := bc.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	// 百度的响应数据是经过压缩的
	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return 0, err
	}
	defer reader.Close()

	var data bytes.Buffer
	_, err = io.Copy(&data, reader)
	if err != nil {
		return
	}
	var jsonData = make(map[string]interface{})
	err = json.Unmarshal(bytes.ReplaceAll(data.Bytes(), []byte(`'`), []byte(`"`)), &jsonData)
	if err != nil {
		return 0, err
	}
	if num, ok := jsonData["listNum"]; ok {
		if numFloat, ok := num.(float64); ok {
			total = int(numFloat) // 将 float64 转为 int
		} else {
			return 0, fmt.Errorf("listNum is not a number")
		}
	}
	return
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
	baseTimeout := 5 * time.Second
	timeout := calculateTimeout(maxNumber, batchSize, bc.routines, baseTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	increment := 0
	if maxNumber > batchSize/2 {
		increment = batchSize
	}
	var wg sync.WaitGroup

	for i := 0; i < maxNumber+increment; i += batchSize {
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
			for _, rules := range defaultFilterRules {
				if rules.Check(strings.ToLower(url)) {
					break Next
				}
			}
			imageUrls[url] = struct{}{}
			if len(imageUrls) >= maxNumber {
				// 主动退出
				//cancel()
				break SELECT
			}
		case <-ctx.Done():
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
	var reader io.ReadCloser
	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err = gzip.NewReader(resp.Body)
	} else {
		reader = resp.Body
	}
	defer reader.Close()
	var data bytes.Buffer
	_, err = io.Copy(&data, reader)
	if err != nil {
		return
	}
	pattern := `"objURL":"(.*?)",`
	re := regexp.MustCompile(pattern)
	for _, data := range re.FindAllStringSubmatch(data.String(), -1) {
		select {
		case <-ctx.Done():
			fmt.Println("cancel")
			return
		default:
			if len(data) > 1 {
				collector <- data[1]
			}
		}
	}
	return

}
