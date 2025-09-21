package request

import (
	"errors"
	"fmt"
	"httpFromTCP/internal/headers"
	"io"
	"regexp"
	"strings"
)

type RequestState int

const SEPARATOR = "\r\n" // CRLF is the separator
const INITIAL_BUFFER_SIZE = 1
const MAX_BUFFER_SIZE = 1024
const (
	Initialized RequestState = iota
	StateRequestLine
	StateHeaders
	StateHeadersDone
	Done
)

var httpVersionRegexMatch = regexp.MustCompile(`^HTTP/\d+(?:\.\d+)?$`)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
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

func newInitializedRequest() *Request {
	return &Request{status: Initialized}
}

func RequestFromReader(r io.Reader) (Request, error) {
	bufferSize := INITIAL_BUFFER_SIZE

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
		readByteCount, err := r.Read(buffer[readIndex:])
		if err != nil {
			if err == io.EOF {
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

	return *request, nil
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
	switch r.status {
	case Done:
		return 0, errors.New("parser: trying to parse completed data")
	case Initialized, StateRequestLine:
		return initializedStateMethod(r, data)
	case StateHeaders:
		return stateHeadersMethod(r, data)
	}
	return 0, fmt.Errorf("Unknown state")
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
	n, done, err := r.Headers.Parse(data)
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, nil
	}
	if done {
		r.status = Done
	}
	return n, nil
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
