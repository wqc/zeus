package log

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type LogCfg struct {
	Level      string `yaml:"log_level"`
	Path       string `yaml:"log_path"`
	RotateSize int    `yaml:"log_rotate_size"`
}

type logFileWriter struct {
	dir      string
	filename string
	path     string
	maxsize  int64
	size     int64
	file     *os.File
	mutex    *sync.Mutex
}

func newLogFileWriter(path string, maxsize int64) *logFileWriter {
	w := new(logFileWriter)
	w.dir, w.filename = filepath.Split(path)
	if w.dir == "" {
		w.dir, _ = os.Getwd()
		if w.dir == "" {
			w.dir = "."
		}
	}

	w.maxsize = maxsize
	w.path = path
	w.mutex = new(sync.Mutex)
	return w
}

func (w *logFileWriter) Open() error {
	f, err := os.OpenFile(w.path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}

	w.size = fi.Size()
	w.file = f

	return nil
}

func (w *logFileWriter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.file != nil {
		return w.file.Close()
	}

	return nil
}

func (w *logFileWriter) Rotate() error {
	if w.size < w.maxsize {
		return nil
	}

	if err := w.file.Close(); err != nil {
		return err
	}

	now := time.Now()
	newpath := fmt.Sprintf("%s_%04d%02d%02d%02d%02d%2d",
		w.path, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
	if err := os.Rename(w.path, newpath); err != nil {
		return err
	}

	w.file = nil
	w.size = 0

	return nil
}

func (w *logFileWriter) Write(data []byte) (int, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.file == nil {
		if err := w.Open(); err != nil {
			return -1, err
		}
	}

	n, err := w.file.Write(data)
	if err != nil {
		return -1, err
	}

	w.size += int64(n)
	if err = w.Rotate(); err != nil {
		return -1, err
	}

	return n, nil
}

type zeusLogFmter struct{}

func (f *zeusLogFmter) Format(entry *logrus.Entry) ([]byte, error) {
	_, file, line, ok := runtime.Caller(7)
	if ok {
		_, file = filepath.Split(file)
	}

	b := entry.Buffer
	if b == nil {
		b = &bytes.Buffer{}
	}

	b.WriteString(entry.Time.Format("2006-01-02 15:04:05 "))
	b.WriteString(f.Level(entry.Level))
	b.WriteByte(' ')

	if ok {
		b.WriteString(file)
		b.WriteByte(':')
		b.WriteString(strconv.Itoa(line))
		b.WriteByte(' ')
	}

	b.WriteString(entry.Message)
	b.WriteByte('\n')

	return b.Bytes(), nil
}

func (f *zeusLogFmter) Level(level logrus.Level) string {
	switch level {
	case logrus.DebugLevel:
		return "[Debug]"

	case logrus.InfoLevel:
		return "[Info ]"

	case logrus.WarnLevel:
		return "[Warn ]"

	case logrus.ErrorLevel:
		return "[Error]"

	case logrus.FatalLevel:
		return "[Fatal]"

	}

	return "[Info ]"
}

func InitConsoleLog() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(new(zeusLogFmter))
	initCosoleLog()
}

func Initlog(path, level string, max int) {
	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.SetFormatter(new(zeusLogFmter))

	if path == "" || path == "console" {
		initCosoleLog()
	} else {
		initFileLog(path, int64(max))
	}
}

func initCosoleLog() {
	logrus.SetOutput(os.Stdout)
}

func initFileLog(path string, maxsize int64) {
	w := newLogFileWriter(path, maxsize)
	logrus.SetOutput(w)
}

func Debug(format string, args ...interface{}) {
	logrus.Debugf(format, args...)
}

func Info(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

func Warn(format string, args ...interface{}) {
	logrus.Warnf(format, args)
}

func Error(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func Fatal(format string, args ...interface{}) {
	logrus.Fatalf(format, args...)
}

func Close() {
	logrus.Exit(0)
}
