# `imagecapture` - 图片抓取 工具

<p align="center">
  <a href="https://github.com/code-innovator-zyx/imagecapture">
   <img alt="ChopperBot" src="https://github.com/twj666/ChopperBot-Doc/blob/master/img/logo.png?raw=true">
  </a>
</p>

<p align="center">
  <strong>图片捕获器</strong>
</p>


<p align="center">
  <a href="https://github.com/code-innovator-zyx/wechat-gptbot/blob/main/README.md">
    <img src="https://img.shields.io/badge/文档-简体中文-blue.svg" alt="简体中文文档" />
  </a>

  <a target="_blank" href='https://github.com/code-innovator-zyx/imagecapture'>
        <img src="https://img.shields.io/github/stars/code-innovator-zyx/imagecapture.svg" alt="github stars"/>
   </a>

   <a target="_blank" href=''>
        <img src="https://img.shields.io/badge/Process-Developing-yellow" alt="github stars"/>
   </a>
</p>

`imagecapture` ImageCapture 是一个用 Go 语言编写的库，旨在从百度和必应等搜索引擎捕获图片。它提供了一个接口，用于搜索和下载图片，并支持多种自定义选项。

## 特性

- **多引擎支持**：支持百度、必应，后续将添加 Google 搜索。
- **高级筛选**：支持根据版权、图片尺寸、动图等进行筛选。
- **并发抓取**：使用并发抓取功能，提高图片抓取效率。
- **去重功能**：自动去重，确保返回的图片 URL 唯一。
- **分页迭代功能**：- 支持大批量图片的分页获取。。

## 安装

通过 `go get` 安装该工具包：

```bash
go get github.com/code-innovator-zyx/imagecapture
```

## 快速开始

### 初始化 BaiduCapture

```go
package main

import (
	"fmt"
	"github.com/code-innovator-zyx/imagecapture"
	"log"
)

func main() {
	keyword := "美女"
	maxImageNums := 20
	// 新建一个百度图片捕获器  routineSize 限制协爬取的携程池数量
	baiduCapture := imagecapture.NewBaiduCapture(5)
	// 搜索图片
	urls, err := baiduCapture.SearchImages(keyword, maxImageNums)
	if err != nil {
		log.Fatalln(err.Error())
	}
	filename := "./beautiful"
	// 可以使用内置下载器下载图片   注：文件后缀会根据图片真是类型进行判断
	suffix, err := baiduCapture.Download(urls[0], filename, nil)
	if err != nil {
		log.Fatalln(err.Error())
	}
	fmt.Println(suffix)
}

```

### 初始化 BingCapture

```go
package main

import (
	"fmt"
	"github.com/code-innovator-zyx/imagecapture"
	"log"
)

func main() {
	keyword := "美女"
	maxImageNums := 20
	// 新建一个必应图片捕获器  routineSize 限制协爬取的携程池数量
	bingCapture := imagecapture.NewBingCapture(5)
	// 搜索图片
	urls, err := bingCapture.SearchImages(keyword, maxImageNums)
	if err != nil {
		log.Fatalln(err.Error())
	}
	filename := "./beautiful"
	// 可以使用内置下载器下载图片   注：文件后缀会根据图片真是类型进行判断
	suffix, err := bingCapture.Download(urls[0], filename, nil)
	if err != nil {
		log.Fatalln(err.Error())
	}
	fmt.Println(suffix)
}
```

## 主要功能

## SearchImages

用于在指定搜索引擎中根据关键词搜索图片。

#### 参数

- `keyword` (string): 搜索关键词。
- `maxNumber` (int): 要返回的最多图片数量。
- `opts` (Option): 可选参数，用于指定其他筛选条件（例如图片尺寸、是否高清、动图等）。

#### 示例

```go
// 使用 WithImageSize、WithHd 等选项来进行筛选
images, err := baiduCapture.SearchImages("sunrise", 20, imagecapture.WithHd(), imagecapture.WithImageSize(imagecapture.Medium))
```

## RangeImages

用于在指定搜索引擎中根据关键词持续搜索图片。

#### 参数

- `keyword` (string): 搜索关键词。
- `callBack` (func(string)bool): 每一批图片的回调函数。
- `opts` (Option): 可选参数，用于指定其他筛选条件（例如图片尺寸、是否高清、动图等）。

#### 示例

```go
capture.RangeImages("老虎", func (urls []string) bool {
return true
})
if err != nil {
t.Error(err.Error())
return
}
})

```

> [更多案例](https://github.com/code-innovator-zyx/imagecapture/tree/main/test)

## 支持的筛选选项

### 仅百度搜索支持以下筛选选项：

#### 1. `WithCopyright()`

过滤版权问题的图片，仅返回无版权限制的图片。

#### 2. `WithImageSize(size ImageSize)`

限制搜索图片的大小。`ImageSize` 可以是以下几种：

- `Small`：小尺寸
- `Medium`：中等尺寸
- `Large`：大尺寸

#### 3. `WithLatest()`

搜索最新的图片，仅返回最近上传或更新的图片。

#### 4. `WithGif()`

搜索动图，返回 `.gif` 格式的图片。

#### 5. `WithHd()`

搜索高清图

## 图片去重

工具 内部会使用 `map` 来去重 URL，确保每个返回的 URL 唯一。这样可以避免重复图片 URL 出现在结果中。

## 配置

### 配置并发度

`BaiduCapture` 和 `BingCapture` 都可以通过传入并发数量来配置并发度，最多支持 6 个并发。

```go
bingCapture := imagecapture.NewBaiduCapture(6) // 最大并发6
```

## 免责声明

本项目仅用于个人学习、研究和开发目的，禁止用于任何非法用途或商业用途。使用本 库 进行的所有操作和行为由用户自行承担风险。

- 本 库 的图片抓取功能仅适用于合法的数据抓取用途，用户应遵守相关法律法规。
- 本 库 使用的第三方图片搜索引擎（如百度、必应等）可能会随时更改其接口或数据访问策略，使用时需自行留意相关的变化。
- 本项目不对通过 库 抓取的任何内容的版权、合法性等问题承担任何责任。

使用本库即表示用户同意并遵守上述条款。

# License

This project is licensed under the MIT License - see the LICENSE file for details.