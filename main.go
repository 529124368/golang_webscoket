package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("websocket listen on 0.0.0.0:8081 ")
	log.SetFlags(0)
	//创建websocket服务器
	server := NewGameServer("0.0.0.0", 8081)
	//服务启动
	server.Start()
}
