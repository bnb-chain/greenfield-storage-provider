package log

import (
	"io"

	"go.uber.org/multierr"
)

func AddSync(w io.Writer) AsyncWriter {
	switch w := w.(type) {
	case AsyncWriter:
		return w
	default:
		return writerWrapper{w}
	}
}

type multiAsyncWriteSyncer []AsyncWriter

// NewMultiWriteSyncer creates a WriteSyncer that duplicates its writes
// and sync calls, much like io.MultiWriter.
func NewMultiWriteSyncer(ws ...AsyncWriter) AsyncWriter {
	if len(ws) == 1 {
		return ws[0]
	}
	return multiAsyncWriteSyncer(ws)
}

// See https://golang.org/src/io/multi.go
// When not all underlying syncers write the same number of bytes,
// the smallest number is returned even though Write() is called on
// all of them.
func (ws multiAsyncWriteSyncer) Write(p []byte) (int, error) {
	var writeErr error
	nWritten := 0
	for _, w := range ws {
		n, err := w.Write(p)
		writeErr = multierr.Append(writeErr, err)
		if nWritten == 0 && n != 0 {
			nWritten = n
		} else if n < nWritten {
			nWritten = n
		}
	}
	return nWritten, writeErr
}

func (ws multiAsyncWriteSyncer) Sync() error {
	var err error
	for _, w := range ws {
		err = multierr.Append(err, w.Sync())
	}
	return err
}

func (ws multiAsyncWriteSyncer) Stop() error {
	var err error
	for _, w := range ws {
		err = multierr.Append(err, w.Stop())
	}
	return err
}
