// Package logging configura a saida padrao de logs da aplicacao.
package logging

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	FilePath   string
	MaxSizeMB  int
	MaxAgeDays int
	MaxBackups int
	Compress   bool
	Stdout     bool
}

func Configure(cfg Config) (io.Closer, error) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC | log.Lshortfile)

	filePath := strings.TrimSpace(cfg.FilePath)
	if filePath == "" {
		log.SetOutput(os.Stdout)
		return nopCloser{}, nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0o750); err != nil {
		return nil, err
	}

	rotatingLogger := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    cfg.MaxSizeMB,
		MaxAge:     cfg.MaxAgeDays,
		MaxBackups: cfg.MaxBackups,
		Compress:   cfg.Compress,
	}

	var output io.Writer = rotatingLogger
	if cfg.Stdout {
		output = io.MultiWriter(os.Stdout, rotatingLogger)
	}
	log.SetOutput(output)

	return rotatingLogger, nil
}

type nopCloser struct{}

func (nopCloser) Close() error { return nil }
