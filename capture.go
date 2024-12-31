package imagecapture

import (
	"fmt"
	"net/url"
	"strings"
)

var defaultFilterRules = []Rule{RULE_DOUYIN, RULES_SINA}

type Rule []string

// 规则校验
func (r Rule) Check(rawURL string) bool {
	// 解析 URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return false // URL 解析失败
	}
	// 获取域名部分
	host := u.Hostname()
	// 判断域名是否属于抖音
	for _, rule := range r {
		if strings.HasSuffix(host, rule) {
			return true
		}
	}
	return false
}

var (
	//  抖音源
	RULE_DOUYIN = Rule{
		"douyin.com",
		"douyinpic.com",
		"ixigua.com",
		"snssdk.com",
		"amemv.com",
		"tiktok.com",
	}
	// 新浪源
	RULES_SINA = Rule{
		"sinaimg.cn",
		"sinajs.cn",
		"sina.com.cn",
		"vip.sina.com",
	}
)

/*
* @Author: zouyx
* @Email:1003941268@qq.com
* @Date:   2024/11/1 下午6:25
* @Package:
 */
type Capture interface {
	Downloader
	/**
	搜索图片： 注 【对图片去重后返回的图片是无序的】
	@param keywords: 搜索关键词
	@param maxNumber: 最多返回的图片数量
	@param opts: 额外参数，支持多种选项
	@return: 返回图片 URL 列表和可能的错误
	*/
	SearchImages(keyword string, maxNumber int, opts ...Option) ([]string, error)

	/**
	分页范围图片搜索：用于分批获取搜索结果
	@param keyword: 搜索关键词
	@param callBack: 回调函数，用于处理返回的每一批次图片 URL 列表的函数。如果 `fn` 返回 `false`，则停止后续搜索
	@param opts: 额外参数，支持多种选项
	@return: 返回当前页码的图片 URL 列表和可能的错误
	*/
	RangeImages(keyword string, callBack func(urls []string) bool, opts ...Option) error
}

type Option func(*query)

type query struct {
	url.Values
}

func newQuery() query {
	return query{url.Values{}}
}

// WithCopyright 过滤版权数据
func WithCopyright() Option {
	return func(query *query) {
		query.Set("copyright", "1")
	}
}

// WithImageSize 搜索的图片大小限制
func WithImageSize(size ImageSize) Option {
	return func(query *query) {
		query.Set("z", fmt.Sprintf("%d", size))
	}
}

// WithLatest 搜索最新图片
func WithLatest() Option {
	return func(query *query) {
		query.Set("latest", "1")
	}
}

// WithGif 搜索动图
func WithGif() Option {
	return func(query *query) {
		query.Set("lm", "6")
	}
}

// WithHd 搜素高清图
func WithHd() Option {
	return func(query *query) {
		query.Set("hd", "1")
	}
}
