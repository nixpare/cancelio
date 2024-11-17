//go:build unix
package cancelio

import (
	"errors"
	"fmt"
	"io"
	"time"

	"golang.org/x/sys/unix"
)

type reader struct {
	rd FdReader
	fd uintptr
	flags int
	cancelCh chan struct{}
}

var ErrCanceled = errors.New("read operation was canceled")

func newReader(rd FdReader) (CancellableReader, error) {
	r := &reader{
		rd: rd,
		fd: rd.Fd(),
		cancelCh: make(chan struct{}),
	}

	var err error
	r.flags, err = unix.FcntlInt(r.fd, unix.F_GETFL, 0)
	if err != nil {
		return nil, err
	}

	_, err = unix.FcntlInt(r.fd, unix.F_SETFL, r.flags | unix.O_NONBLOCK)
	if err != nil {
		return nil, err
	}

	return r, nil
}

var PollIntervalMilli = 1

func (r *reader) Read(p []byte) (n int, err error) {
	for {
		select {
		case <- r.cancelCh:
			err = fmt.Errorf("%w: %w", io.EOF, ErrCanceled)
			return
		default:
			n, err = r.rd.Read(p)
			if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EWOULDBLOCK) {
				time.Sleep(time.Millisecond * time.Duration(PollIntervalMilli))
				continue
			}

			return
		}
	}
}

func (r *reader) Cancel() error {
	r.cancelCh <- struct{}{}
	return nil
}

func (r *reader) Close() error {
	close(r.cancelCh)
	_, err := unix.FcntlInt(r.fd, unix.F_SETFL, r.flags)
	return err
}
