package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C chan string
	conn net.Conn

	server *Server
}

//创建一个用户的API
func NewUser(conn net.Conn,server *Server) *User{
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}
	//启动监听当前的user channel 消息的goroutine
	go user.ListenMessage()
	return user
}

//用户的上线业务
func (this *User) Online(){
	//用戶上线，将用户加入OnlineMap
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	//广播当前用户上线的消息
	this.server.BroadCast(this,"已上线")
}

//用户下线业务
func (this *User) Offline(){
	//用戶下线，将用户加入OnlineMap
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap,this.Name)
	this.server.mapLock.Unlock()

	//广播当前用户下线的消息
	this.server.BroadCast(this,"已下线")
}

//当前user客户端发送消息
func (this *User) SendMsg(msg string){
	this.conn.Write([]byte(msg))
}

//用户处理消息业务
func (this *User) DoMessage(msg string){
	if msg == "who"{
		//
		this.server.mapLock.Lock()
		for _,user := range this.server.OnlineMap{
			onlineMsg := "["+user.Addr+"]"+user.Name+":在线...\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()

	} else if len(msg) >7 && msg[:7] == "rename|" {
		newName := strings.Split(msg,"|")[1]
		//判断新名字是否已经存在
		_,ok := this.server.OnlineMap[newName]
		if ok{
			this.SendMsg("当前用户名被使用！")
		} else{
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap,this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("您已经更新用户名:"+this.Name+"\n")
		}
	} else if len(msg) >4 && msg[:3] == "to|" {
		//获取用户名
		remoteName := strings.Split(msg,"|")[1]
		if remoteName == ""{
			this.SendMsg("消息格式不正确，请使用\"to|张三|你好呀\"")
			return
		}
		//通过用户名得到user对象
		remoteUser,ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("该用户值不存在")
			return
		}
		//获取消息内容将消息发给user对象
		content := strings.Split(msg,"|")[2]
		if content == ""{
			this.SendMsg("发送消息不能为空")
			return
		}
		remoteUser.SendMsg(this.Name+"对您说："+content)
	}else{
		this.server.BroadCast(this,msg)
	}
}

//监听当前的User channel 的方法，一旦有消息，就直接发送给对端客户端
func (this *User) ListenMessage(){
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg+"\n"))
	}
}
