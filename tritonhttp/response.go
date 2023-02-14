package tritonhttp

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

var statusText = map[int]string{
	200: "OK",
	400: "Bad Request",
	404: "Not Found",
}

type Response struct {
	Proto      string // e.g. "HTTP/1.1"
	StatusCode int    // e.g. 200
	StatusText string // e.g. "OK"

	// Headers stores all headers to write to the response.
	Headers map[string]string

	// Request is the valid request that leads to this response.
	// It could be nil for responses not resulting from a valid request.
	// Hint: you might need this to handle the "Connection: Close" requirement
	Request *Request

	// FilePath is the local path to the file to serve.
	// It could be "", which means there is no file to serve.
	FilePath string
}

func (res *Response) Write(w io.Writer) error {
	bw := bufio.NewWriter(w)

	// write initial response line
	statusLine := fmt.Sprintf("%v %v %v\r\n", res.Proto, res.StatusCode, statusText[res.StatusCode])
	if _, err := bw.WriteString(statusLine); err != nil {
		return err
	}

	//write headers
	if res.StatusCode == 200 {
		if _, ok := res.Headers["Connection"]; ok {
			if _, err := bw.WriteString("Connection: " + res.Headers["Connection"] + "\r\n"); err != nil {
				return err
			}
		}
		if _, err := bw.WriteString("Content-Length: " + res.Headers["Content-Length"] + "\r\n"); err != nil {
			return err
		}
		if _, err := bw.WriteString("Content-Type: " + res.Headers["Content-Type"] + "\r\n"); err != nil {
			return err
		}
		if _, err := bw.WriteString("Date: " + res.Headers["Date"] + "\r\n"); err != nil {
			return err
		}
		if _, err := bw.WriteString("Last-Modified: " + res.Headers["Last-Modified"] + "\r\n"); err != nil {
			return err
		}

	}

	if res.StatusCode == 400 {
		if _, ok := res.Headers["Connection"]; ok {
			if _, err := bw.WriteString("Connection: " + res.Headers["Connection"] + "\r\n"); err != nil {
				return err
			}
		}
		if _, err := bw.WriteString("Date: " + res.Headers["Date"] + "\r\n"); err != nil {
			return err
		}
	}

	if res.StatusCode == 404 {
		if _, ok := res.Headers["Connection"]; ok {
			if _, err := bw.WriteString("Connection: " + res.Headers["Connection"] + "\r\n"); err != nil {
				return err
			}
		}
		if _, err := bw.WriteString("Date: " + res.Headers["Date"] + "\r\n"); err != nil {
			return err
		}
	}

	if _, err := bw.WriteString("\r\n"); err != nil {
		return err
	}

	// write body
	if res.FilePath != "" {
		file, err := os.ReadFile(res.FilePath)
		if err != nil {
			return err
		} else if _, err := bw.Write(file); err != nil {
			return err
		}
	}

	if err := bw.Flush(); err != nil {
		return err
	}
	return nil
}
