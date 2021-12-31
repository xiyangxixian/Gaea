package xlog

import (
	"fmt"
	fileClient "github.com/XiaoMi/Gaea/models/logStorage/file"
)

// LogStorageClient 用于日志的客户端接口
type LogStorageClient interface {
	ReOpen() error                 // 开启日志的输出
	Write(logText []byte) error    // 正式写入日志
	WriteErr(logText []byte) error // 正式写入错误日志
	Close() error                  // 关闭日志的输出
}

// LogStorage 将会做出供日志储存使用的共同介面
type LogStorage struct {
	client LogStorageClient
}

// NewLogStorageClient 會建立一個全新的日志储存客户端
// fileName 参数如果为多档输出时，可以用逗号隔开，比如
func NewLogStorageClient(config map[string]string) *LogStorage {
	// 决定日志输出
	storage, ok := config["storage"]
	if !ok {
		storage = "channel"
	}

	switch storage {
	case "channel":
		// channel 为仿真用的双向通道
		// 先略过
	case "console":
		// 先略过
	case "file", "multiFile":
		// file 为单档输出新建对象的内容
		c, err := fileClient.New(config)
		if err != nil {
			fmt.Printf("create fileClient failed, %v\n", err)
			return nil
		}
		return &LogStorage{client: c}
	}
	// log.Fatal("unknown config type")
	return nil
}