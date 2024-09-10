<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [RESTClient 基本使用](#restclient-%E5%9F%BA%E6%9C%AC%E4%BD%BF%E7%94%A8)
  - [加载配置文件并生成config对象](#%E5%8A%A0%E8%BD%BD%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6%E5%B9%B6%E7%94%9F%E6%88%90config%E5%AF%B9%E8%B1%A1)
  - [创建一个RESTClient流程](#%E5%88%9B%E5%BB%BA%E4%B8%80%E4%B8%AArestclient%E6%B5%81%E7%A8%8B)
  - [执行请求](#%E6%89%A7%E8%A1%8C%E8%AF%B7%E6%B1%82)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->


## RESTClient 基本使用

### 加载配置文件并生成config对象
```go
config, err := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")
```

### 创建一个RESTClient流程
```go
restClient, err := rest.RESTClientFor(config)
```

```go
func RESTClientFor(config *Config) (*RESTClient, error) {
	//  这里会限制 GroupVersion（其实是ContentTypeConfig中的，这里是用来序列化请求参数为指定的gv）不能为空
	if config.GroupVersion == nil {
		return nil, fmt.Errorf("GroupVersion is required when initializing a RESTClient")
	}
	// 限制NegotiatedSerializer（其实是ContentTypeConfig中的，这里是用来序列化请求参数）不能为空
	if config.NegotiatedSerializer == nil {
		return nil, fmt.Errorf("NegotiatedSerializer is required when initializing a RESTClient")
	}

	// Validate config.Host before constructing the transport/client so we can fail fast.
	// ServerURL will be obtained later in RESTClientForConfigAndClient()
	_, _, err := defaultServerUrlFor(config)
	if err != nil {
		return nil, err
	}

	httpClient, err := HTTPClientFor(config)
	if err != nil {
		return nil, err
	}

	return RESTClientForConfigAndClient(config, httpClient)
}

// http客户端
func HTTPClientFor(config *Config) (*http.Client, error) {
	//  用来获取传输层（没错，就是七层网络协议中的传输层）表示对象
	transport, err := TransportFor(config)
	if err != nil {
		return nil, err
	}
    // 用来接收创建的rest client
	var httpClient *http.Client
	if transport != http.DefaultTransport || config.Timeout > 0 {
        // 不是默认的，则会创建一个包装了transport的http.Client对象
		httpClient = &http.Client{
			Transport: transport,
			Timeout:   config.Timeout,
		}
	} else {
		httpClient = http.DefaultClient
	}

	return httpClient, nil
}
```
更具体的信息创建
```go
func RESTClientForConfigAndClient(config *Config, httpClient *http.Client) (*RESTClient, error) {
    // ...

	baseURL, versionedAPIPath, err := defaultServerUrlFor(config)
	if err != nil {
		return nil, err
	}

	// 获取请求速率限制器
	rateLimiter := config.RateLimiter
	if rateLimiter == nil {
		qps := config.QPS
		if config.QPS == 0.0 {
			qps = DefaultQPS
		}
		burst := config.Burst
		if config.Burst == 0 {
			burst = DefaultBurst
		}
		if qps > 0 {
			// 创建一个基于令牌桶的速率限制器
			rateLimiter = flowcontrol.NewTokenBucketRateLimiter(qps, burst)
		}
	}

	var gv schema.GroupVersion
	if config.GroupVersion != nil {
		gv = *config.GroupVersion
	}
	clientContent := ClientContentConfig{
		//  指定客户端可以接受的类型。
		AcceptContentTypes: config.AcceptContentTypes,
		// 如果未设置 AcceptContentTypes，则此值将设置为对服务器发出的请求的 Accept 标头，并设置为发送到服务器的任何对象的默认内容类型。
		// 注意 ：如果未设置，则使用“application/json”。
		ContentType:        config.ContentType,
		GroupVersion:       gv,
        // 用于获取多种支持的媒体类型的编码器和解码器。
		Negotiator:         runtime.NewClientNegotiator(config.NegotiatedSerializer, gv),
	}

	restClient, err := NewRESTClient(baseURL, versionedAPIPath, clientContent, rateLimiter, httpClient)
	if err == nil && config.WarningHandler != nil {
		restClient.warningHandler = config.WarningHandler
	}
	return restClient, err
}


// NewRESTClient 创建一个新的 RESTClient。此客户端在指定的路径上执行通用 REST 功能，例如 Get、Put、Post 和 Delete。
func NewRESTClient(baseURL *url.URL, versionedAPIPath string, config ClientContentConfig, rateLimiter flowcontrol.RateLimiter, client *http.Client) (*RESTClient, error) {
  // 如果contentType没有设置，则设置默认为"application/json"
	if len(config.ContentType) == 0 {
		config.ContentType = "application/json"
	}

	base := *baseURL
  // 判断是否以/结尾，否则追加
	if !strings.HasSuffix(base.Path, "/") {
		base.Path += "/"
	}
  // 设置queryParams为空
	base.RawQuery = ""
  // 设置base url对应的分段（用于定位页面位置，在url的#字符后）为空
	base.Fragment = ""

	return &RESTClient{
        // base 是客户端所有调用的根 URL
		base:             &base,
		// versionedAPIPath 是将base URL 连接到资源根的路径（在base url后面追加了group version）
		versionedAPIPath: versionedAPIPath,
		// 描述了 RESTClient 如何编码请求body和解码响应。
		content:          config,
        // 创建传递给请求的 BackoffManager。（用来控制request请求和apiserver交互如果出现异常的处理，简单说就是如果出现异常，就会sleep指数的时间间隔，然后向下执行）
		createBackoffMgr: readExpBackoffConfig,
        // rateLimiter 在此客户端创建的所有请求之间共享。也就是所有与apiserver交互的request的速率限制
		rateLimiter:      rateLimiter,
        // 与apiserver交互的基础。如果未设置，则会设置http.DefaultClient。
		Client: client,
	}, nil
}

```

### 执行请求
```go
	if err := restClient.
		Get().
		Namespace("kube-system").
		Resource("pods").
		VersionedParams(&metav1.ListOptions{Limit: 500}, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result); err != nil {
		panic(err)
	}
```
Get请求
```go
// Get begins a GET request. Short for c.Verb("GET").
func (c *RESTClient) Get() *Request {
	return c.Verb("GET")
}
func (c *RESTClient) Verb(verb string) *Request {
    return NewRequest(c).Verb(verb)
}
```
创建一个新的请求对象，用于访问服务器上的 runtime.Objects。
```go
func NewRequest(c *RESTClient) *Request {
	var backoff BackoffManager
	if c.createBackoffMgr != nil {
		backoff = c.createBackoffMgr()
	}
    // 如果backoff为空，则会赋值noBackoff（不做任何操作）
	if backoff == nil {
		backoff = noBackoff
	}

	var pathPrefix string
	if c.base != nil {
		pathPrefix = path.Join("/", c.base.Path, c.versionedAPIPath)
	} else {
		pathPrefix = path.Join("/", c.versionedAPIPath)
	}
    // 设置request超时时长
	var timeout time.Duration
	if c.Client != nil {
		timeout = c.Client.Timeout
	}

	r := &Request{
		c:              c,
		rateLimiter:    c.rateLimiter,
		backoff:        backoff,
		timeout:        timeout,
		pathPrefix:     pathPrefix,
		retry:          &withRetry{maxRetries: 10},
		warningHandler: c.warningHandler,
	}

	switch {
	case len(c.content.AcceptContentTypes) > 0:
		r.SetHeader("Accept", c.content.AcceptContentTypes)
	case len(c.content.ContentType) > 0:
		r.SetHeader("Accept", c.content.ContentType+", */*")
	}
	return r
}
```
```go
// Resource 设置要访问的资源 ([ns/<namespace>/]<resource>/<name>)
func (r *Request) Resource(resource string) *Request {
	if r.err != nil {
		return r
	}
	if len(r.resource) != 0 {
		r.err = fmt.Errorf("resource already set to %q, cannot change to %q", r.resource, resource)
		return r
	}
	if msgs := IsValidPathSegmentName(resource); len(msgs) != 0 {
		r.err = fmt.Errorf("invalid resource %q: %v", resource, msgs)
		return r
	}
	r.resource = resource
	return r
}

// 命名空间将命名空间范围应用于请求 ([ns/<namespace>/]<resource>/<name>)
func (r *Request) Namespace(namespace string) *Request {
	if r.err != nil {
		return r
	}
	if r.namespaceSet {
		r.err = fmt.Errorf("namespace already set to %q, cannot change to %q", r.namespace, namespace)
		return r
	}
	if msgs := IsValidPathSegmentName(namespace); len(msgs) != 0 {
		r.err = fmt.Errorf("invalid namespace %q: %v", namespace, msgs)
		return r
	}
	r.namespaceSet = true
	r.namespace = namespace
	return r
}
=
// VersionedParams 使用隐式(r.c.content中的gv) RESTClient API 版本和默认参数编解码器将提供的对象序列化为 map[string][]string，然后将这些作为参数添加到请求中。
// 使用它来提供客户端库的版本化查询参数。VersionedParams 不会写入设置了 omitempty 且为空的查询参数。
func (r *Request) VersionedParams(obj runtime.Object, codec runtime.ParameterCodec) *Request {
	return r.SpecificallyVersionedParams(obj, codec, r.c.content.GroupVersion)
}

func (r *Request) SpecificallyVersionedParams(obj runtime.Object, codec runtime.ParameterCodec, version schema.GroupVersion) *Request {
	if r.err != nil {
		return r
	}
  // 序列化obj为对应groupVersion的obj，并将obj转化为map（会舍弃掉obj中设置了 omitempty 且为空的属性）
	params, err := codec.EncodeParameters(obj, version)
	if err != nil {
		r.err = err
		return r
	}
	for k, v := range params {
		if r.params == nil {
			r.params = make(url.Values)
		}
		r.params[k] = append(r.params[k], v...)
	}
	return r
}
```
格式化并执行请求。返回一个 Result 对象以便于处理响应
```go
func (r *Request) Do(ctx context.Context) Result {
	var result Result
	err := r.request(ctx, func(req *http.Request, resp *http.Response) {
		result = r.transformResponse(resp, req)
	})
	if err != nil {
		return Result{err: err}
	}
	if result.err == nil || len(result.body) > 0 {
		metrics.ResponseSize.Observe(ctx, r.verb, r.URL().Host, float64(len(result.body)))
	}
	return result
}
```

```go
// 和Watch类似，只是返回不同的对象，内部处理逻辑相似
// 请求连接到apiserver并在收到apiserver的响应时调用提供的函数fn. 它处理重试行为(retry)和请求的预先验证（requestPreflightCheck）. 它最多会调用一次fn

func (r *Request) request(ctx context.Context, fn func(*http.Request, *http.Response)) error {
	//Metrics for total request latency
	start := time.Now()
	defer func() {
		metrics.RequestLatency.Observe(ctx, r.verb, *r.URL(), time.Since(start))
	}()

	if r.err != nil {
		klog.V(4).Infof("Error in request: %v", r.err)
		return r.err
	}

	if err := r.requestPreflightCheck(); err != nil {
		return err
	}

	client := r.c.Client
	if client == nil {
		client = http.DefaultClient
	}

	// Throttle the first try before setting up the timeout configured on the
	// client. We don't want a throttled client to return timeouts to callers
	// before it makes a single request.
	if err := r.tryThrottle(ctx); err != nil {
		return err
	}

	if r.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.timeout)
		defer cancel()
	}

	isErrRetryableFunc := func(req *http.Request, err error) bool {
		// "Connection reset by peer" or "apiserver is shutting down" are usually a transient errors.
		// Thus in case of "GET" operations, we simply retry it.
		// We are not automatically retrying "write" operations, as they are not idempotent.
		if req.Method != "GET" {
			return false
		}
		// For connection errors and apiserver shutdown errors retry.
		if net.IsConnectionReset(err) || net.IsProbableEOF(err) {
			return true
		}
		return false
	}

	// Right now we make about ten retry attempts if we get a Retry-After response.
	for {
		//  生成一个http.Request对象并配置必要属性，用来和apiserver交互
		req, err := r.newHTTPRequest(ctx)
		if err != nil {
			return err
		}

		if err := r.retry.Before(ctx, r); err != nil {
			return r.retry.WrapPreviousError(err)
		}
		// http.client.Do 执行request请求，并返回请求响应
		resp, err := client.Do(req)
        // 更新url对应的metrics指标
		updateURLMetrics(ctx, r, resp, err)
		// The value -1 or a value of 0 with a non-nil Body indicates that the length is unknown.
		// https://pkg.go.dev/net/http#Request
		if req.ContentLength >= 0 && !(req.Body != nil && req.ContentLength == 0) {
			metrics.RequestSize.Observe(ctx, r.verb, r.URL().Host, float64(req.ContentLength))
		}
		r.retry.After(ctx, r, resp, err)

		done := func() bool {
			defer readAndCloseResponseBody(resp)

			// if the the server returns an error in err, the response will be nil.
			f := func(req *http.Request, resp *http.Response) {
				if resp == nil {
					return
				}
				fn(req, resp)
			}

			if r.retry.IsNextRetry(ctx, r, req, resp, err, isErrRetryableFunc) {
				return false
			}

			f(req, resp)
			return true
		}()
		// 表示已经执行完成
		if done {
			return r.retry.WrapPreviousError(err)
		}
	}
}

```

处理返回的数据到结构化API对象中
```go
func (r *Request) transformResponse(resp *http.Response, req *http.Request) Result {
	var body []byte
	if resp.Body != nil {
		//  读取body中的数据
		data, err := ioutil.ReadAll(resp.Body)
		switch err.(type) {
		case nil:
			// 为空 表示200 接收数据
			body = data
		case http2.StreamError:
			// 表示有请求错误（为什么是http2.StreamError，说明transport中执行的http.client是使用的http2依赖）
			// This is trying to catch the scenario that the server may close the connection when sending the
			// response body. This can be caused by server timeout due to a slow network connection.
			// TODO: Add test for this. Steps may be:
			// 1. client-go (or kubectl) sends a GET request.
			// 2. Apiserver sends back the headers and then part of the body
			// 3. Apiserver closes connection.
			// 4. client-go should catch this and return an error.
			klog.V(2).Infof("Stream error %#v when reading response body, may be caused by closed connection.", err)
			streamErr := fmt.Errorf("stream error when reading response body, may be caused by closed connection. Please retry. Original error: %w", err)
			return Result{
				err: streamErr,
			}
		default:
			klog.Errorf("Unexpected error when reading response body: %v", err)
			unexpectedErr := fmt.Errorf("unexpected error when reading response body. Please retry. Original error: %w", err)
			return Result{
				err: unexpectedErr,
			}
		}
	}

	glogBody("Response Body", body)

	// 验证content type是否准确
	var decoder runtime.Decoder
	contentType := resp.Header.Get("Content-Type")
	if len(contentType) == 0 {
		contentType = r.c.content.ContentType
	}
	if len(contentType) > 0 {
		var err error
		// 解析contentType
		mediaType, params, err := mime.ParseMediaType(contentType)
		if err != nil {
			return Result{err: errors.NewInternalError(err)}
		}
		decoder, err = r.c.content.Negotiator.Decoder(mediaType, params)
		if err != nil {
			// if we fail to negotiate a decoder, treat this as an unstructured error
			switch {
			case resp.StatusCode == http.StatusSwitchingProtocols:
				// no-op, we've been upgraded
			case resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusPartialContent:
				return Result{err: r.transformUnstructuredResponseError(resp, req, body)}
			}
			return Result{
				body:        body,
				contentType: contentType,
				statusCode:  resp.StatusCode,
				warnings:    handleWarnings(resp.Header, r.warningHandler),
			}
		}
	}

	switch {
	case resp.StatusCode == http.StatusSwitchingProtocols:
		// no-op, we've been upgraded
	case resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusPartialContent:
		// calculate an unstructured error from the response which the Result object may use if the caller
		// did not return a structured error.
		retryAfter, _ := retryAfterSeconds(resp)
		err := r.newUnstructuredResponseError(body, isTextResponse(resp), resp.StatusCode, req.Method, retryAfter)
		return Result{
			body:        body,
			contentType: contentType,
			statusCode:  resp.StatusCode,
			decoder:     decoder,
			err:         err,
			warnings:    handleWarnings(resp.Header, r.warningHandler),
		}
	}

	return Result{
		body:        body,
		contentType: contentType,
		statusCode:  resp.StatusCode,
		decoder:     decoder,
		warnings:    handleWarnings(resp.Header, r.warningHandler),
	}
}

```