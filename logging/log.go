package logging

import (
	"io"
	"log"
)

type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warningf(format string, args ...any)
	Errorf(format string, args ...any)
}

const (
	DebugLevel = iota
	InfoLevel
	WarningLevel
	ErrorLevel
)

// LogAdapter 是一个适配器，适配 Go 标准库的 log 包
type LogAdapter struct {
	logger   *log.Logger
	logLevel int
}

// NewLogAdapter 创建一个 LogAdapter 实例
func NewLogAdapter(out io.Writer, level int) *LogAdapter {
	return &LogAdapter{
		logger:   log.New(out, "", log.LstdFlags),
		logLevel: level,
	}
}

// Debugf 打印 debug 级别日志
func (l *LogAdapter) Debugf(format string, args ...any) {
	l.logf(DebugLevel, "DEBUG", format, args...)
}

// Infof 打印 info 级别日志
func (l *LogAdapter) Infof(format string, args ...any) {
	l.logf(InfoLevel, "INFO", format, args...)
}

// Warningf 打印警告级别日志
func (l *LogAdapter) Warningf(format string, args ...any) {
	l.logf(WarningLevel, "WARN", format, args...)
}

// Errorf 打印错误级别日志
func (l *LogAdapter) Errorf(format string, args ...any) {
	l.logf(ErrorLevel, "ERROR", format, args...)
}

func (l *LogAdapter) SetLogger(logger *log.Logger) {
	l.logger = logger
}

func (l *LogAdapter) SetLogLevel(level int) {
	l.logLevel = level
}

// logf 是一个通用的日志输出方法，封装了日志输出格式和级别检查
func (l *LogAdapter) logf(level int, levelStr string, format string, args ...any) {
	if l.logLevel <= level {
		// 日志前缀：[日志级别] 和格式化的日志内容
		l.logger.Printf("[%s] "+format, append([]any{levelStr}, args...)...)
	}
}
