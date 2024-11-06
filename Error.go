package imagecapture

import "errors"

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2024/11/4 上午10:48
* @Package:
 */
var (
	ErrDownloadFailed   = errors.New("failed to download image")        // 下载失败
	ErrMaxRetryExceeded = errors.New("maximum retry attempts exceeded") // 达到最大重试次数
	// 网络连接或URL相关错误
	ErrInvalidURL       = errors.New("invalid URL")        // URL格式不正确
	ErrConnectionFailed = errors.New("connection failed")  // 网络连接失败
	ErrTimeout          = errors.New("connection timeout") // 连接超时
	ErrHostUnreachable  = errors.New("host unreachable")   // 无法访问主机

	// 文件操作相关错误
	ErrFileCreationFailed    = errors.New("file creation failed")    // 创建文件失败
	ErrFileWriteFailed       = errors.New("file write failed")       // 文件写入失败
	ErrFilePermissionDenied  = errors.New("file permission denied")  // 文件权限不足
	ErrDiskSpaceInsufficient = errors.New("insufficient disk space") // 磁盘空间不足

	// 数据读取与解码相关错误
	ErrReadFailed         = errors.New("failed to read data from source") // 数据读取失败
	ErrDataDecodingFailed = errors.New("failed to decode data")           // 数据解码失败

	// 目标路径相关错误
	ErrInvalidTargetPath    = errors.New("invalid target path")           // 目标路径不正确
	ErrDirectoryNotWritable = errors.New("target directory not writable") // 目标目录无法写入
	ErrFileAlreadyExists    = errors.New("file already exists")           // 文件已存在

	// 下载内容相关错误
	ErrUnsupportedFileType     = errors.New("unsupported file type")      // 文件类型不支持
	ErrContentTooLarge         = errors.New("content size exceeds limit") // 内容大小超过限制
	ErrContentChecksumMismatch = errors.New("content checksum mismatch")  // 校验和不匹配，可能数据损坏
)
