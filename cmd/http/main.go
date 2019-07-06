package main

import (
	"bytes"
	"fmt"
	"github.com/scriptllh/tfg"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

var res = "Hello World!\r\n"

type HandleConn struct {
	tfg.BaseHandleConn
	noparse bool
}

func (hc *HandleConn) PreOpen(c tfg.Conn) {
	log.Infof("pre conn open: [conn:%v]", c)
}

/**
 * req : in => 请求读出来的字节    lastRemain => 上一次read 操作没有处理完的数据
 *
 * resp: packet => 处理完后生成业务的结构数据   remain => 这一节数据不够转换成业务数据下次read的时候再处理
 *    isFinRead => 是否继续处理这次留下来的remain，如果为false,则下次数据会按顺序过来，加上这次留下的remain，且是同一个协程处理
 *         如果为true => 则下次数据不会带上remain，且下次read是下一个协程来处理
 *    isHandle => 是否有数据packet给handle执行 ，因为handle是异步执行的
 */
func (hc *HandleConn) Read(in []byte, lastRemain []byte) (packet interface{}, remain []byte, isFinRead bool, isHandle bool, err error) {
	log.Infof("read [in:%v]", string(in))
	if hc.noparse {
		if bytes.Contains(in, []byte("\r\n\r\n")) {
			return nil, nil, true, true, nil
		}
		return nil, nil, false, false, nil
	}
	var req request
	leftover, err := parsereq(in, &req)
	if err != nil {
		// bad thing happened
		return nil, nil, true, true, err
	}

	if len(leftover) == len(in) {
		// request not ready, yet
		return nil, leftover, false, false, nil
	}

	return &req, leftover, false, true, nil
}

/**
 * req : conn => 连接    packet => read 处理过后的业务数据packet
 *
 */

func (hc *HandleConn) Handle(conn tfg.Conn, packet interface{}, err error) {
	log.Infof("handle [packet:%v]", packet)
	if hc.noparse {
		resp := genResp(nil, "200 OK", "", res)
		conn.Write(resp)
		return
	}
	if err != nil || packet == nil {
		// bad thing happened
		resp := genResp([]byte{}, "500 Error", "", err.Error()+"\n")
		conn.Write(resp)
		conn.Close()
		return
	}

	req := packet.(*request)
	req.remoteAddr = conn.RemoteAddr().String()
	resp := genhandle([]byte{}, req)
	conn.Write(resp)
}

func main() {
	handleConn := &HandleConn{
		noparse: true,
	}
	s, err := tfg.NewServer(":6000", handleConn, 0, tfg.RoundRobin)
	if err != nil {
		log.Errorf("new http server [err:%v]", err)
		return
	}
	log.Infof("http server start :%v", "6000")
	if err := s.Serve(); err != nil {
		log.Errorf("serve [err:%v]", err)
		return
	}
}

type request struct {
	proto, method string
	path, query   string
	head, body    string
	remoteAddr    string
}

func genhandle(b []byte, req *request) []byte {
	return genResp(b, "200 OK", "", res)
}

func genResp(b []byte, status, head, body string) []byte {
	b = append(b, "HTTP/1.1"...)
	b = append(b, ' ')
	b = append(b, status...)
	b = append(b, '\r', '\n')
	b = append(b, "Server: tfg\r\n"...)
	b = append(b, "Date: "...)
	b = time.Now().AppendFormat(b, "Mon, 02 Jan 2006 15:04:05 GMT")
	b = append(b, '\r', '\n')
	if len(body) > 0 {
		b = append(b, "Content-Length: "...)
		b = strconv.AppendInt(b, int64(len(body)), 10)
		b = append(b, '\r', '\n')
	}
	b = append(b, head...)
	b = append(b, '\r', '\n')
	if len(body) > 0 {
		b = append(b, body...)
	}
	return b
}

func parsereq(data []byte, req *request) (leftover []byte, err error) {
	sdata := string(data)
	var i, s int
	var top string
	var clen int
	var q = -1
	// method, path, proto line
	for ; i < len(sdata); i++ {
		if sdata[i] == ' ' {
			req.method = sdata[s:i]
			for i, s = i+1, i+1; i < len(sdata); i++ {
				if sdata[i] == '?' && q == -1 {
					q = i - s
				} else if sdata[i] == ' ' {
					if q != -1 {
						req.path = sdata[s:q]
						req.query = req.path[q+1 : i]
					} else {
						req.path = sdata[s:i]
					}
					for i, s = i+1, i+1; i < len(sdata); i++ {
						if sdata[i] == '\n' && sdata[i-1] == '\r' {
							req.proto = sdata[s:i]
							i, s = i+1, i+1
							break
						}
					}
					break
				}
			}
			break
		}
	}
	if req.proto == "" {
		return data, fmt.Errorf("malformed request")
	}
	top = sdata[:s]
	for ; i < len(sdata); i++ {
		if i > 1 && sdata[i] == '\n' && sdata[i-1] == '\r' {
			line := sdata[s : i-1]
			s = i + 1
			if line == "" {
				req.head = sdata[len(top)+2 : i+1]
				i++
				if clen > 0 {
					if len(sdata[i:]) < clen {
						break
					}
					req.body = sdata[i : i+clen]
					i += clen
				}
				return data[i:], nil
			}
			if strings.HasPrefix(line, "Content-Length:") {
				n, err := strconv.ParseInt(strings.TrimSpace(line[len("Content-Length:"):]), 10, 64)
				if err == nil {
					clen = int(n)
				}
			}
		}
	}
	// not enough data
	return data, nil
}
