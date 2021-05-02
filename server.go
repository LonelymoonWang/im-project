package main

import (
	"fmt"
	"net"
)

// Server 服务端结构体
type Server struct {
	Ip   string
	Port int
}

// NewServer 创建一个server 的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
	}

	return server
}

// Start 启动服务器的接口
func (s *Server) Start() {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net listen err:", err)
		return
	}
	defer func(listen net.Listener) {
		err := listen.Close()
		if err != nil {
			fmt.Println("net listen close err:", err)
		}
	}(listen)

	for {
		accept, err := listen.Accept()
		if err != nil {
			fmt.Println("net listen accept err:", err)
			continue
		}
		go s.Handle(accept)

	}
}

// Handle 处理服务端的任务
func (s *Server) Handle(accept net.Conn) {
	fmt.Println("服务端连接成功！")
}
