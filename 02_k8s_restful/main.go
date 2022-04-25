package main

import (
	"github.com/emicklei/go-restful/v3"
	"io"
	"log"
	"net/http"
)

func main() {
	// 创建WebService
	ws := new(restful.WebService)
	// 为WebService设置路由和回调函数
	ws.Route(ws.GET("/hello").To(hello))
	// 将WebService添加到默认生成的Container中
	// 默认生成的container的代码在web_service_container.go的init方法中
	restful.Add(ws)
	// 启动服务
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// 路由对应的回调函数
func hello(req *restful.Request, resp *restful.Response) {
	io.WriteString(resp, "world")
}