package main

import "net"

type User struct {
	Name string
	Addr string
	c    chan string
	conn net.Conn
}

func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		c:    make(chan string),
		conn: conn,
	}

	// 启动接收message的go routine
	go user.ListenMessage()

	return user
}

// ListenMessage 接收消息
func (user *User) ListenMessage() {
	for {
		msg := <-user.c
		_, err := user.conn.Write([]byte(msg + "\n"))
		if err != nil {
			return
		}
	}
}
