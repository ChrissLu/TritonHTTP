package tritonhttp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	responseProto = "HTTP/1.1"
)

type Server struct {
	// Addr specifies the TCP address for the server to listen on,
	// in the form "host:port". It shall be passed to net.Listen()
	// during ListenAndServe().
	Addr string // e.g. ":0"

	// VirtualHosts contains a mapping from host name to the docRoot path
	// (i.e. the path to the directory to serve static files from) for
	// all virtual hosts that this server supports
	//VirtualHosts string
	VirtualHosts map[string]string
}

// ListenAndServe listens on the TCP network address s.Addr and then
// handles requests on incoming connections.
func (s *Server) ListenAndServe() error {

	//Hint: Validate all docRoots, not necessary, already in main
	// if err := s.ValidateServerSetup(); err != nil {
	// 	return fmt.Errorf("server is not setup correctly %v", err)
	// }
	// fmt.Println("Server setup valid!")

	// Hint: create your listen socket and spawn off goroutines per incoming client
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		log.Println("Error: Listening Socket", err.Error())
		return err
	}
	defer listener.Close()

	//accept connections forever
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		log.Println("accepted connection", conn.RemoteAddr())
		go s.HandleConnection(conn)
	}
}

// func (s *Server) ValidateServerSetup() error {
// 	Info, err := os.Stat(s.VirtualHosts)
// 	if err != nil {
// 		return err
// 	}
// 	if !Info.IsDir() {
// 		return fmt.Errorf("not a Dir")
// 	}
// 	return nil

// }

func (s *Server) HandleConnection(conn net.Conn) {
	br := bufio.NewReader(conn)
	for {
		// Set timeout
		if err := conn.SetReadDeadline(time.Now().Add(4 * time.Second)); err != nil {
			log.Printf("Failed to set timeout for connection %v", conn)
			_ = conn.Close()
			return
		}

		// Read next request from the client
		req, err := ReadRequest(br)

		// Handle EOF
		if errors.Is(err, io.EOF) {
			log.Printf("Connection closed by %v", conn.RemoteAddr())
			_ = conn.Close()
			return
		}

		// Handle timeout
		if err, ok := err.(net.Error); ok && err.Timeout() {
			if !req.Sent_some_bytes {
				log.Printf("Connection to %v timed out", conn.RemoteAddr())
				_ = conn.Close()
				return
			} else {
				res := &Response{}
				res.HandleBadRequest()
				_ = res.Write(conn)
				_ = conn.Close()
				return
			}
			// log.Printf("Connection to %v timed out", conn.RemoteAddr())
			// _ = conn.Close()
			// return
		}

		// Handle bad request
		if err != nil {
			log.Printf("Handle bad request for error: %v", err)
			res := &Response{}
			res.HandleBadRequest()
			_ = res.Write(conn)
			_ = conn.Close()
			return
		}

		// Handle good request
		log.Printf("Handle good request: %v", req)
		res := s.HandleGoodRequest(req)
		err = res.Write(conn)
		if err != nil {
			fmt.Println(err)
		}

		if req.Close {
			_ = conn.Close()
			return
		}

	}

}

func (res *Response) init() {
	res.Proto = responseProto
	res.Headers = make(map[string]string)
	res.Headers["Date"] = FormatTime(time.Now())
}

func (res *Response) HandleBadRequest() {
	res.init()
	res.StatusCode = 400
	res.FilePath = ""
	res.Headers["Connection"] = "close"
}

func (s *Server) HandleGoodRequest(req *Request) (res *Response) {
	res = &Response{}
	res.init()
	DocRoot := s.VirtualHosts[req.Host]
	res.FilePath = filepath.Join(DocRoot, req.URL)
	res.FilePath = filepath.Clean(res.FilePath)

	if res.FilePath[:len(DocRoot)] != DocRoot {
		// for security reason
		res.HandleNotFound(req)
	} else if _, err := os.Stat(res.FilePath); os.IsNotExist(err) {
		res.HandleNotFound(req)
	} else {
		res.HandleOK()
		if req.Close {
			res.Headers["Connection"] = "close"
		}
	}
	return res
}

func (res *Response) HandleOK() {
	res.StatusCode = 200
	res.Headers["Content-Type"] = MIMETypeByExtension(filepath.Ext(res.FilePath))
	// file, err := os.Open(res.FilePath)
	// if err != nil {
	// 	fmt.Println("Error opening file", err.Error())
	// 	return
	// }
	// defer file.Close()
	// Info, err := file.Stat()
	Info, err := os.Stat(res.FilePath)
	if err != nil {
		fmt.Println("Error getting file Info", err.Error())
	}
	res.Headers["Content-Length"] = strconv.FormatInt(Info.Size(), 10)
	res.Headers["Last-Modified"] = FormatTime(Info.ModTime())
}
func (res *Response) HandleNotFound(req *Request) {
	res.StatusCode = 404
	res.FilePath = ""
	if req.Close {
		res.Headers["Connection"] = "close"
	}
}

func ReadLine(br *bufio.Reader) (string, error) {
	var line string
	for {
		s, err := br.ReadString('\n')
		line += s
		// Return the error
		if err != nil {
			return line, err
		}
		// Return the line when reaching line end
		if strings.HasSuffix(line, "\r\n") {
			// Striping the line end
			line = line[:len(line)-2]
			fmt.Println(line)
			return line, nil
		}
	}
}
