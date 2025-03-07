package logger

import (
	"fmt"
	"io"
	"math/bits"
	"os"
	"strings"
	"time"
)

const (
	LevelDebug int8 = iota
	LevelWarn
	LevelError
	LevelFatal
)

type Logger struct {
	file  *os.File
	Out   io.Writer
	level int8
}

func NewLogger(filepath string, level string) (*Logger, error) {
	var logger Logger

	switch strings.ToLower(level) {
	case "fatal":
		logger.level = LevelFatal
	case "error":
		logger.level = LevelError
	case "warn":
		logger.level = LevelWarn
	default:
		logger.level = LevelDebug

	}

	logger.Out = os.Stdout

	if filepath != "" {
		file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
		if err != nil {
			return nil, err
		}
		logger.Out = io.MultiWriter(file, os.Stdout)
	}

	return &logger, nil
}

func (l *Logger) Fatal(args ...any) {
	args = append([]any{l.Prefix(), "[FATAL]"}, args...)
	if l.level <= LevelFatal {
		fmt.Fprintln(l.Out, args...)
	}
	os.Exit(1)
}

func (l *Logger) Error(args ...any) {
	args = append([]any{l.Prefix(), "[ERROR]"}, args...)
	if l.level <= LevelError {
		fmt.Fprintln(l.Out, args...)
	}
}

func (l *Logger) Warn(args ...any) {
	args = append([]any{l.Prefix(), "[WARN]"}, args...)
	if l.level <= LevelWarn {
		fmt.Fprintln(l.Out, args...)
	}
}

func (l *Logger) Debug(args ...any) {
	args = append([]any{l.Prefix(), "[DEBUG]"}, args...)
	if l.level <= LevelDebug {
		fmt.Fprintln(l.Out, args...)
	}
}

const httpLogFormat = " │%-7s│%-25s│%-5d│%-16s│%-12s│%-12s│%-12s│\n"

func (l *Logger) HttpLog(method string, url string, code int, remoteAddr string, read int, write int, duration time.Duration) {
	readS := formatBytes(uint64(read))
	writeS := formatBytes(uint64(write))
	durationS := formatDuration(duration)

	if code < 400 {
		l.Debug(fmt.Sprintf(httpLogFormat, method, url, code, remoteAddr, readS, writeS, durationS))
	} else if code < 500 {
		l.Warn(fmt.Sprintf(httpLogFormat, method, url, code, remoteAddr, readS, writeS, durationS))
	} else {
		l.Error(fmt.Sprintf(httpLogFormat, method, url, code, remoteAddr, readS, writeS, durationS))
	}
}

func (l *Logger) Prefix() string {
	return time.Now().Format(time.DateTime)
}

func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func formatBytes(bytes uint64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}

	base := uint(bits.Len64(bytes) / 10)
	val := float64(bytes) / float64(uint64(1<<(base*10)))

	return fmt.Sprintf("%.1f %ciB", val, " KMGTPE"[base])
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Millisecond)
	ms := d.Milliseconds()
	s := ms / 1000
	if s == 0 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%ds%dms", s, ms-(s*1000))
}
