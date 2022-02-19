package main

import (
	"fmt"
	"log"
	"strconv"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

var uuid int64 = 0

//玩家类
type User struct {
	id         int64
	name       string
	gameServer *GameServer
	mapIndex   int64
	x          float64
	y          float64
	con        *websocket.Conn
	cha        chan string
}

//初始化对象
func NewUser(con *websocket.Conn, server *GameServer) *User {
	atomic.AddInt64(&uuid, 1)
	user := &User{
		id:         atomic.LoadInt64(&uuid),
		name:       strconv.FormatInt(atomic.LoadInt64(&uuid), 10),
		gameServer: server,
		mapIndex:   0,
		x:          0,
		y:          0,
		con:        con,
		cha:        make(chan string),
	}
	//监听消息
	go user.ListenMessage()
	return user
}

//业务处理
func (u *User) DoMessage(msg string) {
	//异常处理
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("有致命错误:", r)
			return
		}
	}()
	//who指令
	if msg == "who" {
		var addStr string
		u.gameServer.mapLocak.RLock()
		for k, v := range u.gameServer.OnlineMap {
			if k != u.id {
				addStr += v.name + " "
			}
		}
		u.gameServer.mapLocak.RUnlock()
		if addStr != "" {
			u.SendMsgToCustomer("server广播消息:" + addStr + "在线")
			return
		} else {
			u.SendMsgToCustomer("server广播消息:无人在线")
			return
		}
	}
	//change| 指令 修改名字
	if len(msg) >= 7 && msg[:7] == "change|" {
		changedName := msg[7:]
		if changedName == "" {
			u.SendMsgToCustomer("server广播消息: 名字修改不成功")
			return
		} else {
			u.name = changedName
			u.SendMsgToCustomer("server广播消息: 名字修改成功")
			return
		}

	}
	//myName| 指令 查看自己名字
	if msg == "myName" {
		u.SendMsgToCustomer("myName:" + u.name)
		return
	}
	//whoNotMe| 指令 查看除了自己之外的人
	if msg == "whoNotMe" {
		var strList string
		//显示除自己之外的玩家
		u.gameServer.mapLocak.RLock()
		for k, v := range u.gameServer.OnlineMap {
			if k != u.id {
				x := fmt.Sprintf("%.2f", v.x)
				y := fmt.Sprintf("%.2f", v.y)
				strList += v.name + "$" + x + "$" + y + "|"
			}
		}
		u.gameServer.mapLocak.RUnlock()
		if strList != "" {
			u.SendMsgToCustomer("所有人:" + strList[:len(strList)-1])
			return
		} else {
			u.SendMsgToCustomer("没有人")
			return
		}
	}
	//心跳检测
	if msg == "ping" {
		return
	}
	//广播消息
	u.gameServer.BroadCast(u, msg)
}

//监听消息
func (u *User) ListenMessage() {
	for {
		msg, ok := <-u.cha
		if !ok {
			return
		}
		//发送消息
		u.SendMsgToCustomer(msg)
	}
}

//往客户端发送消息
func (u *User) SendMsgToCustomer(msg string) {
	u.con.WriteMessage(1, []byte(msg))
}

//登录User到OnlineMap
func (u *User) AddUser() {

	u.gameServer.mapLocak.Lock()
	u.gameServer.OnlineMap[u.id] = u
	u.gameServer.mapLocak.Unlock()
	u.gameServer.BroadCast(u, strconv.FormatInt(u.id, 10)+":上线了")
	log.Printf("当前在线人数为:%d", len(u.gameServer.OnlineMap))
}

//从OnlineMap删除User
func (u *User) DelUser() {
	u.gameServer.mapLocak.Lock()
	delete(u.gameServer.OnlineMap, u.id)
	u.gameServer.mapLocak.Unlock()
	u.gameServer.BroadCast(u, strconv.FormatInt(u.id, 10)+":下线了")
	log.Printf("当前在线人数为:%d", len(u.gameServer.OnlineMap))
}
