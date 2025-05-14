package custommiddleware

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/kataras/golog"
	"gitlab.com/ayaka/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	once     sync.Once
	instance *lumberjack.Logger
)

type LogActivityHandler struct {
	Conf *config.Config `inject:"config"`
}

type LogConfig struct {
	LogDir     string
	MaxSize    int // MB
	MaxBackups int // jumlah file
	MaxAge     int // day
	Compress   bool
}

// LogEntry represents the structure of a log entry in JSON format
type LogEntry struct {
	LogLevel    string `json:"logLevel"`
	Timestamp   string `json:"timestamp"`
	User        string `json:"user"`
	Message     string `json:"message"`
	App         string `json:"app"`
	AppVersion  string `json:"appVer"`
	Environment string `json:"env"`
}

// getDefaultConfig returns default logging configuration
func (l *LogActivityHandler) getDefaultConfig() LogConfig {
	return LogConfig{
		LogDir:     l.Conf.Log.FileLocation,
		MaxSize:    l.Conf.Log.FileMaxSize,
		MaxBackups: l.Conf.Log.FileMaxBackup,
		MaxAge:     l.Conf.Log.FileMaxAge,
		Compress:   true,
	}
}

// initLogger menginisialisasi logger dengan singleton pattern
func (l *LogActivityHandler) initLogger() error {
	var err error
	once.Do(func() {
		config := l.getDefaultConfig()

		// Buat direktori log jika belum ada
		if err = os.MkdirAll(config.LogDir, os.ModePerm); err != nil {
			return
		}

		logFile := filepath.Join(config.LogDir, "activity.log")

		// Konfigurasi log rotation
		instance = &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}

		// Set output & level
		golog.SetOutput(instance)
		golog.SetLevel("debug")

		golog.SetTimeFormat("")
		golog.SetPrefix("")

		golog.Handle(func(entry *golog.Log) bool {
			fmt.Fprintln(instance, entry.Message)
			return true
		})
	})

	return err
}

// LogUserInfo mencatat aktivitas pengguna dalam format JSON
func (l *LogActivityHandler) LogUserInfo(user interface{}, loglevel, activity string) error {
	if err := l.initLogger(); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	logEntry := LogEntry{
		LogLevel:    loglevel,
		Timestamp:   time.Now().Format(time.RFC3339),
		User:        user.(string),
		Message:     activity,
		App:         l.Conf.App,
		AppVersion:  l.Conf.AppVer,
		Environment: l.Conf.Env,
	}

	jsonLog, err := formatLogMessageJSON(logEntry)
	if err != nil {
		return fmt.Errorf("failed to format log message: %w", err)
	}

	golog.Info(jsonLog)
	return nil
}

// formatLogMessageJSON converts log entry to JSON string
func formatLogMessageJSON(entry LogEntry) (string, error) {
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// FlushLogs memaksa menulis log yang tersisa dan menutup file
func FlushLogs() error {
	if instance != nil {
		return instance.Close()
	}
	return nil
}

// GetLogFilePath mengembalikan path file log aktif
func GetLogFilePath() string {
	if instance != nil {
		return instance.Filename
	}
	return ""
}

// Helper function untuk membaca log dalam format JSON
func ReadLogEntry(jsonStr string) (*LogEntry, error) {
	var entry LogEntry
	err := json.Unmarshal([]byte(jsonStr), &entry)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}
