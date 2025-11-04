# Go工具库项目规范

## 1. 项目概述
本项目是一个Golang工具库集合，提供了包括数据库操作、缓存、加密、HTTP客户端、错误处理、配置管理等多种常用功能模块，旨在为Go项目开发提供基础支持。

## 2. 目录结构规范

### 2.1 包结构
项目采用功能模块化的包结构，每个主要功能模块都有独立的包目录：
- 功能单一：每个包专注于解决某一类特定问题
- 命名简洁：包名使用小写字母，避免使用下划线或混合大小写
- 职责清晰：不同功能的代码放置在不同的包中

```
├── auth/        # 认证相关功能
├── buffer/      # 缓冲区操作
├── cache/       # 缓存实现
├── client/      # 各类客户端实现
├── config/      # 配置管理
├── email/       # 邮件发送功能
├── encoding/    # 编码解码
├── errcode/     # 错误码定义
├── helper/      # 工具函数
├── middleware/  # HTTP中间件
├── process/     # 进程管理
├── redis/       # Redis客户端封装
├── rest/        # RESTful API相关工具
├── rpc/         # RPC实现
├── sign/        # 签名加密
└── ...          # 其他功能模块
```

### 2.2 文件组织
- 每个包包含主要功能文件和测试文件
- 测试文件命名为`*_test.go`
- 核心逻辑与辅助功能分离

## 3. 命名规范

### 3.1 包命名
- 包名应短小精悍，使用小写字母
- 避免使用下划线或混合大小写
- 包名应能清晰表达其功能，如`httpclient`、`cache`等

### 3.2 结构体与类型命名
- 使用大驼峰命名法（PascalCase）
- 类型名应能清晰表达其用途，如`Config`、`Client`等

```go
// 正确示例
type Config struct {}
type RedisClient struct {}
```

### 3.3 函数与方法命名
- 使用大驼峰命名法（PascalCase）
- 函数名应描述其行为，如`NewClient`、`GetJSON`等
- 短小精悍，通常由动词+名词组成

```go
// 正确示例
func NewClient() *Client {}
func GetJSON(ctx context.Context, url string) ([]byte, error) {}
```

### 3.4 常量命名
- 使用大驼峰命名法（PascalCase）或全大写加下划线
- 常量名应清晰表达其含义

```go
// 正确示例
const DefaultTimeout = 3 * time.Second
const ContentTypeJSON = "application/json; charset=utf-8"
```

### 3.5 变量命名
- 局部变量使用小驼峰（camelCase）
- 全局变量使用大驼峰（PascalCase）
- 简洁明了，避免过于冗长的变量名

## 4. 代码风格规范

### 4.1 注释规范
- 包级别注释：每个包应有简短的包级别注释，说明包的用途
- 函数注释：公共函数应添加注释，说明其功能、参数和返回值
- 重要逻辑注释：复杂的业务逻辑应有适当的注释
- 注释语言：统一使用英文

```go
// httpclient 包提供了HTTP客户端的封装，支持GET、POST等请求方法
package httpclient

// GetJSON 发送GET请求并返回JSON数据
// ctx: 上下文，用于控制请求超时等
// url: 请求URL地址
// options: 可选配置项
// return: 返回JSON字节数组和错误信息
func GetJSON(ctx context.Context, url string, options ...Option) ([]byte, error) {
    // 实现代码
}
```

### 4.2 代码格式
- 遵循Go标准代码格式
- 使用`gofmt`或`goimports`工具格式化代码
- 缩进使用制表符（Tab）
- 每行代码长度控制在合理范围内，避免过长

### 4.3 代码组织
- 相关的代码放在一起
- 函数定义顺序合理，通常按依赖关系排序
- 导出与非导出函数分组放置

## 5. 错误处理规范

### 5.1 错误返回
- 函数返回错误作为最后一个返回值
- 错误信息应清晰明了，包含足够的上下文
- 不忽略错误，及时处理或向上传递

```go
// 正确示例
func OpenFile(path string) (*os.File, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("open file %s failed: %w", path, err)
    }
    return file, nil
}
```

### 5.2 自定义错误
- 使用`errcode`包中定义的错误类型或创建新的错误类型
- 错误应包含错误码和错误信息

```go
// 使用预定义错误
if err != nil {
    return nil, errcode.ErrInvalidParam
}

// 创建新的错误
return nil, errcode.NewError(10001, "invalid parameter")
```

### 5.3 Panic处理
- 在公共API中避免使用panic
- 对于可能发生panic的操作，使用defer和recover进行处理
- 提供安全的包装函数，如`helper.GoRecover`和`helper.SafeGo`

```go
// 安全地执行可能引发panic的函数
helper.SafeGo(func() {
    // 可能引发panic的代码
})
```

## 6. 接口设计规范

### 6.1 接口定义
- 接口应简洁，专注于单一职责
- 接口名使用er结尾，如`Reader`、`Writer`
- 方法数量适中，避免过度设计

```go
// 示例接口定义
type DB interface {
    Create(r Record) error
    Update(r Record) error
    Delete(id string) error
    Get(id string) (Record, error)
    Iter() func(yield func(Record, error) bool)
}
```

### 6.2 依赖注入
- 使用接口实现依赖注入
- 避免硬编码依赖，提高代码的可测试性

## 7. 测试规范

### 7.1 测试文件
- 测试文件命名为`*_test.go`
- 测试文件与被测试文件放在同一目录下

### 7.2 测试函数
- 测试函数命名为`TestXxx`
- 使用Go标准库`testing`包
- 测试用例应覆盖正常、边界和错误情况

```go
// 测试函数示例
func TestGetJSON(t *testing.T) {
    // 测试逻辑
}
```

### 7.3 表驱动测试
- 推荐使用表驱动测试模式
- 将测试用例组织成表格形式，提高代码复用性和可读性

```go
// 表驱动测试示例
tests := []struct {
    name     string
    input    string
    expected string
}{{
    name:     "test case 1",
    input:    "input1",
    expected: "output1",
}, {
    name:     "test case 2",
    input:    "input2",
    expected: "output2",
}}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // 测试逻辑
    })
}
```

## 8. 并发编程规范

### 8.1 Goroutine管理
- 使用`helper.SafeGo`启动goroutine
- 确保goroutine中的panic被正确捕获
- 避免goroutine泄漏

### 8.2 同步机制
- 使用`sync.Mutex`、`sync.RWMutex`等进行并发控制
- 避免不必要的锁竞争
- 使用`sync.WaitGroup`等待一组goroutine完成

## 9. 依赖管理规范

### 9.1 Go模块
- 使用Go模块（go.mod）管理依赖
- 明确指定依赖版本
- 定期更新依赖，确保安全性和稳定性

### 9.2 第三方库选择
- 选择成熟、维护良好的第三方库
- 优先选择官方或广泛使用的库
- 注意控制依赖数量，避免过度依赖

## 10. Git规范

### 10.1 .gitignore文件
- 包含常见的Go项目忽略项
- 忽略编译产物、IDE配置文件等

```
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
```

### 10.2 代码提交
- 提交信息清晰明了，描述具体变更内容
- 避免提交不相关的代码
- 定期合并和更新代码

## 11. 其他规范

### 11.1 性能优化
- 避免不必要的对象创建
- 合理使用内存缓冲区
- 注意避免常见的性能陷阱

### 11.2 安全性
- 处理用户输入时注意安全检查
- 敏感数据加密存储和传输
- 避免常见的安全漏洞

### 11.3 文档
- 为公共API提供清晰的文档
- 重要功能模块应有使用说明
- 及时更新文档，保持与代码一致