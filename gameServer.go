package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//游戏服务器类
type GameServer struct {
	Ip        string
	Port      int
	OnlineMap map[int64]*User
	msg       chan string
	mapLocak  sync.RWMutex
	upgrader  websocket.Upgrader
}

//websocket通信升级
func (g *GameServer) handle(w http.ResponseWriter, r *http.Request) {
	con, err := g.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	//创建玩家
	user := NewUser(con, g)
	//登录用户信息到onlinenmap
	user.AddUser()
	isAlive := make(chan bool)

	//异常处理
	defer con.Close()
	//处理链接
	go func() {
		//消息处理
		for {
			_, message, err := con.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				user.DelUser()
				break
			}
			//log.Printf("recv: %s", message)
			user.DoMessage(string(message))
			isAlive <- true
		}

	}()
	//心跳检测
	for {
		select {
		case <-isAlive:
		//超时
		case <-time.After(time.Second * 600):
			user.DelUser()
			close(user.cha)
			con.Close()
			return
		}
	}
}

//初始化对象
func NewGameServer(ip string, port int) *GameServer {
	//websocket 配置
	var config = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// 鉴权
		CheckOrigin: func(r *http.Request) bool {
			//鉴权TODO
			// if r.URL.Query().Get("token") != "'" && r.URL.Query().Get("uid") != "" {
			// 	key := "$%zimugeWODGKZNNSJD&^"
			// 	uid := r.URL.Query().Get("uid")
			// 	data := []byte(uid)
			// 	has := md5.Sum(data)
			// 	new_key := fmt.Sprintf("%x", has) + key
			// 	data = []byte(new_key)
			// 	has = md5.Sum(data)
			// 	new_key = fmt.Sprintf("%x", has)
			// 	if new_key == r.URL.Query().Get("token") {
			// 		return true
			// 	} else {
			// 		return false
			// 	}

			// } else {
			// 	return false
			// }
			return true

		},
	}
	gameServer := &GameServer{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[int64]*User),
		msg:       make(chan string),
		upgrader:  config,
	}
	//启动监听
	go gameServer.ListenMessage()
	return gameServer
}

//消息监听
func (g *GameServer) ListenMessage() {
	for {
		msg := <-g.msg
		end := strings.Index(msg, "^")
		res := msg[:end]
		g.mapLocak.RLock()
		for _, cn := range g.OnlineMap {
			if res != cn.name {
				cn.cha <- msg[end+1:]
			}
		}
		g.mapLocak.RUnlock()
	}

}

//广播消息
func (g *GameServer) BroadCast(user *User, msg string) {
	sendMsg := user.name + "^" + msg
	g.msg <- sendMsg

}

//开启服务
func (g *GameServer) Start() {
	//http 路由注册
	http.HandleFunc("/", g.handle)
	//启动HTTP服务器
	http.ListenAndServe(fmt.Sprintf("%s:%d", g.Ip, g.Port), nil)
}
