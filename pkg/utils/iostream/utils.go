package iostream

import (
	"os"
	"strings"

	iostreamTY "github.com/jkandasa/autoeasy/pkg/types/iostream"
	"go.uber.org/zap"
)

type ioStreamsLogger struct {
	isErrorWriter bool
}

func (io *ioStreamsLogger) Write(data []byte) (int, error) {
	msg := strings.TrimSuffix(string(data), "\n")
	if io.isErrorWriter {
		zap.L().Error(msg)
	} else {
		zap.L().Info(msg)
	}

	return len(data), nil
}

func GetLogWriter() *iostreamTY.IOStreams {
	return &iostreamTY.IOStreams{
		In:     os.Stdin,
		Out:    &ioStreamsLogger{isErrorWriter: false},
		ErrOut: &ioStreamsLogger{isErrorWriter: true},
	}
}
