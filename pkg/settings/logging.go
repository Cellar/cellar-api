package settings

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	loggingKey          = "logging."
	loggingLevelKey     = loggingKey + "level"
	loggingEnableStdOut = loggingKey + "enable_stdout"
	loggingDirectoryKey = loggingKey + "directory"
)

type ILoggingConfiguration interface {
	Locations() ([]io.Writer, error)
	Level() (log.Level, error)
}

type LoggingConfiguration struct{}

func NewLoggingConfiguration() *LoggingConfiguration {
	viper.SetDefault(loggingLevelKey, log.InfoLevel)
	viper.SetDefault(loggingDirectoryKey, "")
	viper.SetDefault(loggingEnableStdOut, true)
	return &LoggingConfiguration{}
}

func (lgc *LoggingConfiguration) Locations() (locations []io.Writer, err error) {
	logDirectory := viper.GetString(loggingDirectoryKey)
	if logDirectory != "" && filepath.IsAbs(logDirectory) {
		if writer, err := openLogFile(logDirectory); err == nil {
			locations = append(locations, writer)
		} else {
			log.WithError(err).
				Errorf("Unable to open log file in directory '%s'", logDirectory)
		}
	}

	if viper.GetBool(loggingEnableStdOut) {
		locations = append(locations, os.Stdout)
	}
	return
}

func openLogFile(directory string) (io.Writer, error) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err = os.MkdirAll(directory, 644)
		if err != nil {
			return nil, err
		}
	}
	datestamp := time.Now().Format("2006-01-02")
	fileName := fmt.Sprintf("cellar-%s.log", datestamp)
	filePath := filepath.Join(directory, fileName)
	return os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0755)
}

func (lgc LoggingConfiguration) Level() (level log.Level, err error) {
	levelStr := []byte(viper.GetString(loggingLevelKey))
	err = level.UnmarshalText(levelStr)

	return
}
