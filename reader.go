package cancelio

import "io"

type FdReader interface {
	io.Reader
	Fd() uintptr
}

type CancellableReader interface {
	io.Reader
	Cancel() error
	Close() error
}

func NewCancellableReader(rd FdReader) (CancellableReader, error) {
	return newReader(rd)
}
