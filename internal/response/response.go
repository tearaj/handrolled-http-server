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

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusLine := getStatusLine(statusCode)
	_, err := w.Write([]byte(statusLine + constants.SEPARATOR))
	return err
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	_, err := w.Write([]byte(headers.GetAsString()))
	return err
}

func WriteBody(w io.Writer, data []byte) error {
	_, err := w.Write(data)
	return err
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
