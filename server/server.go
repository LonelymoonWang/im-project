package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// Server 服务端结构体
type Server struct {
	Ip   string
	Port int

	// 在线用户
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	// 消息管道
	Message chan string
}

// NewServer 创建一个server
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

// Start 启动服务器 Server的构造器
func (s *Server) Start() {
	// 使用net.listen 监听TCP networks
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))

	if err != nil {
		fmt.Println("net listen err:", err)
		return
	}

	defer func(listen net.Listener) {
		err := listen.Close()
		if err != nil {
			fmt.Println("net listen close err:", err)
			return
		}
	}(listen)

	// 启动发送message的go routine
	go s.ListenMessage()

	for {
		accept, err := listen.Accept()
		if err != nil {
			fmt.Println("net listen accept err:", err)
			continue
		}
		go s.Handler(accept)
	}
}

// ListenMessage 发送消息
func (s *Server) ListenMessage() {
	for {
		msg := <-s.Message

		s.mapLock.Lock()

		for _, user := range s.OnlineMap {
			user.c <- msg
		}

		s.mapLock.Unlock()

	}
}

// BroadCast 广播信息给用户
func (s *Server) BroadCast(user *User, message string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + message
	s.Message <- sendMsg
}

// Handler 处理服务端的任务 Server的构造器
func (s *Server) Handler(accept net.Conn) {
	//fmt.Println("服务端连接成功！")

	user := NewUser(accept, s)

	// 用户上线
	user.OnLine()

	// 监听用户是否活跃的管道
	isLive := make(chan bool)

	go func() {
		buf := make([]byte, 4096)
		for {
			//var buf [128]byte
			read, err := accept.Read(buf)
			if read == 0 {
				user.OffLine()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("conn read err", err)
				return
			}
			//decodeRune, _ := utf8.DecodeRune(buf[:read-1])

			msg := string(buf[:read-1])

			user.DoMessage(msg)

			// 用户的任意消息，代表当前用户是一个活跃的用户
			isLive <- true

		}
	}()

	// 当前handle阻塞
	for {
		select {
		case <-isLive:
			// 当前用户是活跃的，应该重置定时器
			// 不做任何事情，为了激活select 更新下面的定时器
		case <-time.After(time.Second * 60 * 60):
			// 已经超时
			// 将当前的用户强制关闭
			user.SendMsg("Good Bye!")

			// 销毁用的资源
			close(user.c)

			// 关闭连接
			err := accept.Close()
			if err != nil {
				return
			}

			return
		}
	}
}
