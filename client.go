package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIP string
	ServerPort int
	Name string
	conn net.Conn
	flag int //当前client模式
}

func NewClient(serverIP string,serverPort int) *Client{
	client := &Client{
		ServerIP:   serverIP,
		ServerPort: serverPort,
		flag: 999,
	}

	//连接服务器
	conn,err := net.Dial("tcp",fmt.Sprintf("%s:%d",serverIP,serverPort))
	if err != nil{
		fmt.Println("net.dial error:",err)
	}
	client.conn = conn
	return  client
}

//处理server回应的消息,直接显示到标准输出
func (client *Client) DealResponse(){
	//一旦client.conn有数据，就直接copy到标准输出，永久阻塞监听，不需要自己写for循环来实现
	io.Copy(os.Stdout,client.conn)

}

//显示菜单
func (client *Client) menu() bool{
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")
	fmt.Scanln(&flag)
	if flag >=0 && flag <=3{
		client.flag = flag
		return true
	}else{
		fmt.Println("请输入合法范围内数字")
		return false
	}
}
//公聊模式
func (client *Client) PublicChat(){
	var chatMsg string
	//提示输入信息
	fmt.Println("请输入聊天内容，exit退出")
	fmt.Scanln(&chatMsg)

	for chatMsg!= "exit" {
		// 发给服务器
		//消息不为空发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg +"\n"
			_,err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn write err :",err)
				break
			}
		}
		chatMsg = ""
		fmt.Println("请输入聊天内容，exit退出")
		fmt.Scanln(&chatMsg)
	}
}

//私聊
//查询在线用户
func (client *Client) selectUsers(){
	sendMsg := "who\n"
	_,err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn write err :",err)
		return
	}
}
//私聊模式
func (client *Client) privateChat(){
	var remoteName,chatMsg  string
	client.selectUsers()
	fmt.Println("请输入聊天对象[用户名]，exit退出")
	fmt.Scanln(&remoteName)
	for remoteName != "exit"{
		fmt.Println("请输入消息内容，exit退出")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit"{
			if len(chatMsg) != 0{
				sendMsg := "to|"+remoteName+"|"+chatMsg+"\n\n"
				_,err := client.conn.Write([]byte(sendMsg))
				if err != nil{
					fmt.Println("conn write err : ",err)
					break
				}
			}
			chatMsg = ""
			fmt.Println("请输入消息内容，exit退出")
			fmt.Scanln(&chatMsg)
		}
	}
	client.selectUsers()
	fmt.Println("请输入聊天对象[用户名]，exit退出")
	fmt.Scanln(&remoteName)
}

//改名
func (client *Client) rename() bool{
	fmt.Println("请输入新用户名")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|"+client.Name+"\n"
	_,err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.write error:",err)
		return false
	}
	return true
}

func (client *Client) Run(){
	for client.flag != 0{
		for client.menu() != true{

		}
		//处理不同模式
		switch client.flag {
		case 1:
			//公聊
			client.PublicChat()
			break
		case 2:
			//私聊
			client.privateChat()
			break
		case 3:
			//改名
			client.rename()
			break
		}
	}
}

var serverIP string
var serverPort int
func init(){
	//生成编译的二进制使用文档
	//./client -ip 127.0.0.1 -port 8888
	flag.StringVar(&serverIP,"ip","127.0.0.1","设置服务器的IP，默认是127.0.0.1")
	flag.IntVar(&serverPort,"port",8888,"设置服务器的端口，默认是8888")
}

func main(){
	//命令行解析
	flag.Parse()

	client := NewClient(serverIP,serverPort)
	if client == nil{
		fmt.Println("链接服务器失败")
		return
	}
	//开启goroutine处理server的回执消息
	go client.DealResponse()

	fmt.Println("连接服务器成功")

	client.Run()
}
