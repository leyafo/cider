# 一个 API 设计问题带来的思考
这两天在使用公司其他部门设计的一个 API，在和同事交流的过程中发现 HTTP API 设计的一个很关键的问题点，所以把思考的一些东西记录在这里。      
我的 API server 是公司其他部门上游发送过来的请求，我的服务器接收请求，并响应数据。上游的 API 服务器设计时并未过多的考虑接入的问题，最开始的一个版本是把所有的数据全部放在 URL querystring 里面去给下游的 API server 去解析，不同的 API 通过里面的 `?Action=CreateUser` 这样的参数形式去解析。参数全部放到 querystring 明显会带来以下问题：  
1. 所有的参数都没有类型，全部是 string 类型，其他的数据类型需要再次做解析。    
2. HTTP 有长度限制，无法发送过多的数据。  
3. URL 需要 base64 转义，这又为参数传递多带来一层解码负担。  
基于这些我们看的见的 query string 问题，我们公司上游的 HTTP server 新做了一版接口。这次的做法是把所有的参数全部都放到 HTTP body 里面，这里看似已经解决了 querystring 所有的问题，但根本的问题并没有解决（咱们先不谈为什么不用 RESTful）。  
下面是这个新版本的 HTTP body 的参数形式。
```json
    {
        "Action":"CreateUser",
        "Name":"John",
        "Email":"John@example.com",
        ....
    }
```
这里仍旧事把 Action 参数放到 body 里面，下游的 HTTP 服务器对于发送过来的信息需要首先去解码 body 里面的 Action 参数，然后再通过 Action 参数去判断这是具体属于哪一个 API 的服务。你需要通过一些手段把这个 Action map 到一个具体的入口点。这带来一个问题，就是我们的每个 API 需要的参数都是不一样的，当我们去序列化的时候我们无法为所有 API 都使用同一个数据结构去序列化数据。下面的两个 HTTP body 除了 Action 可以使用同样的方式去序列化以外，其他的的数据结构完全无法通用。
```json
    //json 1
    {
        "Action":"CreateUser",
        "Name":"John",
        "Email":"John@example.com",
        ....
    }

    //json 2
    {
        "Action":"CreateFoo",
        "FooName":"John",
        "FooEmail":"John@example.com",
        ....
    }
```
这就需要让接收参数的 API server 需要单独为每一个 API 的参数单独设定一个数据结构。那么问题就来了，我们在不知道 Action 的情况下，如何知道里面到底要拿哪一个数据结构去序列化？这里比较 trick 的方式就是先不管三七二十一，直接先把 action 序列化出来，再把 body 丢到具体的 API 去再次序列化。很明显，这是非常丑陋且低效的做法，并且你还没有别的方式可选。这就是设计带来的问题，即使代码再漂亮性能也会非常糟糕。  
到这里我们再回过头来看头提到 RESTFul 如何完美的解决这个问题。如果我们拿以上两个 API 用 RESTful 来实现就是如下效果：
```json
POST http://cn.bing.com/user
{
    "Name":"John",
    "Email":"John@example.com",
    ....
}

POST http://cn.bing.com/foo
{
    "FooName":"John",
    "FooEmail":"John@example.com",
    ....
}
```
这样我们通过 method 和 HTTP path 就能指定到对应的 API，也不会存在二次解析这样的问题。在我们不知道一个东西为什么会有各种各样的条条框框和标准时我们最好按照标准去做，否则就最终做出来的东西就非常丑陋。  
接下来请您思考一个问题：为什么 API Server 的返回码应该和 HTTP 保持语义上的一致（如：200 成功，201 创建成功，404 资源不存在），而不要全部 API 统一使用 200 返回或者自己在 body 里定义一套返回码？
