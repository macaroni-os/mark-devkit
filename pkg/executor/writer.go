/*
Copyright © 2024 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package executor

import (
	log "github.com/macaroni-os/mark-devkit/pkg/logger"
)

type ExecutorWriter struct {
	Type string
}

func NewExecutorWriter(t string) *ExecutorWriter {
	return &ExecutorWriter{Type: t}
}

func (w *ExecutorWriter) Write(p []byte) (int, error) {
	logger := log.GetDefaultLogger()
	switch w.Type {
	case "stderr":
		logger.Msg("info", false, false,
			logger.Aurora.Bold(
				logger.Aurora.BrightRed(string(p)),
			),
		)
	case "stdout":
		logger.Msg("info", false, false,
			string(p),
		)
	}

	return len(p), nil
}

func (w *ExecutorWriter) Close() error {
	return nil
}
