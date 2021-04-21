# nsq client 源码剖析
NSQ 客户端实现与编程语言无关，它通过 TCP 与 nsq 连接。主要实现的功能分为 5 大块，分别是：connection 管理, message 处理，producer, resumer, 并发控制。由于 producer 的流程比较简单，只是单纯的发把消息发送到 nsq，在这里不做描述。本文基于 [go-nsq](https://github.com/nsqio/go-nsq) 源码分析除 producer 外其他 4 大块的主要代码流程。

## Connection 模块

nsq 客户端主要分为两块：producer, resumer. 它们都使用 connection 模块与 nsq 连接。以下是 NewConn 初始化的主要结构。

```go
	return &Conn{
		addr: addr,  //服务器地址

		config:   config,  //基本的服务端配置
		delegate: delegate, //producer 与 resumer 用来处理消息回调的接口。

		maxRdyCount:      2500,   //并发控制数量
		lastMsgTimestamp: time.Now().UnixNano(),  //上一条消息抵达的时间

		cmdChan:         make(chan *Command),	  //与服务器的 cmd 管道
		msgResponseChan: make(chan *msgResponse),  //消息回应管道
		exitChan:        make(chan int),		   //退出信号
		drainReady:      make(chan int),		   //中断信号
	}
```

以上并不是完整 Conn 结构体字段，为了简便，我们现在只需要关注这几个字段。producer 和 resumer 会调用 Connect() 函数与 nsq 建立连接。以下是建立连接的过程，为了便于理解，本文截取的代码都忽略了一些无关紧要的细节。

```go
func (c *Conn) Connect() (*IdentifyResponse, error) {
	//...
	conn, err := dialer.Dial("tcp", c.addr)
	//...
	c.conn = conn.(*net.TCPConn)
	c.r = conn
	c.w = conn

	_, err = c.Write(MagicV2)   //与服务器通信的版本号
	//...

	resp, err := c.identify()
	
    //...
    c.wg.Add(2)
	//...
	go c.readLoop()
	go c.writeLoop()
	return resp, nil
}
```

在与 NSQ 建立连接时会首先发送 identify 指令告诉服务器自己这边的一些配置。如 clientid, hostname, 心跳间隔时长，超时时长等一些配置信息。在建立连接后会分别启动一个 readloop 和 writeloop 的 goroutine 与服务器通信。以下是 readloop 的基本处理流程。

```go
	for {
	    //...
		frameType, data, err := ReadUnpackedResponse(c)
        //....
		if frameType == FrameTypeResponse && bytes.Equal(data, []byte("_heartbeat_")) {
			c.delegate.OnHeartbeat(c)
			err := c.WriteCommand(Nop())
			//...
			continue
		}

		switch frameType {
		case FrameTypeResponse:
			c.delegate.OnResponse(c, data)
		case FrameTypeMessage:
			msg, err := DecodeMessage(data)
			//...
			c.delegate.OnMessage(c, msg)
		case FrameTypeError:
			//...
		default:
			//...
		}
	}

```

如果消息是 heartbeat 会回复一条空指令告诉服务器自己的存活状态。（heartbeat 发送的间隔时长在建立 identify 指令时传入的）。producer 与 resumer 会在调用 NewConn 函数时传入一个自己的 delegate 用来给 connection 模块回调处理消息。

我们再来看看 writeloop 做的事情。

```go
	for{
		select {
		case <-c.exitChan:
			//...
			goto exit
		case cmd := <-c.cmdChan:
			err := c.WriteCommand(cmd)
			//...
		case resp := <-c.msgResponseChan:
			// Decrement this here so it is correct even if we can't respond to nsqd
			msgsInFlight := atomic.AddInt64(&c.messagesInFlight, -1)

			if resp.success {
				c.log(LogLevelDebug, "FIN %s", resp.msg.ID)
				c.delegate.OnMessageFinished(c, resp.msg)
				c.delegate.OnResume(c)
			} else {
				c.log(LogLevelDebug, "REQ %s", resp.msg.ID)
				c.delegate.OnMessageRequeued(c, resp.msg)
				if resp.backoff {
					c.delegate.OnBackoff(c)
				} else {
					c.delegate.OnContinue(c)
				}
			}

			err := c.WriteCommand(resp.cmd)
			//...
		}
	}
```

writeloop 主要关心的是发送 cmd 指令，和回复消息处理的状态。在 nsq 客户端中一个消息的处理状态有四种。分别是：

1.FIN 处理成功，告诉服务端可以放心的丢弃掉这条消息。

2.REQ 处理失败，告诉服务端这条消息需要重新入队。

3.TOUCH 告诉服务端需要更多的时间处理这条消息。

4.消息处理超时，服务端会根据消息处理时长判断消息是否需要重新入队。

这里为了消息处理更高效，使用了一个单独的 channel 发送 FIN 和 REQ 状态指令。

## Message 处理流程

在 consumer 模块内，有一个 handleloop 不断轮询 incomingMessages 管道来接收 connection 模块发过来的消息。connection 通过注册进来的 delegate 回调 OnMessage 相关的接口，OnMessage 会把 message 发送到 incomingMessages 管道里面。最后 handlerLoop 轮询 incomingMessages 管道获取消息，通过调用用户注册的 handler 接口来回调处理消息。以下是简单的代码流程。

```go
c.delegate.OnMessage(c, msg) //conn.go:530

//delegates.go:112
func (d *consumerConnDelegate) OnMessage(c *Conn, m *Message){
    d.r.onConnMessage(c, m) 
}

r.incomingMessages <- msg  //consumer.go:648 in onConnmessage function

//consumer.go:1106 in handlerLoop function
for {
	message, ok := <-r.incomingMessages
    //...
	err := handler.HandleMessage(message)
	//...
}
```

## Resumer

resumer 模块是整个 nsq client 的根本，夸张一点的说，整个 nsq client 就是围绕着 resumer 模块在做文章。这里为了讲解清晰我把 resumer 的并发控制放到下一段讲，这段主要讲解释 resumer 与 nsq 如何建立连接，如何接受来自多个 nsq 的消息。

每个 resumer 会订阅一个 topic 和 channel（此 *channel* 非彼 *channel*）。一个 consumer 仅会与一个 topic 关联。 resumer 通过连接 nsqlookup 去获取对应的 nsq 地址并与之建立连接。也就是说一个 consumer 会建立多条连接去连不同的 nsq。我们可以通过看 consumer 和 nsqlookup 之间的基本流程得知这一点。

```go
func (r *Consumer) ConnectToNSQLookupd(addr string) error {
	// ....
	r.lookupdHTTPAddrs = append(r.lookupdHTTPAddrs, addr)
	numLookupd := len(r.lookupdHTTPAddrs)
	r.mtx.Unlock()

	// if this is the first one, kick off the go loop
	if numLookupd == 1 {
		r.queryLookupd()
		r.wg.Add(1)
		go r.lookupdLoop()
	}

	return nil
}
```

在这里仅仅启动了一个 goroutine 去连接 nsqlookup， 由于 lookup 并不是一个需要频繁请求的资源，不需要启动多个 goroutine 去轮询。 多个 nsqlookup 会放到 lookupdHTTPAddrs 这个 slice 里面，在 lookupdLoop 轮询时会随机挑选一个没有被请求过的，以下代码我们可以看到这一点。

```go
func (r *Consumer) nextLookupdEndpoint() string {
	r.mtx.RLock()
	if r.lookupdQueryIndex >= len(r.lookupdHTTPAddrs) {
		r.lookupdQueryIndex = 0
	}
	addr := r.lookupdHTTPAddrs[r.lookupdQueryIndex]
	num := len(r.lookupdHTTPAddrs)
	r.mtx.RUnlock()
	r.lookupdQueryIndex = (r.lookupdQueryIndex + 1) % num
    //...
}
```

lookupdLoop 这个函数会每隔一段时间调用 queryLookupd，在这个函数里面调用 HTTP API 去获取到对应的 nsq 地址，最后调用 ConnectToNSQD 与 nsq 建立新的连接并订阅消息。以下是代码流程。

```go
//lookupdLoop
ticker = time.NewTicker(r.config.LookupdPollInterval)
for {
    select {
    case <-ticker.C:
        r.queryLookupd()
	case <-r.lookupdRecheckChan:
        r.queryLookupd()
    case <-r.exitChan:
        goto exit
    }
//queryLookupd
for _, addr := range nsqdAddrs {
	err = r.ConnectToNSQD(addr)
	//...
}
//ConnectToNSQD
conn := NewConn(addr, &r.config, &consumerConnDelegate{r})
resp, err := conn.Connect()
cmd := Subscribe(r.topic, r.channel)
err = conn.WriteCommand(cmd)
```

以上是 consumer 与 nsq 连接的基本流程。如果你不调用 ConnectToNSQLookupd 去与 nsq 建立连接，那么 consumer 仅仅只会与一个 nsq 连接。

## 并发控制

接下来就是整个 nsq client 的重头戏，并发控制。这也是整个客户端最难理解的部分。我们先来通过 nsq 的文档了解一些基本的情况。

1. consumer 有一个 maxInFlight 变量用来标识可以处理的最大消息数量。totalRdyCount 变量用来计算可接收消息数量，它永远不会大于 maxInFlight，每当收到一个消息后会对其减 1。
2. connection 的 rdyCount 用来标识当前 connection 可接收的消息数量，如果收到一个消息就对其减 1。它自己也有一个 maxInFlight 变量标识正在处理的消息数量。
3. 客户端会在连接的 identify 配置时传入一个 MaxRdyCount，服务端对每个 connection 发送的消息永远不超过这个数字。
4. consumer 会维护一个 RDY 状态，每次消息处理完后更新 RDY 状态，如果数量过大，会重新将一些 connection 的 RDY 数字初始化为 0。

为了便于理解 consumer.totalRdyCount,  conn.rdyCount 和 maxInFlight 我们可以通过一个消息的处理流程看看 conn 和 resumer 是如何维护这几个变量的。

```go
//in conn.readLoop function
case FrameTypeMessage:
    atomic.AddInt64(&c.rdyCount, -1)
    atomic.AddInt64(&c.messagesInFlight, 1)
    atomic.StoreInt64(&c.lastMsgTimestamp, time.Now().UnixNano())

    c.delegate.OnMessage(c, msg)

func (r *Consumer) onConnMessage(c *Conn, msg *Message) {
	atomic.AddInt64(&r.totalRdyCount, -1)
	atomic.AddUint64(&r.messagesReceived, 1)
	r.incomingMessages <- msg
	r.maybeUpdateRDY(c)
}

//in conn.writeLoop function
case resp := <-c.msgResponseChan:
msgsInFlight := atomic.AddInt64(&c.messagesInFlight, -1)

//in consumer.onConnMessage function
atomic.AddInt64(&r.totalRdyCount, -1)
//in consumer.updateRDY function
rdyCount := c.RDY()
maxPossibleRdy := int64(r.getMaxInFlight()) - atomic.LoadInt64(&r.totalRdyCount) + rdyCount
//in consumer.redistributeRDY function
maxInFlight := r.getMaxInFlight()
availableMaxInFlight := int64(maxInFlight) - atomic.LoadInt64(&r.totalRdyCount)
//in consumer.sendRDY function
atomic.AddInt64(&r.totalRdyCount, -c.RDY()+count)
```

从以上代码我们可以看出 messagesInFlight 相当于对一个消息进行入队出队操作。conn 的 totalRdyCount 表示可处理消息的数量，每当处理一个消息，这个数量就会减 1。consumer 的 totalRdyCount 包含的是所有 connection（consumer 与 connection 是一对多的关系）的可处理 Rdy 的数量。

我们可以通过看 resumer 的 redistributeRDY()  这个函数来观察并发如何控制的 。

```go
func (r *Consumer) redistributeRDY() {
	//...
	//找到一些处理时间比较长的 connection，并对其 RDY 置为 0，让它不要再接收新的消息。
	possibleConns := make([]*Conn, 0, len(conns))
	for _, c := range conns {
		lastMsgDuration := time.Now().Sub(c.LastMessageTime())
		lastRdyDuration := time.Now().Sub(c.LastRdyTime())
		rdyCount := c.RDY()
		if rdyCount > 0 {
			if lastMsgDuration > r.config.LowRdyIdleTimeout {
				r.updateRDY(c, 0)
			} else if lastRdyDuration > r.config.LowRdyTimeout {
				r.updateRDY(c, 0)
			}
		}
		possibleConns = append(possibleConns, c)
	}

    //计算空余可以分配的的 RDY 数量
	availableMaxInFlight := int64(maxInFlight) - atomic.LoadInt64(&r.totalRdyCount)
	if r.inBackoff() {
		availableMaxInFlight = 1 - atomic.LoadInt64(&r.totalRdyCount)
	}
	//重新随机分配空余的 RDY
	for len(possibleConns) > 0 && availableMaxInFlight > 0 {
		availableMaxInFlight--
		r.rngMtx.Lock()
		i := r.rng.Int() % len(possibleConns)
		r.rngMtx.Unlock()
		c := possibleConns[i]
		// delete
		possibleConns = append(possibleConns[:i], possibleConns[i+1:]...)
		r.updateRDY(c, 1)
	}
}
```

LastMessageTime 是上一条消息处理前设定的时间。

lastRdyTimestamp 是这个 connection 上一次调用 SetRdyCount 的时间。

这两个时间的对比主要是为了找出那些处理消息比较慢的连接，尽量让其他的 connection 去接收消息。

在 onConnMessage 这个函数里面，我们可以看到 maybeUpdateRDY 这个函数，这表示每当处理一条消息，都有可能重新进行并发控制计算。以下是这个函数的主要流程：

```go
//...
remain := conn.RDY()
lastRdyCount := conn.LastRDY()
count := r.perConnMaxInFlight()

// refill when at 1, or at 25%, or if connections have changed and we're imbalanced
if remain <= 1 || remain < (lastRdyCount/4) || (count > 0 && count < remain) {
    r.updateRDY(conn, count)
}
//...
```



## 小结

nsq 客户端并不是太复杂，开发一个客户端只需要处理好连接，做好消息接收，以及最重要的并发控制即可。在业务层面的代码，我们需要着重关注每个任务的执行时长，以及根据机器的性能调整并发控制。
