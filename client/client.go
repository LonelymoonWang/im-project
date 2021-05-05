package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int // 当前client的模式
}

// NewClient 创建一个客户端对象
func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       9999,
	}
	dial, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net dial error:", err)
		return nil
	}
	client.conn = dial
	return client
}

var serverIp string
var serverPort int

//./client.exe -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "default 127.0.0.1")
	flag.IntVar(&serverPort, "port", 8888, "default 8888")
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.Menu() != true {

		}
		switch client.flag {
		case 1:
			client.PublicChat()
			break
		case 2:
			client.PrivateChat()
			break
		case 3:
			client.UpdateName()
			break
		}
	}
}

// Menu 创建菜单
func (client *Client) Menu() bool {
	var num int

	fmt.Println("1.public chat")
	fmt.Println("2.private chat")
	fmt.Println("3.rename")
	fmt.Println("0.exit")

	_, err := fmt.Scanln(&num)
	if err != nil {
		return false
	}

	if num >= 0 && num <= 3 {
		client.flag = num
		return true
	} else {
		fmt.Println(">>>>请输入合法范围内的数字<<<<")
		return false
	}
}

// PublicChat 公聊
func (client *Client) PublicChat() {
	//提示用户输入消息
	var chatMsg string

	fmt.Println(">>>>please input chat message,until input [exit].")
	_, err := fmt.Scanln(&chatMsg)
	if err != nil {
		return
	}

	for chatMsg != "exit" {
		//发给服务器

		//消息不为空则发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>please input chat message,until input [exit].")
		_, err := fmt.Scanln(&chatMsg)
		if err != nil {
			return
		}
	}
}

// PrivateChat 私聊
func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.SelectUsers()
	fmt.Println(">>>>please input username, until input [exit]:")
	_, err := fmt.Scanln(&remoteName)
	if err != nil {
		return
	}

	for remoteName != "exit" {
		fmt.Println(">>>>please input chat message,until input [exit].")
		_, err := fmt.Scanln(&chatMsg)
		if err != nil {
			return
		}

		for chatMsg != "exit" {
			//消息不为空则发送
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>>>please input chat message,until input [exit].")
			_, err := fmt.Scanln(&chatMsg)
			if err != nil {
				return
			}
		}

		client.SelectUsers()
		fmt.Println(">>>>please input username, until input [exit]:")
		_, err = fmt.Scanln(&remoteName)
		if err != nil {
			return
		}
	}
}

// UpdateName 更新用户名
func (client *Client) UpdateName() bool {
	fmt.Println(">>>>please input username:")
	_, err := fmt.Scanln(&client.Name)
	if err != nil {
		return false
	}

	sendMsg := "rename|" + client.Name + "\n"
	_, err = client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

// SelectUsers 查询在线用户
func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write err:", err)
		return
	}
}

// DealResponse 处理server回应的消息， 直接显示到标准输出即可
func (client *Client) DealResponse() {
	//一旦client.conn有数据，就直接copy到stdout标准输出上, 永久阻塞监听
	io.Copy(os.Stdout, client.conn)
}

func main() {

	// 命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("dial error!")
		return

	}

	//开一个go程处理服务端的回执消息
	go client.DealResponse()

	fmt.Println("dial success!")

	// 启动客户端的run方法--业务方法
	client.Run()

}
