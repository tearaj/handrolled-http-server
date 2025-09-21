package headers

import (
	"bytes"
	"fmt"
	"httpFromTCP/internal/constants"
	"maps"
	"strings"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (int, bool, error) {
	// n := 0
	separatorIndex := bytes.Index(data, []byte(constants.SEPARATOR))
	if separatorIndex == 0 {
		return len(constants.SEPARATOR), true, nil
	}
	if separatorIndex == -1 {
		return 0, false, nil
	}
	header := string(data[:separatorIndex])

	headers, err := validateHeader(header)

	if err != nil {
		return 0, false, err
	}
	maps.Copy(h, headers.normalizeHeaderKeys())
	return separatorIndex + len(constants.SEPARATOR), false, nil
}

func validateHeader(header string) (Headers, error) {
	headerParts := strings.SplitN(header, ":", 2)
	headers := make(Headers)
	if len(headerParts) != 2 {
		return Headers{}, fmt.Errorf("header: header has too many parts")
	}
	headerName := headerParts[0]
	headerValue := headerParts[1]
	hasInvalidSpaces := strings.HasSuffix(headerName, " ")
	if hasInvalidSpaces {
		return Headers{}, fmt.Errorf("header: header name or valud has spaces in invalid locations")
	}
	trimmedHeaderName := strings.TrimSpace(headerName)
	if strings.Contains(trimmedHeaderName, " ") {
		return Headers{}, fmt.Errorf("header: header name contains spaces")
	}
	headers[trimmedHeaderName] = strings.TrimSpace(headerValue)
	return headers, nil
}

func (h Headers) normalizeHeaderKeys() Headers {
	normalizedHeaders := make(Headers, len(h))
	for k := range maps.Keys(h) {
		normalizedHeaders[strings.ToLower(k)] = h[k]
	}
	return normalizedHeaders
}

func NewHeaders() Headers {
	return make(map[string]string)
}
