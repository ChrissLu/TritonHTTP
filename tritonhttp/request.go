package tritonhttp

import (
	"bufio"
	"fmt"
	"strings"
)

type Request struct {
	Method string // e.g. "GET"
	URL    string // e.g. "/path/to/a/file"
	Proto  string // e.g. "HTTP/1.1"

	// Headers stores the key-value HTTP headers
	Headers map[string]string

	Host            string // determine from the "Host" header
	Close           bool   // determine from the "Connection" header
	Sent_some_bytes bool
}

func ReadRequest(br *bufio.Reader) (req *Request, err error) {
	req = &Request{}
	req.Sent_some_bytes = false

	// Read start line
	line, err := ReadLine(br)
	if err != nil {
		return nil, err
	}

	//initial request line
	fields := strings.SplitN(line, " ", 3)
	if len(fields) != 3 {
		fmt.Println(111111)
		return nil, fmt.Errorf("400")
	}
	req.Method = fields[0]
	req.Sent_some_bytes = true
	if req.Method != "GET" {
		fmt.Println(22222)
		return nil, fmt.Errorf("400")
	}
	req.URL = fields[1]
	req.Proto = fields[2]
	if req.Proto != "HTTP/1.1" { //should be moved?
		fmt.Println(33333)
		return nil, fmt.Errorf("400")
	}

	// read headers
	req.Headers = make(map[string]string)
	req.Close = false
	hasHost := false
	for {
		line, err := ReadLine(br)
		if err != nil {
			return nil, err
		}
		if line == "" {
			// This marks header end
			break
		}

		fields := strings.SplitN(line, ":", 2)
		if len(fields) != 2 {
			fmt.Println(4444)
			return nil, fmt.Errorf("400")
		}
		key := CanonicalHeaderKey(strings.TrimSpace(fields[0]))
		value := strings.TrimSpace(fields[1])

		if key == "Connection" && value == "close" {
			req.Close = true
		} else if key == "Host" {
			req.Host = value
			hasHost = true
		} else {
			fmt.Println(key)
			req.Headers[key] = value
		}
	}

	if !hasHost {
		fmt.Println(5555)
		return nil, fmt.Errorf("400")
	}
	return req, nil
}
