package mocks

import (
	"fmt"
	"inventory-movement-processing/pkg/logger"
	"testing"
)

type Logger struct {
	t *testing.T // Giữ reference tới testing.T để đánh dấu fail khi có log Error
}

// Hàm khởi tạo giúp truyền nhanh testing.T vào logger
func NewMockLogger(t *testing.T) *Logger {
	return &Logger{t: t}
}

func (m *Logger) Debug(args ...interface{}) {}
func (m *Logger) Info(args ...interface{})  {}
func (m *Logger) Warn(args ...interface{})  {}

func (m *Logger) Debugf(format string, args ...interface{}) {}
func (m *Logger) Infof(format string, args ...interface{})  {}
func (m *Logger) Warnf(format string, args ...interface{})  {}

// Khi code chạy vào log Error, logger sẽ in ra dạng log tùy biến và làm tốn (Fail) test case
func (m *Logger) Error(args ...interface{}) {
	if m.t != nil {
		m.t.Helper() // Giúp báo đúng dòng code bị lỗi trong file test
		m.t.Error(args...)
	} else {
		fmt.Println(args...)
	}
}

func (m *Logger) Errorf(format string, args ...interface{}) {
	if m.t != nil {
		m.t.Helper()
		m.t.Errorf(format, args...)
	} else {
		fmt.Printf(format+"\n", args...)
	}
}

func (m *Logger) With(key string, value interface{}) logger.Logger {
	return m
}

func (m *Logger) WithFields(fields logger.Fields) logger.Logger {
	return m
}