package request

import (
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
)

type RequestState int

const SEPARATOR = "\r\n" // CRLF is the separator
var BUFFER_SIZE = 1024

const (
	Initialized RequestState = iota
	Done
)

type Request struct {
	RequestLine RequestLine
	status      RequestState
}

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

var httpVersionRegexMatch = regexp.MustCompile(`^HTTP/\d+(?:\.\d+)?$`)

func RequestFromReader(r io.Reader) (Request, error) {
	buffer := make([]byte, BUFFER_SIZE)
	request := newInitializedRequest()
	readIndex := 0
	for !request.isDone() {
		readByteCount, err := r.Read(buffer[readIndex:])
		log.Print(readByteCount)
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
			copy(buffer, buffer[parsedCount:])
			readIndex = parsedCount
		}
	}

	return *request, nil
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
	log.Print("total read bytes: ", totalReadBytes)
	return totalReadBytes, &RequestLine{method, resource, version}, err
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
	if r.isDone() {
		return 0, errors.New("parser: trying to parse completed data")
	}
	if !r.isValidStatus() {
		return 0, fmt.Errorf("parser: unknown state %v", r.status)
	}

	bytesParsed, requestLine, err := parseRequestLine(string(data))
	if err != nil {
		return 0, err
	}
	if bytesParsed == 0 {
		return 0, nil
	}

	r.RequestLine = *requestLine
	r.status = Done

	return bytesParsed, nil
}
func (r *Request) isDone() bool {
	return r.status == Done
}
func (r *Request) isValidStatus() bool {
	return r.status >= Initialized && r.status <= Done
}
func newInitializedRequest() *Request {
	return &Request{status: Initialized}
}
