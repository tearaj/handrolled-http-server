package request

import (
	"errors"
	"fmt"
	"httpFromTCP/internal/headers"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type RequestState string

const SEPARATOR = "\r\n" // CRLF is the separator
const INITIAL_BUFFER_SIZE = 1
const MAX_BUFFER_SIZE = 1024
const (
	Initialized      RequestState = "INITIALIZED"
	StateRequestLine RequestState = "STATE_REQUEST_LINE"
	StateHeaders     RequestState = "STATE_HEADERS"
	StateHeadersDone RequestState = " STATE_HEADERS_DONE"
	StateBody        RequestState = "STATE_BODY"
	StateBodyDone    RequestState = "STATE_BODY_DONE"
	Done             RequestState = "DONE"
)

var httpVersionRegexMatch = regexp.MustCompile(`^HTTP/\d+(?:\.\d+)?$`)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	status      RequestState
}

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

func (r *Request) isDone() bool {
	return r.status == Done
}

func (r *Request) isValidContentLength() error {
	if !r.isDone() {
		return fmt.Errorf("body: content length check done before body parsing completion")
	}
	contentLengthStr, ok := r.Headers.Get(headers.CONTENT_LENGTH)
	if !ok {
		return nil
	}
	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		return fmt.Errorf("body: %w", err)
	}
	if contentLength != len(r.Body) {
		return fmt.Errorf("body: content-length reported not matching actual")
	}
	return nil
}

func (r *Request) getContentLength() (int, error) {
	contentLengthStr, ok := r.Headers.Get(headers.CONTENT_LENGTH)
	if !ok {
		return 0, nil
	}
	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		return 0, err
	}
	return contentLength, nil
}
func newInitializedRequest() *Request {
	return &Request{status: Initialized}
}

func RequestFromReader(r io.Reader) (Request, error) {
	bufferSize := INITIAL_BUFFER_SIZE
	reader := r

	buffer := make([]byte, bufferSize)
	request := newInitializedRequest()
	readIndex := 0
	for !request.isDone() {
		if readIndex == bufferSize {
			increasedBuffer, increasedBufferSize, err := increaseBufferSize(buffer, MAX_BUFFER_SIZE)
			buffer = increasedBuffer
			bufferSize = increasedBufferSize
			if err != nil {
				return Request{}, err
			}
		}
		readByteCount, err := reader.Read(buffer[readIndex:])
		if err != nil {
			if err == io.EOF {
				request.status = Done
				break
			}
			return Request{}, err
		}
		readIndex += readByteCount
		parsedCount, err := request.parse(buffer[:readIndex])
		if err != nil {
			return Request{}, err
		}
		if parsedCount != 0 {
			remainingUnparsedBytes := readIndex - parsedCount
			copy(buffer, buffer[parsedCount:readIndex])
			readIndex = remainingUnparsedBytes
		}
	}

	err := request.isValidContentLength()

	return *request, err
}

func extractVersion(versionStr string) (string, error) {
	matchesFormat := httpVersionRegexMatch.MatchString(versionStr)
	if !matchesFormat {
		return "", errors.New("http version: does not match expected pattern of")
	}
	parts := strings.Split(versionStr, "/")
	versionNumberStr := parts[1]
	return versionNumberStr, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	remainingData := data

	for {
		switch r.status {
		case Done:
			return totalBytesParsed, nil
		case Initialized, StateRequestLine:
			n, err := initializedStateMethod(r, remainingData)
			if err != nil {
				return totalBytesParsed, err
			}
			if n == 0 {
				return totalBytesParsed, nil
			}
			totalBytesParsed += n
			remainingData = remainingData[n:]

		case StateHeaders:
			n, err := stateHeadersMethod(r, remainingData)
			if err != nil {
				return totalBytesParsed, err
			}
			if n == 0 {
				return totalBytesParsed, nil
			}
			totalBytesParsed += n
			remainingData = remainingData[n:]

		case StateBody:
			n, err := stateBodyMethod(r, remainingData)
			if err != nil {
				return totalBytesParsed, err
			}
			totalBytesParsed += n
			return totalBytesParsed, nil

		default:
			return totalBytesParsed, fmt.Errorf("unknown state when parsing request")
		}

		if len(remainingData) == 0 {
			return totalBytesParsed, nil
		}
	}
}

func initializedStateMethod(r *Request, data []byte) (int, error) {
	bytesParsed, requestLine, err := parseRequestLine(string(data))
	if err != nil {
		return 0, err
	}
	if bytesParsed == 0 {
		return 0, nil
	}

	r.RequestLine = *requestLine
	r.Headers = headers.NewHeaders()
	r.status = StateHeaders
	return bytesParsed, nil

}

func stateHeadersMethod(r *Request, data []byte) (int, error) {
	totalBytesParsed := 0
	remainingData := data

	for {
		n, done, err := r.Headers.Parse(remainingData)
		if err != nil {
			return totalBytesParsed, err
		}

		if n == 0 {
			return totalBytesParsed, nil
		}

		totalBytesParsed += n
		remainingData = remainingData[n:]

		if done {
			_, ok := r.Headers.Get(headers.CONTENT_LENGTH)
			if ok {
				r.status = StateBody
			} else {
				r.status = Done
			}
			return totalBytesParsed, nil
		}

		if len(remainingData) == 0 {
			return totalBytesParsed, nil
		}
	}
}

func stateBodyMethod(r *Request, data []byte) (int, error) {
	contentLength, err := r.getContentLength()
	if err != nil {
		return 0, err
	}
	bytesNeeded := contentLength - len(r.Body)
	if bytesNeeded <= 0 {
		r.status = Done
		return 0, nil
	}
	parsableByteCount := min(bytesNeeded, len(data))
	r.Body = append(r.Body, data[:parsableByteCount]...)
	if len(r.Body) == contentLength {
		r.status = Done
	}
	return len(data), err
}

func increaseBufferSize(currentBuffer []byte, maxBufferSize int) ([]byte, int, error) {
	if len(currentBuffer) == maxBufferSize {
		return currentBuffer, maxBufferSize, fmt.Errorf("maximum buffer size of %v reached", maxBufferSize)
	}
	tmpBuffer := make([]byte, 2*len(currentBuffer))
	copy(tmpBuffer, currentBuffer)

	return tmpBuffer, 2 * len(currentBuffer), nil
}

func parseRequestLine(requestStr string) (int, *RequestLine, error) {
	separatorIndex := strings.Index(requestStr, SEPARATOR)
	if separatorIndex == -1 {
		return 0, nil, nil
	}
	requestLineStr := requestStr[:separatorIndex]
	parts := strings.Split(requestLineStr, " ")
	if len(parts) != 3 {
		return 0, &RequestLine{}, errors.New("number of parts in request line must be 3")
	}
	method := parts[0]
	resource := parts[1]
	version, err := extractVersion(parts[2])
	totalReadBytes := len(requestLineStr) + len(SEPARATOR)
	return totalReadBytes, &RequestLine{method, resource, version}, err
}
