package mysql

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"sync"
	"testing"
)

// TestConnWithoutDB 为用来测试数据库的详细连线流程，以下测试不使用 MariaDB 的服务器，只是单纯的单元测试
func TestConnWithoutDB(t *testing.T) {
	// 准备资料库的回应资料
	mysqlResponse := []uint8{
		// 资料长度
		93, 0, 0,
		// 自增序列号码
		0,
		// 以下 93 笔数据
		// 数据库的版本号
		10, 53, 46, 53, 46,
		53, 45, 49, 48, 46,
		53, 46, 49, 50, 45,
		77, 97, 114, 105, 97,
		68, 66, 45, 108, 111,
		103,
		// 数据库的版本结尾
		0,
		// 连线编号
		16, 0, 0, 0,
		// Salt
		81, 64, 43, 85, 76, 90, 97, 91,
		// filter
		0,
		// 取得功能标志
		254, 247,

		// 之后再处理
		33, 2, 0, 255, 129,
		21, 0, 0, 0, 0,
		0, 0, 15, 0, 0,
		0, 34, 53, 36, 85,
		93, 86, 117, 105, 49,
		87, 65, 125, 0, 109,
		121, 115, 113, 108, 95,
		110, 97, 116, 105, 118,
		101, 95, 112, 97, 115,
		115, 119, 111, 114, 100,
		0}

	// 测试开始
	t.Run("测试数据库连线的详细流程", func(t *testing.T) {
		// 先建立 pipe
		read, write := net.Pipe()

		// 实现缓存
		reader := bufio.NewReaderSize(read, connBufferSize) // 用来模拟 Gaea 读取数据
		writer := bufio.NewWriter(write)                    // 用来模拟数据库回传数据

		// 进行等待作业
		wg := sync.WaitGroup{}
		wg.Add(1)

		// 启动数据库
		go func() {
			_, err := writer.Write(mysqlResponse) // 回传给客户端
			require.Equal(t, err, nil)
			err = writer.Flush() // 把缓存资料写进 pipe
			require.Equal(t, err, nil)
			err = write.Close() // 资料写入完成，终结连线
			require.Equal(t, err, nil)

			// 工作完成
			wg.Done()
		}()

		// 读取开头的资讯
		var header [4]byte
		count, err := io.ReadFull(reader, header[:])
		require.Equal(t, count, 4)
		require.Equal(t, err, nil)

		// 等待中
		wg.Wait()

		// 取得自增序列号码
		sequence := header[3]
		require.Equal(t, sequence, uint8(0x0))

		// 取得版本号的资料范围
		length := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)
		require.Equal(t, length, 93)
		require.Equal(t, err, nil)

		// 取得版本号资讯
		data := make([]byte, length)
		count, err = io.ReadFull(reader, data)
		require.Equal(t, count, 93)
		require.Equal(t, err, nil)

		// 取得版本号结尾的位置
		pos := 1 + bytes.IndexByte(data[1:], 0x00) + 1

		// 取得版本号
		version := string(data[1 : pos-1])
		require.Equal(t, version, "5.5.5-10.5.12-MariaDB-log")

		// 取得连线编号
		connectionid := binary.LittleEndian.Uint32(data[pos : pos+4])
		require.Equal(t, connectionid, uint32(16))

		// 取得 Salt
		salt := data[pos+4 : pos+12]
		require.Equal(t, salt, []uint8{81, 64, 43, 85, 76, 90, 97, 91})

		// 取得功能标志
		capability := uint32(binary.LittleEndian.Uint16(data[pos+13 : pos+15]))
		require.Equal(t, capability, uint32(63486))

		// 之后再处理
	})
}
