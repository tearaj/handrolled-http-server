package response

import (
	"fmt"
	"httpFromTCP/internal/constants"
	"httpFromTCP/internal/headers"
	"io"
)

type StatusCode int

const (
	STATUS_CODE_OK                    StatusCode = 200
	STATUS_CODE_BAD_REQUEST           StatusCode = 400
	STATUS_CODE_INTERNAL_SERVER_ERROR StatusCode = 500
)

type WriterState = string

const (
	StateInitialized     WriterState = "INITIALIZED"
	StateStatusLine      WriterState = "STATUS_LINE"
	StateHeaders         WriterState = "STATE_HEADERS"
	StateBody            WriterState = "STATE_BODY"
	StateChunkedBody     WriterState = "STATE_CHUNKED_BODY"
	StateChunkedBodyDone WriterState = "STATE_CHUNKED_BODY_DONE"
	StateError           WriterState = "STATE_ERROR"
)

type Writer struct {
	w           io.Writer
	writerState WriterState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w, StateInitialized}
}
func (w *Writer) Write(data []byte) (int, error) {
	return w.w.Write(data)
}
func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != StateInitialized {
		return fmt.Errorf("response: writing status line while not initialized")
	}
	statusLine := getStatusLine(statusCode)
	_, err := w.Write([]byte(statusLine + constants.SEPARATOR))
	w.writerState = StateStatusLine
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != StateStatusLine {
		return fmt.Errorf("response: writing headers without writing status line")
	}
	_, err := w.Write([]byte(headers.GetAsString()))
	v, ok := headers.Get("transfer-encoding")
	if !ok || v != "chunked" {
		w.writerState = StateHeaders
	}
	w.writerState = StateChunkedBody
	return err
}

func (w *Writer) WriteBody(data []byte) error {
	if w.writerState != StateHeaders {
		return fmt.Errorf("response: writing body without writing headers")
	}
	_, err := w.Write(data)
	w.writerState = StateBody
	return err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.writerState != StateChunkedBody {
		return 0, fmt.Errorf("response: writing chunked body while in wrong state")
	}
	count := len(p)
	lengthLine := fmt.Sprintf("%x%s", count, constants.SEPARATOR)
	_, err := w.Write([]byte(lengthLine))
	if err != nil {
		return 0, err
	}
	dataLine := append(p, []byte(constants.SEPARATOR)...)
	n, err := w.Write(dataLine)
	return n, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.writerState != StateChunkedBody {
		return 0, fmt.Errorf("response: writing chunked body done while in wrong state")
	}
	closingString := fmt.Sprintf("0%s%s", constants.SEPARATOR, constants.SEPARATOR)
	n, err := w.Write([]byte(closingString))
	if err != nil {
		w.writerState = StateError
		return 0, err
	}
	w.writerState = StateChunkedBodyDone
	return n, err
}

func getStatusLine(statusCode StatusCode) string {
	switch statusCode {
	case STATUS_CODE_OK:
		return fmt.Sprintf("HTTP/1.1 %v OK", STATUS_CODE_OK)
	case STATUS_CODE_BAD_REQUEST:
		return fmt.Sprintf("HTTP/1.1 %v Bad Request", STATUS_CODE_BAD_REQUEST)
	case STATUS_CODE_INTERNAL_SERVER_ERROR:
		return fmt.Sprintf("HTTP/1.1 %v Internal Server Error", STATUS_CODE_INTERNAL_SERVER_ERROR)
	}
	return fmt.Sprintf("HTTP/1.1 %v", statusCode)
}
