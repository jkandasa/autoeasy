package iostream

import "io"

type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}
