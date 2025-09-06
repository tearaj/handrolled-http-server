package request

import (
	"errors"
	"io"
	"log"
	"regexp"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	Method        string
	RequestTarget string
	HttpVersion   string
}

var httpVersionRegexMatch = regexp.MustCompile(`^HTTP/\d+(?:\.\d+)?$`)

func RequestFromReader(r io.Reader) (Request, error) {
	readBytes, err := io.ReadAll(r)
	if err != nil {
		log.Print("Error while reading: ", err)
		return Request{}, err
	}
	readString := string(readBytes)
	split := strings.Split(readString, "\r\n")
	requestLineStr := split[0]
	requestLine, err := parseRequestLine(requestLineStr)
	if err != nil {
		return Request{}, err
	}
	return Request{requestLine}, nil
}

func parseRequestLine(requestLineStr string) (RequestLine, error) {
	parts := strings.Split(requestLineStr, " ")
	if len(parts) != 3 {
		return RequestLine{}, errors.New("number of parts in request line must be 3")
	}
	method := parts[0]
	resource := parts[1]
	version, err := extractVersion(parts[2])
	return RequestLine{method, resource, version}, err
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
