package imagecapture

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"golang.org/x/net/html"
	"net/http"
	"strconv"
	"sync"
	"time"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2024/11/2 下午10:54
* @Package:
 */

type BingFormat struct {
	Murl string `json:"murl"`
	Turl string `json:"turl"`
}

type BingCapture struct {
	client   *http.Client
	headers  map[string]string
	baseUrl  string
	q        query
	routines int
	Downloader
}

func NewBingCapture(routineSize int) Capture {
	header := map[string]string{
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
	}
	bc := &BingCapture{
		client: &http.Client{
			Transport: &http.Transport{
				MaxConnsPerHost: 10,
				MaxIdleConns:    5,
			},
			Timeout: 5 * time.Second,
		},
		baseUrl:  "https://cn.bing.com/images/async",
		headers:  header,
		q:        newQuery(),
		routines: routineSize,
	}
	bc.Downloader = newDownloader(bc.client, bc.headers)
	return bc.init()
}

func (bc *BingCapture) init() Capture {
	bc.q.Set("scenario", "ImageBasicHover")
	bc.q.Set("datsrc", "N_I")
	bc.q.Set("ch", "918")
	bc.q.Set("layout", "ColumnBased")
	bc.q.Set("mmasync", "1")
	bc.q.Set("count", "30")
	return bc
}
func (bc *BingCapture) RangeImages(keyword string, callBack func([]string) bool, opts ...Option) error {
	q := bc.q
	q.Set("q", keyword)
	for _, option := range opts {
		option(&q)
	}
	batchSize := 60
	// 任务超时时间
	timeout := 3 * time.Second
	// 必应拿不到这个数据
	total := 10000
	var collector = make(chan string, batchSize)
	for i := 0; i < total; i += batchSize {
		q.Set("first", strconv.Itoa(i))
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		queryURL := fmt.Sprintf("%s?%s", bc.baseUrl, q.Encode())
		go func() {
			defer close(collector)
			bc.searchBing(ctx, queryURL, collector)
		}()
		var urls = make([]string, 0, batchSize)
	WAIT:
		for {
			select {
			case <-ctx.Done():
				// 超时了
				break WAIT
			case url, ok := <-collector:
				if !ok {
					break WAIT
				}
				urls = append(urls, url)
			}
		}
		cancel()
		if !callBack(urls) {
			return nil
		}
		collector = make(chan string, batchSize)

	}
	return nil
}
func (bc *BingCapture) SearchImages(keyword string, maxNumber int, opts ...Option) ([]string, error) {
	pool, err := ants.NewPool(bc.routines)
	if err != nil {
		return nil, err
	}
	defer pool.Release()
	q := bc.q
	q.Set("q", keyword)
	for _, option := range opts {
		option(&q)
	}
	batchSize := 35
	var collector = make(chan string, maxNumber)
	// 设置单个任务的基础超时（例如 3 秒）
	baseTimeout := 10 * time.Second
	timeout := calculateTimeout(maxNumber, batchSize, bc.routines, baseTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	var wg sync.WaitGroup
	defer cancel()
	for i := 0; i < maxNumber; i += batchSize {
		q.Set("first", strconv.Itoa(i))
		queryURL := fmt.Sprintf("%s?%s", bc.baseUrl, q.Encode())
		wg.Add(1)
		pool.Submit(func() {
			bc.searchBing(ctx, queryURL, collector)
			wg.Done()
		})
	}
	var filter = make(map[string]struct{}, maxNumber)
	var urls = make([]string, 0, maxNumber)
	go func() {
		wg.Wait()
		close(collector)
	}()
SELECT:
	for {
		select {
		case url, ok := <-collector:
			if !ok {
				// 所有goroutine 执行完了，但是数量不够，任然要返回的
				break SELECT
			}
			if _, ok := filter[url]; !ok {
				filter[url] = struct{}{}
				urls = append(urls, url)
			}
			if len(urls) >= maxNumber {
				break SELECT
			}
		case <-ctx.Done():
			// 超时了,但是爬到的数据还是要给你的
			break SELECT
		}
	}
	return urls, nil
}

func (bc *BingCapture) searchBing(ctx context.Context, url string, collector chan<- string) {
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

		// 请求并解析 HTML
		resp, err := bc.client.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		doc, err := html.Parse(resp.Body)
		if err != nil {
			return
		}
		var pool, _ = ants.NewPool(bc.routines)
		wg := sync.WaitGroup{}
		// 函数递归解析并检查链接
		var f func(*html.Node)
		f = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "a" {
				for _, attr := range n.Attr {
					select {
					case <-ctx.Done():
						return
					default:
						if attr.Key == "class" && attr.Val == "iusc" {
							for _, attr := range n.Attr {
								select {
								case <-ctx.Done():
									return
								default:
									if attr.Key == "m" {
										byteData := []byte(attr.Val)
										wg.Add(1)
										pool.Submit(func() {
											defer func() {
												wg.Done()
											}()
											var bf BingFormat

											err = json.Unmarshal(byteData, &bf)
											if err != nil {
												return
											}
											if !bc.checkUseful(bf.Murl) {
												collector <- bf.Turl
												return
											}
											collector <- bf.Murl
										})
									}
								}
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
		wg.Wait()
		pool.Release()
	}
}
func (bc *BingCapture) checkUseful(url string) bool {
	if url == "" {
		return false
	}
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false
	}
	for k, v := range bc.headers {
		req.Header.Set(k, v)
	}
	resp, err := bc.client.Do(req)
	if nil != err {
		return false
	}
	defer resp.Body.Close()
	// 检查状态码是否为 2xx
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}
	return false
}
