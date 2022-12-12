package iostream

import (
	"os"

	iostreamTY "github.com/jkandasa/autoeasy/pkg/types/iostream"
	"go.uber.org/zap"
)

type ioStreamsLogger struct {
	isErrorWriter bool
}

func (io *ioStreamsLogger) Write(data []byte) (int, error) {
	if io.isErrorWriter {
		zap.L().Error(string(data))
	} else {
		zap.L().Info(string(data))
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
