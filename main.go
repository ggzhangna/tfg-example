package main

import (
	"fmt"
	"github.com/scriptllh/tfg"
	log "github.com/sirupsen/logrus"
	"time"
)

type HandleConn struct {
	tfg.BaseHandleConn
}

func (hc *HandleConn) PreOpen(c tfg.Conn) {
	log.Infof("pre conn open: [conn:%v]", c)
}

type req struct {
	s string
}


/**
 * req : in => 请求读出来的字节    lastRemain => 上一次read 操作没有处理完的数据
 *
 * resp: packet => 处理完后生成业务的结构数据   remain => 这一节数据不够转换成业务数据下次read的时候再处理
 *    isFinRead => 是否继续处理这次留下来的remain，如果为false,则下次数据会按顺序过来，加上这次留下的remain，且是同一个协程处理
 *         如果为true => 则下次数据不会带上remain，且下次read是下一个协程来处理
 *    isHandle => 是否有数据packet给handle执行 ，因为handle是异步执行的
 */
func (hc *HandleConn) Read(in []byte, lastRemain []byte) (packet interface{}, remain []byte, isFinRead bool, isHandle bool) {
	s := string(in)
	req := &req{
		s: s,
	}
	log.Infof("read [data:%v]", req)
	return req, nil, false, true
}

/**
 * req : conn => 连接    packet => read 处理过后的业务数据packet
 *
 */

func (hc *HandleConn) Handle(conn tfg.Conn, packet interface{}) {
	req := packet.(*req)
	log.Infof("handle req [data:%v]", req)
	time.Sleep(time.Millisecond * time.Duration(10))
	n, err := conn.Write([]byte("tfg la la la"))
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Infof("handle write [n:%v]", n)
}

func main() {
	var handleConn HandleConn
	s, err := tfg.NewServer(":6000", &handleConn, 0, tfg.RoundRobin)
	if err != nil {
		log.Errorf("new server [err:%v]", err)
		return
	}
	if err := s.Serve(); err != nil {
		log.Errorf("serve [err:%v]", err)
		return
	}
}
