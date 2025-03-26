package imagecapture

import (
	"context"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2024/11/2 下午10:55
* @Package:
 */

const maxDownloadRoutines = 8

type Downloader interface {
	// 下载图片到指定文件路径。如果 writer 不为 nil，将数据写入 writer；否则保存到 filename 文件中。
	// @param url: 图片 URL
	// @param filename: 保存的文件名
	// @param writer: 可选的 io.Writer 用于写入数据
	// @return: 下载成功返回文件名后缀  eg：[png],返回可能的错误
	Download(url, filename string, writer io.Writer) (string, error)
	//// 批量下载所有图片到指定目录，是否以图片的 MD5 值命名，返回已下载成功的文件路径。
	//// @param urls: 图片 URL 列表
	//// @param dir: 保存目录
	//// @param useMd5Naming: 是否使用 MD5 值命名
	//// @return: 返回已成功下载的文件路径列表和可能的错误
	BatchDownload(urls []string, dir string, useMd5Naming bool) ([]string, error)
}

// Downloader 包含重试和流控制属性
type downloader struct {
	client     *http.Client
	retryTimes int // 最大重试次数
	header     http.Header
	bufferSize int // 缓冲区大小
	md5        hash.Hash
	connPool   *sync.Pool    // 连接池
	timeout    time.Duration // 请求超时时间
}

// newDownloader 创建新的下载器
func newDownloader(client *http.Client, h map[string]string) Downloader {
	handle := &downloader{
		client:     client,
		retryTimes: 3,
		bufferSize: 64 * 1024, //64kb
		header:     make(http.Header, len(h)),
		timeout:    10 * time.Second,
		connPool: &sync.Pool{
			New: func() interface{} {
				return &http.Client{
					Timeout: 10 * time.Second,
					Transport: &http.Transport{
						MaxIdleConns:        100,
						MaxIdleConnsPerHost: 10,
						IdleConnTimeout:     90 * time.Second,
					},
				}
			},
		},
	}
	for k, v := range h {
		handle.header.Set(k, v)
	}
	return handle
}

func (d *downloader) get(url string, newWriter func(string) (io.Writer, error), mdCallback func(string)) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header = d.header

	// 使用连接池中的client
	client := d.connPool.Get().(*http.Client)
	defer d.connPool.Put(client)

	// 设置请求上下文和超时
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	req = req.WithContext(ctx)

	try := 0
	var resp *http.Response
	for {
		if try >= d.retryTimes {
			return ErrMaxRetryExceeded
		}
		resp, err = client.Do(req)
		if err == nil {
			break
		}
		try += 1
		time.Sleep(time.Duration(try) * 100 * time.Millisecond)
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("download [%s] failed: %s\n", url, resp.Status)
			return fmt.Errorf("download [%s] failed: %s", url, resp.Status)
		}
		defer resp.Body.Close()
		imageReader, err := NewImageReader(resp.Body, mdCallback != nil)
		if err != nil {
			return err
		}
		writer, err := newWriter(imageReader.Type())
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, imageReader)
		if err != nil {
			return err
		}
		if mdCallback != nil {
			mdCallback(imageReader.Md5())
		}
		return nil
	}
	return nil
}

// create io.writer
func newFileWriter(filename string) (*os.File, error) {
	if filename == "" {
		return nil, ErrInvalidTargetPath
	}
	// check the file if exist
	if _, err := os.Stat(filename); err == nil {
		return nil, ErrFileAlreadyExists // already exit
	} else if !os.IsNotExist(err) {
		return nil, err // Other errors, such as permission problems
	}
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}
	// Open the file in create mode with permissions set to 0644
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return nil, fmt.Errorf(ErrFileCreationFailed.Error()+": %w", err)
	}
	return file, nil
}

func (d *downloader) Download(url, filename string, writer io.Writer) (fileSuffix string, err error) {
	var release func()
	err = d.get(url, func(suffix string) (io.Writer, error) {
		fileSuffix = suffix
		if writer == nil {
			// 写文件
			writer, err = newFileWriter(fmt.Sprintf("%s.%s", filename, suffix))
			if nil != err {
				return nil, err
			}
			release = func() {
				err = writer.(*os.File).Close()
				if err != nil {
					fmt.Println("failed to close file:", err)
				}
			}
		}
		return writer, nil
	}, nil)
	if release != nil {
		release() // 确保打开的文件在完成后被关闭
	}
	return
}

func (d *downloader) BatchDownload(urls []string, dir string, useMd5Naming bool) ([]string, error) {
	// firstly created dir
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}
	pool, _ := ants.NewPool(maxDownloadRoutines)
	defer pool.Release()
	var collector = make(chan string, maxDownloadRoutines)
	var wg sync.WaitGroup
	var paths = make([]string, 0, len(urls))
	// 设置单个任务的基础超时（例如 3 秒）
	baseTimeout := 5 * time.Second
	timeout := calculateTimeout(len(urls), 1, maxDownloadRoutines, baseTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	defer cancel()
	for i := range urls {
		var url = urls[i]
		wg.Add(1)
		err := pool.Submit(func() {
			defer wg.Done()
			d.saveFile(url, dir, useMd5Naming, collector)
		})
		if err != nil {
			return nil, err
		}
	}
	go func() {
		wg.Wait()
		close(collector)
	}()
SELECT:
	for {
		select {
		case path, ok := <-collector:
			if !ok {
				break SELECT
			}
			paths = append(paths, path)
		case <-ctx.Done():
			break SELECT
		}
	}
	return paths, nil
}

func (d *downloader) saveFile(url, dir string, useMd5Naming bool, collector chan<- string) {
	var release func()
	uuid, err := GenerateUUID()
	if err != nil {
		return
	}
	var filename string
	var md5Callback func(string)
	var imageSuffix string
	if useMd5Naming {
		md5Callback = func(md5 string) {
			oldName, newName := filename, fmt.Sprintf("%s/%s.%s", dir, md5, imageSuffix)
			err = os.Rename(oldName, newName)
			if err != nil {
				fmt.Println("failed to rename file:", err)
			}
			filename = newName
		}
	}
	err = d.get(url, func(suffix string) (io.Writer, error) {
		imageSuffix = suffix
		filename = fmt.Sprintf("%s/%s.%s", dir, uuid, suffix)
		writer, err := newFileWriter(filename)
		if err != nil {
			return nil, err
		}
		release = func() {
			err = writer.Close()
			if err != nil {
				fmt.Println("failed to close writer:", err)
			}
		}
		return writer, nil
	}, md5Callback)
	defer func() {
		if release != nil {
			release()
		}
	}()
	if err != nil {
		return
	}
	collector <- filename
}
