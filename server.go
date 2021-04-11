package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	IP   string
	Port int
	//在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播的channel
	Message chan string
}

//创建一个server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

//监听message广播消息channel的goroutine,一旦有消息就发送给全部的在线User
func (this *Server) ListenMessager() {
	for {
		//将msg发送给全部的在线user
		msg := <-this.Message
		this.mapLock.Lock()
		for _,cli := range this.OnlineMap{
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

//广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}
func (this *Server) Handler(conn net.Conn) {
	//当前链接的业务
	//fmt.Println("连接建立成功")
	user := NewUser(conn,this)
	user.Online()
	//监听用户是否活跃的channel
	isLive := make(chan bool)
	// 接收客户端发送的消息
	go func() {
		buf := make([]byte,4096)
		for{
			n,err := conn.Read(buf)
			if n == 0{
				user.Offline()
				return
			}
			if err != nil && err != io.EOF{
				fmt.Println("Conn Read err:",err)
				return
			}

			//缓存用户的消息，去除\n
			msg:= string(buf[:n-1])

			//用户针对msg进行处理
			user.DoMessage(msg)

			//用户的任意消息代表当前用户是活跃的
			isLive <- true
		}
	}()
	//当前handler阻塞
	for{
		select {
		case <- isLive:
			//当前用户活跃应该重置定时器
			//不做任何事情，为了激活select,更新下面的定时器

		case <- time.After(time.Second * 60):
			//已经超时
			user.SendMsg("你被踢了")
			//将当前的user强制关闭
			close(user.C)
			conn.Close()
			//退出当前的handler
			return

		}

	}


}

//启动服务器接口
func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.IP, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	//close listen socket
	defer listener.Close()

	//启动监听message的goroutine
	go this.ListenMessager()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Listener accept err:", err)
			continue
		}
		//do handler
		go this.Handler(conn)
	}
}
