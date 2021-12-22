# 日志分流说明 multiFile

> 1. 原本小米的程式码是支援把日志写入到 console 和 档案里，现在在基于其基础，把日志进行分流，程式码有不同的区块，每个区块会写入不同的档案
> 2. 小米的设定档中，在 etc 目录下，有 namespace 资料夹，如果能做到每个 namespace 的日志都分别集中到一个日志档，分别有自己的日志档，是很实用的功能

## 1 日志分流程式

> 整个程式码在编写的过程中，会尽量把档案 multiFile.go 里面的程式码尽量依附在 file.go 里，这样子可以避免重复的程式码

### 1-1 设定档的设定方式

这里是要比较 把日志写入档案 (在设定值为 log_output=file) 和 把日志写入档案并分流  (在设定值为 log_output=multiFile)，在设定档细节的不同

#### 1-1-1 temp

//

#### 1-1-2 原设定值

在把 日志写入档案 (在设定值为 log_output=file)，其设定档如下 (在这次新增和修改程式码后，持续支援此设定值)

在 /etc/gaea.ini 设定档，设定值如下

```ini
;log config
log_path=./logs
log_level=Notice
log_filename=gaea
log_output=file
```

在把 日志写入档案并分流 (在设定值为 log_output=multiFile)，其设定档如下

#### 1-1-3 新设定值 (统一设定)

一定要有两个以上的设定档，才能做日志的分流，

如果只有一个档案，那就不如使用 设定值为 log_output=file

反正如果要做日志分流，至少要设定两个以上的日志档案

在 /etc/gaea.ini 设定档，设定值如下

```ini
;log config
log_path=./logs
log_level=Notice
log_filename=default,log1
service_name=svc1
log_output=multiFile
```

- 在这个设定档，会自产生两个日志档 logs/default.log 和 logs/log1.log，

- 因为等级只有一个值为 Notice，所有的档案都采用此值作为预设值，都为为 Notice
- 服务名称为统一设定，logs/default.log 和 logs/log1.log 的服务名称都为 svc1

#### 1-1-4 新设定值 (各别设定)

一样跟 版本一 一样，要有两个以上的设定档，才能做日志的分流

在 /etc/gaea.ini 设定档，设定值如下

```ini
;log config
log_path=./logs
log_level=Notice,debug
log_filename=default,log1
service_name=svc1,svc2
log_output=multiFile
```

- 这时会产生两个日志档 logs/default.log 和 logs/log1.log，

- 但是这次不同，使用两个等级，一个为 Notice，另一个为 Debug
  相对应，logs/default.log 就会使用等级 Notice，logs/log1.log 就会使用等级 Debug
- 服务名称为各别设定，logs/default.log 的服务名称都为 svc1，logs/log1.log 的服务名称都为 svc2

#### 预设档案

预设日志档案为设定值 log_filename=default,log1 的第一个元素，也就是 default.log

当日志找不到指定要写入的日志档案时，会直接写入预设的日志档案，也就是default.log

#### 指定日志档案的方式

在程式码中，可以指定要写入的日志档案

指定方式如下

1. 指定格式如下，在 format 参数里，以双分号 :: 为界线
2. 双分号 :: 以前的为 指定的日志档案
3. 双分号 :: 以后的为 指定的日志格式，%s 也加在这里

设定例子如下

```go
// 在 /etc/gaea.ini 设定档，设定值 指定
// log_filename=default,log1
// log_output=multiFile

err = ps.Notice("record1") // 在没有指定写入的日志档时，会先写到预设的日志档里，也就是 default.log
err = ps.Notice("log1::record2") // 有指定写入日志档 log1.log，所以会把日志写到档案 log1.log 里
err = ps.Notice("log2::record3") // 有指定写入日志档 log2.log，但是 log2.log 并不存在于 gaea.ini 的设定里，所以只能把日志写到预设的日志档里，也就是 default.log
```

### 1-2 程式的设计逻辑

#### 1-2-1 日志编号的问题

不管是 (1) 终端机 console 、(2) 档案 file 还是 (3) 多档案输出 multiFile，都会有日志编号

比如在

- (1) 终端机 console 显示的日志编号为 800000001
- (2) 档案 file 显示的日志编号为 900000001

这次所修改的程式，一开始就指定多档输出的日志编号为 1000000001

- (3) 档案 multiFile 显示的日志编号为 1000000001

虽然 多档输出 multiFile 的程式码是依附在 档案 file，但是日志编号还是依然可以正常指定，因为原本程式就有预留可以把传入 日志编号 参数的地方

```go
// Notice 显示 Notice 的资讯，格式为 档名::日志格式为第一个参数
func (ps *XMultiFileLog) Notice(format string, a ...interface{}) error {
    // >>>>> >>>>> >>>>> >>>>> >>>>> 预先处理的部份
    
	// 先拆分 format 字串
	logFile, newFormat := ps.prepareMultiFile(format)
    
    // >>>>> >>>>> >>>>> >>>>> >>>>> 保留原程式的部份

	// 以下程式码尽量保留
	if ps.multi[logFile].level > NoticeLevel {
		return nil
	}

    // 传入新的 newFormat 参数 (最后在这里指定日志编号)
	return ps.multi[logFile].noticex(XMultiFileLogDefaultLogID, newFormat, a...)
}
```

以上程式码进行以下拆解

- 预先处理的部份
  在这个部份会把 format 字串，拆分成

  ​     1 日志档案变数 logFile
  ​     2 新的 format 格式字串 newFormat

  之后做后续处理

- 保留原程式的部份
  在这个部份整个逻辑会和 file.go 的程式码很像
  就在最后 noticex 函式里，会传入日志编号的参数 XMultiFileLogDefaultLogID 值为 1000000001

#### 1-2-2 多档输出的物件设计

物件的程式码如下

```go
// XMultiFileLog is the multi file logger
type XMultiFileLog struct {
	// 这里设定预设写入日志的档名，不在这里作日志 Log 分流的相关设定，避免进行上锁
	defaultXLog string               // 预设要写入的日志档案
	multi       map[string]*XFileLog // 多档案的输出
}
```

- 有一个 defaultXLog 会记录预设要写入的预设日志档案，当程式无法得知要写入那个日志档案时，就会直接把日志写入日志档案里

- multi 元素把 档案输出 file 物件 和其档案名称做对应，

  也就是 XMultiFileLog (多档日志输出) 是由多个 XFileLog (单档日志输出) 所组成的

#### 1-2-3 多档输出的物件初始化

主程式 main 会配有一个 副程式 initXLog，副程式 initXLog 的作用如下：

- 整理和组成 init 函式的 config 参数
- 把现在的 日志设定值 设置 成全域变数

#### 1-2-4 initXLog 程式码如下

- 将会组成 cfg 变数后，依照设定值会传入任一个 console、file 和 multiFile 的 init 的方法
- 所以知道单元测试的剖面会切在 副程式 initXLog 的后面
- 其实一直会有想把 cfg["filename"] = filename 这一行中的 filename 字串改成驼峰式命名 fileName，也就是
   cfg["filename"] = filename 改成 cfg["fileName"] = fileName，但后面看起来好像不协调，因为整个 initXLog 函式里将会只有 filename 很明显使用驼峰式命名 fileName，就先算了 

```go
func initXLog(output, path, filename, level, service string) error {
	cfg := make(map[string]string)
	cfg["path"] = path
	cfg["filename"] = filename
	cfg["level"] = level
	cfg["service"] = service
	cfg["skip"] = "5"
    // 设置xlog打印方法堆栈需要跳过的层数, 5目前为调用log.Debug()等方法的方法名, 比xlog默认值多一层.

	logger, err := xlog.CreateLogManager(output, cfg)
	if err != nil {
		return err
	}

	log.SetGlobalLogger(logger) // 设定成全域日志变数
	return nil
}
```

#### 1-2-5 副程式 init 程式码如下

> 多档输出 multiFuile 的副程式 init 程式码如下

```go
func (ps *XMultiFileLog) Init(config map[string]string) (err error) {
	// 先初始化 ps 的 multi 的对应 map
	ps.multi = make(map[string]*XFileLog)

	// 有三个设定值使用逗号，分别是 fileName，service 和 level，要特别处理

	// 产生 fileName 阵列
	var filename []string
	fStr, ok := config["filename"] // 先确认 fileName 设定值是否存在
	if ok {                        // 如果 fileName 值 存在
		filename = strings.Split(fStr, ",") // 以逗点分隔开来
	}
	if !ok { // 如果 fileName 值 不 存在
		err = fmt.Errorf("init XFileLog failed, not found filename")
		return
	}
    
    // 以下程式码(略过)
}
```

日志档分流 XMultiFileLog 设定值整理成以下的表格

| initXLog 设定值 | 可设定多值 | 可设定单值 | 必须存在 | 说明     |
| --------------- | ---------- | ---------- | -------- | -------- |
| path            | 不可以     | 一定要     | 必须     | 日志路径 |
| filename        | 支援       | 不支援     | 必须     | 分流档案 |
| service         | 支援       | 支援       | 必须     | 服务名称 |
| level           | 支援       | 支援       | 必须     | 日志等级 |

- 关于 path 设定，当在 gaea.ini 指定日志路径为 ./logs 时，当在 /home/panhong/go/src/github.com/panhongrainbow/Gaea 目录执行时，将会在 /home/panhong/go/src/github.com/panhongrainbow/Gaea/logs 目录下写资料

## 2 单元测试的运作方式