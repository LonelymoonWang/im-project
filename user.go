package main

import "net"

type User struct {
	Name string
	Addr string
	c    chan string
	conn net.Conn

	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		c:      make(chan string),
		conn:   conn,
		server: server,
	}

	// 启动接收message的go routine
	go user.ListenMessage()

	return user
}

// OnLine 用户一上线就把用户信息存进server里的上线用户map中，并且广播已上线信息
func (user *User) OnLine() {

	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()

	user.server.BroadCast(user, "已上线！")

}

// OffLine 下线功能，用户一退出系统，就更新用户该server中的在线用户列表
func (user *User) OffLine() {
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()

	user.server.BroadCast(user, "已下线！")
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

// DoMessage 用户处理消息的业务
func (user *User) DoMessage(msg string) {
	if msg == "who" {
		user.server.mapLock.Lock()
		for _, u := range user.server.OnlineMap {
			onlineMsg := "[" + u.Addr + "]" + u.Name + ":" + "在线！\n"
			user.SendMsg(onlineMsg)
		}
		user.server.mapLock.Unlock()
	} else {
		user.server.BroadCast(user, msg)
	}
}

// SendMsg 给当前用户的客户端发送信息
func (user *User) SendMsg(msg string) {
	_, err := user.conn.Write([]byte(msg))
	if err != nil {
		return
	}
}
