package imagecapture

import (
	"bytes"
	"hash"
	"io"
	"net/http"
	"os"
	"reflect"
	"testing"
)

/*
* @Author: zouyx
* @Email: zouyx@knowsec.com
* @Date:   2024/11/6 下午4:40
* @Package:
 */

func Test_downloader_BatchDownload(t *testing.T) {
	type fields struct {
		client     *http.Client
		retryTimes int
		header     http.Header
		bufferSize int
		md5        hash.Hash
	}
	type args struct {
		urls         []string
		dir          string
		useMd5Naming bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &downloader{
				client:     tt.fields.client,
				retryTimes: tt.fields.retryTimes,
				header:     tt.fields.header,
				bufferSize: tt.fields.bufferSize,
				md5:        tt.fields.md5,
			}
			got, err := d.BatchDownload(tt.args.urls, tt.args.dir, tt.args.useMd5Naming)
			if (err != nil) != tt.wantErr {
				t.Errorf("BatchDownload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BatchDownload() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_downloader_Download(t *testing.T) {
	type fields struct {
		client     *http.Client
		retryTimes int
		header     http.Header
		bufferSize int
		md5        hash.Hash
	}
	type args struct {
		url      string
		filename string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantWriter string
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &downloader{
				client:     tt.fields.client,
				retryTimes: tt.fields.retryTimes,
				header:     tt.fields.header,
				bufferSize: tt.fields.bufferSize,
				md5:        tt.fields.md5,
			}
			writer := &bytes.Buffer{}
			err := d.Download(tt.args.url, tt.args.filename, writer)
			if (err != nil) != tt.wantErr {
				t.Errorf("Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t.Errorf("Download() gotWriter = %v, want %v", gotWriter, tt.wantWriter)
			}
		})
	}
}

func Test_downloader_get(t *testing.T) {
	type fields struct {
		client     *http.Client
		retryTimes int
		header     http.Header
		bufferSize int
		md5        hash.Hash
	}
	type args struct {
		url        string
		newWriter  func(string) (io.Writer, error)
		mdCallback func(string)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &downloader{
				client:     tt.fields.client,
				retryTimes: tt.fields.retryTimes,
				header:     tt.fields.header,
				bufferSize: tt.fields.bufferSize,
				md5:        tt.fields.md5,
			}
			if err := d.get(tt.args.url, tt.args.newWriter, tt.args.mdCallback); (err != nil) != tt.wantErr {
				t.Errorf("get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_downloader_saveFile(t *testing.T) {
	type fields struct {
		client     *http.Client
		retryTimes int
		header     http.Header
		bufferSize int
		md5        hash.Hash
	}
	type args struct {
		url          string
		dir          string
		useMd5Naming bool
		collector    chan<- string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &downloader{
				client:     tt.fields.client,
				retryTimes: tt.fields.retryTimes,
				header:     tt.fields.header,
				bufferSize: tt.fields.bufferSize,
				md5:        tt.fields.md5,
			}
			d.saveFile(tt.args.url, tt.args.dir, tt.args.useMd5Naming, tt.args.collector)
		})
	}
}

func Test_newDownloader(t *testing.T) {
	type args struct {
		client *http.Client
		h      map[string]string
	}
	tests := []struct {
		name string
		args args
		want Downloader
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newDownloader(tt.args.client, tt.args.h); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newDownloader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newFileWriter(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    *os.File
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newFileWriter(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("newFileWriter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newFileWriter() got = %v, want %v", got, tt.want)
			}
		})
	}
}
