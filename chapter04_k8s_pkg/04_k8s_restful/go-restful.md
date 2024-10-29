<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [github.com/emicklei/go-restful  -->k8s使用的web框架](#githubcomemickleigo-restful----k8s%E4%BD%BF%E7%94%A8%E7%9A%84web%E6%A1%86%E6%9E%B6)
  - [数据结构](#%E6%95%B0%E6%8D%AE%E7%BB%93%E6%9E%84)
    - [1. 核心数据结构Route](#1-%E6%A0%B8%E5%BF%83%E6%95%B0%E6%8D%AE%E7%BB%93%E6%9E%84route)
    - [2. webservice](#2-webservice)
    - [3. container](#3-container)
  - [流程分析](#%E6%B5%81%E7%A8%8B%E5%88%86%E6%9E%90)
    - [路由分发函数 dispatch](#%E8%B7%AF%E7%94%B1%E5%88%86%E5%8F%91%E5%87%BD%E6%95%B0-dispatch)
      - [路由选择](#%E8%B7%AF%E7%94%B1%E9%80%89%E6%8B%A9)
  - [第三方应用：k8s](#%E7%AC%AC%E4%B8%89%E6%96%B9%E5%BA%94%E7%94%A8k8s)
  - [参考](#%E5%8F%82%E8%80%83)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# github.com/emicklei/go-restful  -->k8s使用的web框架


go-restful是一个用go语言开发的快速构建restful风格的web框架。k8s最核心的组件kube-apiserver使用到了该框架

## 数据结构
![](.go-restful_images/restful_data_structure.png)

go-restful定义了三个重要的数据结构：

- Route：表示一条路由，包含url、回调处理函数
- Webservice：表示一个服务
- Container：表示一个服务器

三者的关系如下：

- go-restful支持多个container，一个container相当于一个http server，不同的container监控不同的地址和端口
- 每个container可以包含多个webservice，相当于一组不同服务的分类
- 每个webservice包含多个Router（路由），Router根据http请求的URL路由到对应的处理函数（Handler Func）


### 1. 核心数据结构Route
```go
type Route struct {
	ExtensionProperties
	Method   string
	Produces []string
	Consumes []string
	Path     string // webservice root path + described path
	Function RouteFunction // 单路由处理函数
	Filters  []FilterFunction // 拦截器
	If       []RouteSelectionConditionFunction

	// cached values for dispatching
	relativePath string
	pathParts    []string
	pathExpr     *pathExpression // cached compilation of relativePath as RegExp

	// documentation
	Doc                     string
	Notes                   string
	Operation               string
	ParameterDocs           []*Parameter
	ResponseErrors          map[int]ResponseError
	DefaultResponse         *ResponseError
	ReadSample, WriteSample interface{} // structs that model an example request or response payload

	// Extra information used to store custom information about the route.
	Metadata map[string]interface{}

	// marks a route as deprecated
	Deprecated bool

	//Overrides the container.contentEncodingEnabled
	contentEncodingEnabled *bool

	// indicate route path has custom verb
	hasCustomVerb bool

	// if a request does not include a content-type header then
	// depending on the method, it may return a 415 Unsupported Media
	// Must have uppercase HTTP Method names such as GET,HEAD,OPTIONS,...
	allowedMethodsWithoutContentType []string
}

type RouteFunction func(*Request, *Response)
```

### 2. webservice
```go
type WebService struct {
	rootPath       string
	pathExpr       *pathExpression // cached compilation of rootPath as RegExp
	routes         []Route
	produces       []string
	consumes       []string
	pathParameters []*Parameter
	filters        []FilterFunction
	documentation  string
	apiVersion     string

	typeNameHandleFunc TypeNameHandleFunction

	dynamicRoutes bool

	// protects 'routes' if dynamic routes are enabled
	routesLock sync.RWMutex
}
```
WebService拥有一组Route，这组Router有公共的rootPath，

### 3. container
```go
type Container struct {
	webServicesLock        sync.RWMutex
	webServices            []*WebService
	ServeMux               *http.ServeMux
	isRegisteredOnRoot     bool
	containerFilters       []FilterFunction
	doNotRecover           bool // default is true
	recoverHandleFunc      RecoverHandleFunction
	serviceErrorHandleFunc ServiceErrorHandleFunction
	router                 RouteSelector // default is a CurlyRouter (RouterJSR311 is a slower alternative)
	contentEncodingEnabled bool          // default is false
}
```
一个Container包含多个Service，不同的Container监听不同的ip地址或端口，他们之间提供的服务的独立的。

## 流程分析
1. 创建WebService对象

2. 为WebService对象添加路由地址和处理函数
```go
ws.Route(ws.GET("/hello").To(hello))
```

```go
// Get方法内部new了一个RouteBuilder，用于构造Route对象
func (w *WebService) GET(subPath string) *RouteBuilder {
  // 典型的建造者模式用法
  return new(RouteBuilder).typeNameHandler(w.typeNameHandleFunc).servicePath(w.rootPath).Method("GET").Path(subPath)
}

// 建造者模式：给属性赋值
// 其他的方法类似，就不再展开
func (b *RouteBuilder) typeNameHandler(handler TypeNameHandleFunction) *RouteBuilder {
  b.typeNameHandleFunc = handler
  return b
}


// Get方法后，属性并没有完全构造完，handler处理函数是用单独的To方法赋值的
func (b *RouteBuilder) To(function RouteFunction) *RouteBuilder {
  b.function = function
  return b
}

// Path specifies the relative (w.r.t WebService root path) URL path to match. Default is "/".
func (b *RouteBuilder) Path(subPath string) *RouteBuilder {
	b.currentPath = subPath
	return b
}
```

根据RouteBuilder生成Route对象
```go
func (w *WebService) Route(builder *RouteBuilder) *WebService {
  w.routesLock.Lock()
  defer w.routesLock.Unlock()
  // 填充默认值
  builder.copyDefaults(w.produces, w.consumes)
  // 调用RouteBuilder的Build方法，构造Route
  // 并将Route添加到routes列表中
  w.routes = append(w.routes, builder.Build())
  return w
}

// Build方法返回Route对象
func (b *RouteBuilder) Build() Route {
  // ...
  route := Route{
    Method:                           b.httpMethod,
    Path:                             concatPath(b.rootPath, b.currentPath),
    // ...
  }
  route.postBuild()
  return route
}
```

3. 将WebService添加到Container中（这里没有声明Container，用的默认Container）
```go
restful.Add(ws)
```

```go
// 定义全局变量,作为默认的Container
var DefaultContainer *Container

// init函数在别的包import时，自动触发。也就是只要引用了go-restful框架，就会默认有一个Container
func init() {
  DefaultContainer = NewContainer()
  // 这里将Golang中标准http包下的默认路由对象DefaultServeMux赋值给Container的ServeMux
  // 这里要特别注意，正是因为这个地方的逻辑，就能回答前面我们提出的问题。go-restful和http库，通过这个赋值建立了关联关系。
  DefaultContainer.ServeMux = http.DefaultServeMux
}

// 生成默认的container
func NewContainer() *Container {
  return &Container{
    webServices:            []*WebService{},
    ServeMux:               http.NewServeMux(),
    isRegisteredOnRoot:     false,
    containerFilters:       []FilterFunction{},
    doNotRecover:           true,
    recoverHandleFunc:      logStackOnRecover,
    serviceErrorHandleFunc: writeServiceError,
    // 默认的路由选择器用的是CurlyRouter
    router:                 CurlyRouter{},
    contentEncodingEnabled: false}
}

// 将WebService添加到默认Container中
func Add(service *WebService) {
  DefaultContainer.Add(service)
}

// Add
func (c *Container) Add(service *WebService) *Container {
  ...
  // if rootPath was not set then lazy initialize it
  if len(service.rootPath) == 0 {
    service.Path("/")
  }

  // 判断有没有重复的RootPath，不同的WebService，rootPath不能重复
  for _, each := range c.webServices {
    if each.RootPath() == service.RootPath() {
      log.Printf("WebService with duplicate root path detected:['%v']", each)
      os.Exit(1)
    }
  }

  if !c.isRegisteredOnRoot {
    // 核心逻辑：为servcie添加handler处理函数
    // 这里将c.ServeMux作为参数传入，这个值是前面提到的http.DefaultServeMux
    c.isRegisteredOnRoot = c.addHandler(service, c.ServeMux)
  }
  // 将webServices添加到container的webservice列表中
  c.webServices = append(c.webServices, service)
  return c
}

// addHandler
func (c *Container) addHandler(service *WebService, serveMux *http.ServeMux) bool {
  pattern := fixedPrefixPath(service.RootPath())
  // ...
  // 这里的关键函数：serveMux.HandleFunc，是Golang标准包中实现路由的函数
  // go-restful中将路由处理函数统一交给c.dispatch函数，可以看出整个go-restful框架中，最核心的就是这个函数了
  if !alreadyMapped {
    serveMux.HandleFunc(pattern, c.dispatch)
    if !strings.HasSuffix(pattern, "/") {
      serveMux.HandleFunc(pattern+"/", c.dispatch)
    }
  }
  return false
}
```



### 路由分发函数 dispatch
如何由container -> webservice -> handler 实现层级分发？ 
go-restful框架通过serveMux.HandleFunc(pattern, c.dispatch)函数，一边连接了Golang提供的官方http扩展机制，另一边通过一个dispatch实现了路由的分发，这样就不用单独写很多的handler了。

这个函数的核心是c.router.SelectRoute，根据请求找到合适的webservice和route
```go
func (c *Container) dispatch(httpWriter http.ResponseWriter, httpRequest *http.Request) {
  ...
  // 根据请求，找到最合适的webService和route
  // 这个方法后面单独介绍
  func() {
    // ...
    webService, route, err = c.router.SelectRoute(
      c.webServices,
      httpRequest)
  }()
  // ...
  if err != nil {
    // 构造过滤器
    chain := FilterChain{Filters: c.containerFilters, Target: func(req *Request, resp *Response) {
      switch err.(type) {
      case ServiceError:
        ser := err.(ServiceError)
        c.serviceErrorHandleFunc(ser, req, resp)
      }
      // TODO
    }}
    // 运行Container的过滤器
    chain.ProcessFilter(NewRequest(httpRequest), NewResponse(writer))
    return
  }

  // 尝试将router对象转为PathProcessor对象
  // 我们使用的是默认的Container，前面介绍过router默认用的CurlyRouter，
  // SelectRoute的其中一个实现类RouterJSR311，也实现了PathProcessor。所以如果用了RouterJSR311,这里接口转换才能获取到值
  // 而默认的CurlyRouter并没有实现PathProcessor接口，因此这里转换后是空值，会走到下一个if语句
  pathProcessor, routerProcessesPath := c.router.(PathProcessor)
  if !routerProcessesPath {
    // 使用默认的路处理器
    pathProcessor = defaultPathProcessor{}
  }
  // 从request的url请求中抽取参数
  pathParams := pathProcessor.ExtractParameters(route, webService, httpRequest.URL.Path)
  wrappedRequest, wrappedResponse := route.wrapRequestResponse(writer, httpRequest, pathParams)
  // 如果有filter的话，处理将所有的filter添加到filter链中
  if size := len(c.containerFilters) + len(webService.filters) + len(route.Filters); size > 0 {
    // compose filter chain
    allFilters := make([]FilterFunction, 0, size)
    allFilters = append(allFilters, c.containerFilters...)
    allFilters = append(allFilters, webService.filters...)
    allFilters = append(allFilters, route.Filters...)
    chain := FilterChain{Filters: allFilters, Target: route.Function}
    chain.ProcessFilter(wrappedRequest, wrappedResponse)
  } else {
    // no filters, handle request by route
    // 没有filter，通过route直接处理请求
    route.Function(wrappedRequest, wrappedResponse)
  }
}
```

#### 路由选择
前面的dispatch中提到的c.router.SelectRoute的作用是选择合适的webservice和route，这里专门介绍一下。

container中的router属性是一个RouteSelector接口
```go
type RouteSelector interface {
  // SelectRoute根据输入的http请求和webservice列表，找到一个路由并返回
  SelectRoute(
    webServices []*WebService,
    httpRequest *http.Request) (selectedService *WebService, selected *Route, err error)
}
```

go-restful框架中共有两个实现类：

- CurlyRouter
- RouterJSR311
```go
// 选择路由功能
func (c CurlyRouter) SelectRoute(
  webServices []*WebService,
  httpRequest *http.Request) (selectedService *WebService, selected *Route, err error) {
  // 解析url，根据'/'拆分为token列表
  requestTokens := tokenizePath(httpRequest.URL.Path)
  // 根据tokens列表和webservice的路由表做匹配，返回一个最合适的webservice
  detectedService := c.detectWebService(requestTokens, webServices)
  ...
  // 返回webservice中匹配的routes集合
  candidateRoutes := c.selectRoutes(detectedService, requestTokens)
  ...
  // 从前面的list中找到最合适的route
  selectedRoute, err := c.detectRoute(candidateRoutes, httpRequest)
  if selectedRoute == nil {
    return detectedService, nil, err
  }
  return detectedService, selectedRoute, nil
}

// 选择webservice
func (c CurlyRouter) detectWebService(requestTokens []string, webServices []*WebService) *WebService {
  var best *WebService
  score := -1
  for _, each := range webServices {
    // 计算webservice的得分
    matches, eachScore := c.computeWebserviceScore(requestTokens, each.pathExpr.tokens)
    // 返回得分最高的webservice
    if matches && (eachScore > score) {
      best = each
      score = eachScore
    }
  }
  // 将得分最高的webservice返回
  return best
}

// 计算webservice得分
func (c CurlyRouter) computeWebserviceScore(requestTokens []string, tokens []string) (bool, int) {
  if len(tokens) > len(requestTokens) {
    return false, 0
  }
  score := 0
  for i := 0; i < len(tokens); i++ {
    each := requestTokens[i]
    other := tokens[i]
    if len(each) == 0 && len(other) == 0 {
      score++
      continue
    }
    if len(other) > 0 && strings.HasPrefix(other, "{") {
      // no empty match
      if len(each) == 0 {
        return false, score
      }
      score += 1
    } else {
      // not a parameter
      if each != other {
        return false, score
      }
      score += (len(tokens) - i) * 10 //fuzzy
    }
  }
  return true, score
}

// 初选：匹配path，返回一批Route作为备选
func (c CurlyRouter) selectRoutes(ws *WebService, requestTokens []string) sortableCurlyRoutes {
  // 选中的Route存放到sortableCurlyRoutes中
  candidates := make(sortableCurlyRoutes, 0, 8)
  // 遍历webservice下所有的route
  for _, each := range ws.routes {
    // paramCount：正则命中
    // staticCount：完全匹配命中
    matches, paramCount, staticCount := c.matchesRouteByPathTokens(each.pathParts, requestTokens, each.hasCustomVerb)
    // 如果匹配，加入到备选列表中
    if matches {
      candidates.add(curlyRoute{each, paramCount, staticCount}) // TODO make sure Routes() return pointers?
    }
  }
  // 排序备选的route
  sort.Sort(candidates)
  return candidates
}

// 二次筛选：匹配属性等信息。返回最合适的Route
func (c CurlyRouter) detectRoute(candidateRoutes sortableCurlyRoutes, httpRequest *http.Request) (*Route, error) {
  // tracing is done inside detectRoute
  return jsr311Router.detectRoute(candidateRoutes.routes(), httpRequest)
}

// 匹配多个属性是否匹配：method、content-type、accept
func (r RouterJSR311) detectRoute(routes []Route, httpRequest *http.Request) (*Route, error) {
  candidates := make([]*Route, 0, 8)
  for i, each := range routes {
    ok := true
    for _, fn := range each.If {
      if !fn(httpRequest) {
        ok = false
        break
      }
    }
    if ok {
      candidates = append(candidates, &routes[i])
    }
  }
  if len(candidates) == 0 {
    if trace {
      traceLogger.Printf("no Route found (from %d) that passes conditional checks", len(routes))
    }
    return nil, NewError(http.StatusNotFound, "404: Not Found")
  }

  // 判断 http method 是否匹配
  previous := candidates
  candidates = candidates[:0]
  for _, each := range previous {
    if httpRequest.Method == each.Method {
      candidates = append(candidates, each)
    }
  }
  if len(candidates) == 0 {
    if trace {
      traceLogger.Printf("no Route found (in %d routes) that matches HTTP method %s\n", len(previous), httpRequest.Method)
    }
    allowed := []string{}
  allowedLoop:
    for _, candidate := range previous {
      for _, method := range allowed {
        if method == candidate.Method {
          continue allowedLoop
        }
      }
      allowed = append(allowed, candidate.Method)
    }
    header := http.Header{"Allow": []string{strings.Join(allowed, ", ")}}
    return nil, NewErrorWithHeader(http.StatusMethodNotAllowed, "405: Method Not Allowed", header)
  }

  // 判断 Content-Type 是否匹配
  contentType := httpRequest.Header.Get(HEADER_ContentType)
  previous = candidates
  candidates = candidates[:0]
  for _, each := range previous {
    if each.matchesContentType(contentType) {
      candidates = append(candidates, each)
    }
  }
  if len(candidates) == 0 {
    if trace {
      traceLogger.Printf("no Route found (from %d) that matches HTTP Content-Type: %s\n", len(previous), contentType)
    }
    if httpRequest.ContentLength > 0 {
      return nil, NewError(http.StatusUnsupportedMediaType, "415: Unsupported Media Type")
    }
  }

  // 判断 accept 是否匹配
  previous = candidates
  candidates = candidates[:0]
  accept := httpRequest.Header.Get(HEADER_Accept)
  if len(accept) == 0 {
    accept = "*/*"
  }
  for _, each := range previous {
    if each.matchesAccept(accept) {
      candidates = append(candidates, each)
    }
  }
  if len(candidates) == 0 {
    if trace {
      traceLogger.Printf("no Route found (from %d) that matches HTTP Accept: %s\n", len(previous), accept)
    }
    available := []string{}
    for _, candidate := range previous {
      available = append(available, candidate.Produces...)
    }
    return nil, NewError(
      http.StatusNotAcceptable,
      fmt.Sprintf("406: Not Acceptable\n\nAvailable representations: %s", strings.Join(available, ", ")),
    )
  }
  // 如果有多个匹配，返回第一个
  return candidates[0], nil
}
```

## 第三方应用：k8s 

## 参考

- [k8s使用的web框架：go-restful 源码分析](https://cloud.tencent.com/developer/article/1951995)