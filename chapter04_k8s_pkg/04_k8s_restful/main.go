package main

import (
	"io"
	"log"
	"net/http"

	"github.com/emicklei/go-restful/v3"
)

// GET http://localhost:8080/secret
// and use admin,admin for the credentials

func main() {

	// 创建WebService
	ws := new(restful.WebService)

	// 为WebService设置路由和回调函数
	ws.Route(ws.GET("/secret").
		Filter(basicAuthenticate).
		To(secret)).
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	// 将WebService添加到默认生成的Container中
	// 默认生成的container的代码在web_service_container.go的init方法中
	restful.Add(ws)

	// 启动服务
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// 路由对应的回调函数
func basicAuthenticate(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	// usr/pwd = admin/admin
	u, p, ok := req.Request.BasicAuth()
	if !ok || u != "admin" || p != "admin" {
		resp.AddHeader("WWW-Authenticate", "Basic realm=Protected Area")
		resp.WriteErrorString(401, "401: Not Authorized")
		return
	}
	chain.ProcessFilter(req, resp)
}

func secret(req *restful.Request, resp *restful.Response) {
	io.WriteString(resp, "42")
}
