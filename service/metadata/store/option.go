package store

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// when enable = true, sql will be printed
type optionLogger struct {
	Enable bool
}

func (ol *optionLogger) Apply(cfg *gorm.Config) error {
	if cfg.Logger == nil {
		lcfg := logger.Config{
			SlowThreshold:             0, /* 0 will disable print slow sql*/
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: false,
			Colorful:                  false,
		}

		if ol.Enable {
			lcfg.SlowThreshold = 200 * time.Millisecond
			lcfg.LogLevel = logger.Info
		}

		cfg.Logger = logger.New(newLogWriter(), lcfg)
	}
	return nil
}

func (ol *optionLogger) AfterInitialize(*gorm.DB) error {
	return nil
}

func newLogWriter() *logWriter {
	return &logWriter{}
}

// Writer log writer interface
type logWriter struct {
}

func (lw *logWriter) Printf(fmt string, args ...interface{}) {
	print("dummy logger for now")
}
