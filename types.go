package imagecapture

/*
* @Author: zouyx
* @Email:1003941268@qq.com
* @Date:   2024/11/1 下午5:58
* @Package:
 */

/*
需要搜索的图片大小
*/
type ImageSize uint

const (
	_                ImageSize = iota
	ImageSize_SMALL            // 小图
	ImageSize_MEDIUM           // 中图
	ImageSize_LARGE            // 大图
	_
	_
	_
	_
	_
	ImageSize_ENORMOUS // 特大图片
)
