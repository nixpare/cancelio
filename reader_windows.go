package cancelio

import (
	"errors"
	"fmt"
	"io"

	"golang.org/x/sys/windows"
)

type reader struct {
	rd         FdReader
	handle     windows.Handle
	overlapped windows.Overlapped
}

func newReader(rd FdReader) (CancellableReader, error) {
	r := &reader{
		rd:     rd,
		handle: windows.Handle(rd.Fd()),
	}

	var err error
	r.overlapped.HEvent, err = windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *reader) Read(p []byte) (n int, err error) {
	var bytesRead uint32
	err = windows.ReadFile(r.handle, p, &bytesRead, &r.overlapped)
	if err != nil {
		if errors.Is(err, windows.ERROR_OPERATION_ABORTED) {
			err = fmt.Errorf("%w: %w", io.EOF, err)
		}
		return
	}

	n = int(bytesRead)
	return
}

func (r *reader) Cancel() error {
	err := windows.CancelIoEx(r.handle, &r.overlapped)
	if err != nil && !errors.Is(err, windows.ERROR_NOT_FOUND) {
		return err
	}

	return nil
}

func (r *reader) Close() error {
	return windows.CloseHandle(r.overlapped.HEvent)
}
