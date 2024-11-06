package imagecapture

import (
	"hash"
	"io"
	"reflect"
	"testing"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2024/11/6 下午2:30
* @Package:
 */

func TestImageReader_Md5(t *testing.T) {
	type fields struct {
		Reader io.Reader
		ty     string
		md5    hash.Hash
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ir := ImageReader{
				Reader: tt.fields.Reader,
				ty:     tt.fields.ty,
				md5:    tt.fields.md5,
			}
			if got := ir.Md5(); got != tt.want {
				t.Errorf("Md5() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageReader_Read(t *testing.T) {
	type fields struct {
		Reader io.Reader
		ty     string
		md5    hash.Hash
	}
	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantN   int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ir := ImageReader{
				Reader: tt.fields.Reader,
				ty:     tt.fields.ty,
				md5:    tt.fields.md5,
			}
			gotN, err := ir.Read(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotN != tt.wantN {
				t.Errorf("Read() gotN = %v, want %v", gotN, tt.wantN)
			}
		})
	}
}

func TestImageReader_Type(t *testing.T) {
	type fields struct {
		Reader io.Reader
		ty     string
		md5    hash.Hash
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ir := ImageReader{
				Reader: tt.fields.Reader,
				ty:     tt.fields.ty,
				md5:    tt.fields.md5,
			}
			if got := ir.Type(); got != tt.want {
				t.Errorf("Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewImageReader(t *testing.T) {
	type args struct {
		reader      io.Reader
		callBackMd5 bool
	}
	tests := []struct {
		name    string
		args    args
		want    *ImageReader
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewImageReader(tt.args.reader, tt.args.callBackMd5)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewImageReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewImageReader() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkImageType(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkImageType(tt.args.data); got != tt.want {
				t.Errorf("checkImageType() = %v, want %v", got, tt.want)
			}
		})
	}
}
