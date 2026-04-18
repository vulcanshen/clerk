package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/vulcanshen/clerk/internal/config"
)

var cleanOnce sync.Once

func logDir(cfg config.Config) string {
	return filepath.Join(config.ExpandPath(cfg.Output.Dir), "log")
}

func logPath(cfg config.Config) string {
	name := time.Now().Format("20060102") + "-clerk.log"
	return filepath.Join(logDir(cfg), name)
}

func write(cfg config.Config, level, msg string) {
	path := logPath(cfg)
	os.MkdirAll(filepath.Dir(path), 0755)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	ts := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(f, "[%s] [%s] %s\n", ts, level, msg)

	cleanOnce.Do(func() { cleanOldLogs(cfg) })
}

func cleanOldLogs(cfg config.Config) {
	dir := logDir(cfg)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -cfg.Log.RetentionDays)

	for _, e := range entries {
		name := e.Name()
		// parse date from filename: YYYYMMDD-clerk.log
		if len(name) < 8 {
			continue
		}
		t, err := time.Parse("20060102", name[:8])
		if err != nil {
			continue
		}
		if t.Before(cutoff) {
			os.Remove(filepath.Join(dir, name))
		}
	}
}

func Info(cfg config.Config, msg string) {
	write(cfg, "INFO", msg)
}

func Infof(cfg config.Config, format string, args ...interface{}) {
	write(cfg, "INFO", fmt.Sprintf(format, args...))
}

func Error(cfg config.Config, msg string) {
	write(cfg, "ERROR", msg)
}

func Errorf(cfg config.Config, format string, args ...interface{}) {
	write(cfg, "ERROR", fmt.Sprintf(format, args...))
}

func LogPath(cfg config.Config) string {
	return logPath(cfg)
}
