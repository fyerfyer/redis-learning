package utils

import (
	"log"
	"runtime"
	"strings"
)

//// ConvertToBytes 将任意类型转换为字节数组
//func ConvertToBytes(value interface{}) ([]byte, error) {
//	if value == nil {
//		return nil, fmt.Errorf("cannot convert nil value")
//	}
//
//	// 如果已经是字节数组，直接返回
//	if bytes, ok := value.([]byte); ok {
//		return bytes, nil
//	}
//
//	// 使用JSON序列化其他类型
//	return json.Marshal(value)
//}
//
//// ConvertFromBytes 将字节数组转换为指定类型
//func ConvertFromBytes(bytes []byte, result interface{}) error {
//	return json.Unmarshal(bytes, result)
//}
//
//// FormatCacheKey 格式化缓存键名，可以用于添加前缀或格式化键名
//func FormatCacheKey(prefix, key string) string {
//	if prefix == "" {
//		return key
//	}
//	return prefix + ":" + key
//}

// LogError 记录错误日志，包含文件和行号
func LogError(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	// 提取文件名（不包含路径）
	parts := strings.Split(file, "/")
	fileName := parts[len(parts)-1]

	log.Printf("[ERROR] %s:%d - "+format, append([]interface{}{fileName, line}, v...)...)
}

// LogInfo 记录普通信息日志
func LogInfo(format string, v ...interface{}) {
	log.Printf("[INFO] "+format, v...)
}

//// TruncateDuration 确保持续时间不小于最小值且不大于最大值
//func TruncateDuration(d, min, max time.Duration) time.Duration {
//	if d < min {
//		return min
//	}
//	if d > max {
//		return max
//	}
//	return d
//}

//// IsZeroOrNegative 检查持续时间是否为零或负数
//func IsZeroOrNegative(d time.Duration) bool {
//	return d <= 0
//}
