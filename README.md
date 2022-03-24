
# golang_webscoket
## websocket服务器demo
**websocket 客户端测试网站推荐
http://www.websocket-test.com/**


启动项目
1.  go mod init 项目名
2.   go mod tidy
3.  go run *.go or go build

入口文件
`main.go`
```golang
func main() {
	fmt.Println("websocket listen on 0.0.0.0:8081 ")
	log.SetFlags(0)
	//创建websocket服务器
	server := NewGameServer("0.0.0.0", 8081)
	//服务启动
	server.Start()
}
```

