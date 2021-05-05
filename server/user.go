package main

import (
	"net"
	"strings"
)

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

	user.server.BroadCast(user, "Online!")

}

// OffLine 下线功能，用户一退出系统，就更新用户该server中的在线用户列表
func (user *User) OffLine() {
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()

	user.server.BroadCast(user, "OffLine!")
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
			onlineMsg := "[" + u.Addr + "]" + u.Name + ":" + "is Online!\n"
			user.SendMsg(onlineMsg)
		}
		user.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 消息格式 rename|张思
		newName := strings.Split(msg, "|")[1]
		//判断该用户名是否存在
		_, ok := user.server.OnlineMap[newName]
		if ok {
			user.SendMsg("The current user name already exists!\n")
		} else {

			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.Name = newName
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()

			user.SendMsg("User name modified successfully!" + user.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 消息格式是“to|张思|消息内容”
		//1.先获取用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			user.SendMsg("The message format is incorrect,please use \"to|username|msg\"format\n")
			return
		}
		//2.根据用户名获取用户对象
		remoteUser, ok := user.server.OnlineMap[remoteName]
		if !ok {
			user.SendMsg("this user is not exists!\n")
			return
		}
		//3.获取消息内容，通过对方的User对象将消息内容发送出去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			user.SendMsg("no content!\n")
			return
		}
		remoteUser.SendMsg(user.Name + "said to you:" + content + "\n")
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
