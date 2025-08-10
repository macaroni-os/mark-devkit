/*
Copyright © 2024-2025 Macaroni OS Linux
See AUTHORS and LICENSE for the license details and contributors.
*/
package executor

import (
	log "github.com/macaroni-os/mark-devkit/pkg/logger"
)

type ExecutorWriter struct {
	Type  string
	Quiet bool
}

func NewExecutorWriter(t string, q bool) *ExecutorWriter {
	return &ExecutorWriter{
		Type:  t,
		Quiet: q,
	}
}

func (w *ExecutorWriter) Write(p []byte) (int, error) {
	logger := log.GetDefaultLogger()
	if !w.Quiet {
		switch w.Type {
		case "stderr":
			logger.Msg("info", true, false, string(p))
		case "stdout":
			logger.Msg("info", true, false, string(p))
		}
	}

	return len(p), nil
}

func (w *ExecutorWriter) Close() error {
	return nil
}
