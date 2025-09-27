package headers

import (
	"bytes"
	"fmt"
	"httpFromTCP/internal/constants"
	"httpFromTCP/internal/utils"
	"maps"
	"strconv"

	"regexp"
	"strings"
)

type Headers map[string]string

const CONTENT_LENGTH = "content-length"

var validHeaderNamesRegex = regexp.MustCompile("^[A-Za-z0-9!#$%&'*+\\-.^_`|~]+$")

func NewHeaders() Headers {
	return make(map[string]string)
}

func (h Headers) Parse(data []byte) (int, bool, error) {
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
	h.mergeHeaders(headers.normalizeHeaderKeys())
	return separatorIndex + len(constants.SEPARATOR), false, nil
}

func (h Headers) Get(key string) (string, bool) {
	val, ok := h[strings.ToLower(key)]
	return val, ok
}

func (h Headers) Set(headerName string, headerValue string) {
	h[strings.ToLower(headerName)] = headerValue
}

func GetDefaultHeaders(contentSize int) Headers {
	headers := NewHeaders()
	headers.Set("content-length", strconv.Itoa(contentSize))
	headers.Set("connection", "close")
	headers.Set("content-type", "text/plain")
	return headers
}

func (h Headers) Replace(key, value string) {
	h[strings.ToLower(key)] = value
}

func (h Headers) Remove(key string) {
	delete(h, key)
}

func (h Headers) GetAsStringWithoutFinalTermination() string {
	headersString := ""
	for k, v := range h {
		headersString += fmt.Sprintf("%v: %v%v", k, v, constants.SEPARATOR)
	}
	return headersString
}

func (h Headers) GetAsString() string {
	headersString := ""
	for k, v := range h {
		headersString += fmt.Sprintf("%v: %v%v", k, v, constants.SEPARATOR)
	}
	headersString += constants.SEPARATOR
	return headersString
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
	if !validHeaderNamesRegex.Match([]byte(trimmedHeaderName)) {
		return Headers{}, fmt.Errorf("header: header name contains unsupported characters")
	}
	if strings.Contains(trimmedHeaderName, " ") {
		return Headers{}, fmt.Errorf("header: header name contains spaces")
	}
	headers[trimmedHeaderName] = strings.TrimSpace(headerValue)
	return headers, nil
}

func (h Headers) normalizeHeaderKeys() Headers {
	normalizedHeaders := make(Headers, len(h))
	for k := range h {
		normalizedHeaders[strings.ToLower(k)] = h[k]
	}
	return normalizedHeaders
}

func (h1 Headers) mergeHeaders(h2 Headers) {
	h3 := utils.MergeMaps(h1, h2, resolver)
	maps.Copy(h1, h3)
}

func resolver(a, b string) string {
	return fmt.Sprintf("%s, %s", a, b)
}
