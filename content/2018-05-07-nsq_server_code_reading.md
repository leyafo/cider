# NSQ 服务端源码剖析
NSQ 服务端代码我们可以先从主模块 nsqd 的 main 函数开始看起，这里是整个程序的入口。在一个 nsqd 启动前会有一些配置和并调用 nsqd.LoadMetadata() 和 nsqd.PersistMetadata() 来加载消息和持久化消息。为了关注主要的流程，我们这里可以忽律掉这些我们并不需要太过于关心的部分。

在 main 函数里面，会分别启动一个 httpserver 和 TCPSever，它们俩做的实际上是同样的事情，这里我们可以只关心 TCPServer 的流程即可。

NSQ 服务端的每个 TCP 连接都是长连接，一旦建立，数据会不断的通过 Socket 发送和接收数据。启动好 TCP 监听后，通过不断的轮询接受 client 的连接。建立好连接后通过版本号初次握手，拒绝掉一切不符合协议的客户端。服务端会为每一个进来的连接创建一个 goroutine handler。

以下是整个服务端主要的 TCP 流程。
```go
// nsq.go in main function
tcpListener, err := net.Listen("tcp", n.getOpts().TCPAddress)
n.Lock()
n.tcpListener = tcpListener
n.Unlock()
tcpServer := &tcpServer{ctx: ctx}
n.waitGroup.Wrap(func() {
	protocol.TCPServer(n.tcpListener, tcpServer, n.logf)
})
// internal/protocol/tcp_server.go
func TCPServer(listener net.Listener, handler TCPHandler, logf lg.AppLogFunc) {
	for {
		clientConn, err := listener.Accept()
		go handler.Handle(clientConn)
	}
}
// nsqd/tcp.go
func (p *tcpServer) Handle(clientConn net.Conn) {
	buf := make([]byte, 4)
	_, err := io.ReadFull(clientConn, buf)
	protocolMagic := string(buf)

	var prot protocol.Protocol
	switch protocolMagic {
	case "  V2":
		prot = &protocolV2{ctx: p.ctx}
	default:
        //拒绝其他任何一切不符合协议的连接
		protocol.SendFramedResponse(clientConn, frameTypeError, []byte("E_BAD_PROTOCOL"))
		clientConn.Close()
		return
	}
 	//...
	err = prot.IOLoop(clientConn)
}


```

每个连接的 goroutine handle 里面都会调用 IOLoop，这个函数是很典型的 TCP 的 Reader 和 Writer 的工作方式。这个函数本身就是一个 reader for 循环去不断的读取客户端发过来的数据，并序列化成对应的指令去执行。 

```go
// nsqd/protocol_v2.go
func (p *protocolV2) IOLoop(conn net.Conn) error {
	var line []byte

	clientID := atomic.AddInt64(&p.ctx.nsqd.clientIDSequence, 1)
	client := newClientV2(clientID, conn, p.ctx)

	messagePumpStartedChan := make(chan bool)
	go p.messagePump(client, messagePumpStartedChan) 
	<-messagePumpStartedChan //等待 Pump 函数初始化好

	for {
        //...
		line, err = client.Reader.ReadSlice('\n')
		//...
		params := bytes.Split(line, separatorBytes)

		var response []byte
		response, err = p.Exec(client, params)  //执行指令
	}
	//...
}
```

IOLoop 里面 messagePump 函数实际上是一个TCP write 函数，它把消息写入到 Socket 里面，这个函数叫 Pump 的意思是客户端接受的消息都来自这里。

```go
func (p *protocolV2) messagePump(client *clientV2, startedChan chan bool) {
	//...
	flushed := true
	close(startedChan) //通知 Reader 可以开始读取消息了。
	for {
		//...
		select {
		//...
		case msg := <-memoryMsgChan:
            //...
			subChannel.StartInFlightTimeout(msg, client.ID, msgTimeout)
			client.SendingMessage()
			err = p.SendMessage(client, msg)
			if err != nil {
				goto exit
			}
			flushed = false
		case <-client.ExitChan:
			goto exit
		}
	}
}
```

messagePump 这个函数的数据主要来自于 go-channel，它实际上就是将内存数据写入到网络中。

## Topic

在详细介绍 Topic 之前我们先简单的了解一下 Topic 的基本概念以及它在 nsq 里面扮演的角色以及它与 channel 之间的关系（为了不混淆概念，这篇文章里的 channel 对应的是 nsq 的 channel，golang 里面的 channel 全部以 go-channel 代替）。

一个 nsqd 存在着多个 topic，每个 topic 有多个 channel。每个 channel 接收来自 topic 的拷贝，消息可以多路发送，每个消息放到一个 channel 里面，多个订阅者订阅一个 channel 可分布式接收消息，实现负载均衡。

我们先来看看 Topic  的构造

```go
func NewTopic(topicName string, ctx *context, deleteCallback func(*Topic)) *Topic {
	t := &Topic{
	//...
	}

	if strings.HasSuffix(topicName, "#ephemeral") {
		t.ephemeral = true
		t.backend = newDummyBackendQueue()
	} else {
		//...
		t.backend = diskqueue.New(
			//...
		)
	}

	t.waitGroup.Wrap(func() { t.messagePump() })
	t.ctx.nsqd.Notify(t)  //通知 nsqlookupd 注册这个 topic
	return t
}
```

这里 Topic 会创建一个 queue 做为消息的备份用，如果 topic name 里面含有 #ephemeral，表示不备份这个 topic 的消息。newDummyBackendQueue 后面实际上就是一系列空操作，消息一旦进入这里面就会完全丢弃。随后启动一个 messagePump 用来接收消息并处理（每个 topic 都会有一个 messagePump 不断轮询）。

```go
func (t *Topic) messagePump() {
	t.RLock()
	for _, c := range t.channelMap {
		chans = append(chans, c)
	}
	t.RUnlock()
	for {
		select {
		case msg = <-memoryMsgChan:  //t.memoryMsgChan
		case buf = <-backendChan:    //t.backend
			msg, err = decodeMessage(buf)
			//...
		case //...
		case <-t.exitChan:
			goto exit
		}

		for i, channel := range chans {
			chanMsg := msg
			// copy the message because each channel
			// needs a unique instance but...
			// fastpath to avoid copy if its the first channel
			// (the topic already created the first copy)
			if i > 0 {
				chanMsg = NewMessage(msg.ID, msg.Body)
				chanMsg.Timestamp = msg.Timestamp
				chanMsg.deferred = msg.deferred
			}
			if chanMsg.deferred != 0 {
				channel.PutMessageDeferred(chanMsg, chanMsg.deferred)
				continue
			}
			err := channel.PutMessage(chanMsg)
		}
	}
}
```

上面我们可以看到，消息的主要来源有两个地方 memoryMsgChan 和 backendChan 。每当收到消息后会投递到当前 topic 下面所有的 channel（这就是为什么两个consumer 订阅同一个消息，不同的 channel 都会收到同一条消息的原因）。如果只有一个 channel 消息会直接投递，多个 channel 会拷贝一份再投递。

我们先来看看 backendChan 的消息是如何过来的。diskqueue 里面会维护一个读写文件的 ioLoop，这里面会接收一条消息并写到文件或者从文件里读一条消息并投递出去。

```go
func (d *diskQueue) ioLoop() {
	var dataRead []byte
	var err error
	var count int64
	var r chan []byte

	syncTicker := time.NewTicker(d.syncTimeout)

	for {
		//...
		if (d.readFileNum < d.writeFileNum) || (d.readPos < d.writePos) {
			dataRead, err = d.readOne()
			//...
			r = d.readChan
		} else {
			r = nil
		}

		select {
		case r <- dataRead:  //把消息投递出去
			count++
			// moveForward sets needSync flag if a file is removed
			d.moveForward()
		case dataWrite := <-d.writeChan:   //把消息写到文件
			count++
			d.writeResponseChan <- d.writeOne(dataWrite)
		case <-d.exitChan:
			goto exit
		}
	}
}
```

我们再来看看 TCP 数据流过来的消息是怎样进到 topic 里面来的。

```go
func (p *protocolV2) IOLoop(conn net.Conn) error {
	//...
    response, err = p.Exec(client, params)
    //...
}
func (p *protocolV2) Exec(client *clientV2, params [][]byte) ([]byte, error) {
    switch{
    //...
	case bytes.Equal(params[0], []byte("PUB")):
		return p.PUB(client, params)
    }
}
func (p *protocolV2) PUB(client *clientV2, params [][]byte) ([]byte, error) {
    //...
    topic := p.ctx.nsqd.GetTopic(topicName)
	msg := NewMessage(topic.GenerateID(), messageBody)
	err = topic.PutMessage(msg)
    //...
}
// PutMessage writes a Message to the queue
func (t *Topic) PutMessage(m *Message) error {
	//...
	err := t.put(m)
	//...
	return nil
}
func (t *Topic) put(m *Message) error {
	select {
	case t.memoryMsgChan <- m:
	default:
        //....
		err := writeMessageToBackend(b, m, t.backend)
		bufferPoolPut(b)
		t.ctx.nsqd.SetHealth(err)
	}
	return nil
}

```

以上我们可以看到，消息从 TCP 过来后，经过解码后获取全局的 nsqd 实例找到对应的 topic 进行投递。topic 会找到它下面的 channel 投递（topic 的 pump message 会轮询 memoryMsgChan ）， 如果失败会保存到 backend 里面（就是 diskqueue）。	另外就是 flush 操作也会把 memoryChan 里的消息写到 disk 里面。

## Channel

一个 Channel 里面的消息有两种类型，一种是延迟消息，另一种是普通的消息。我们在 nsq 的 client 接口可以看到两种不同消息有不同的 publish 接口。每种消息分别用一个 hash 表和一个优先队列存储在内存中。下文会详细提到两种不同数据结构的作用。

以下是 channel 内部的数据结构。

```go
type Channel struct {
	// 64bit atomic vars need to be first for proper alignment on 32bit platforms
	requeueCount uint64   //重新入队的消息条目数
	messageCount uint64   //进入 Channel 的消息总数
	timeoutCount uint64	  //超时的消息数

	sync.RWMutex		  //全局 channel 读写锁

	topicName string      //每个 channel 都和一个 topic 相关联
	name      string      //channel 名字
	ctx       *context    //全局 nsq 实例变量

	backend BackendQueue  //消息备份队列

	memoryMsgChan chan *Message   //内存队列，消息进入 channel 的入口。
	exitFlag      int32				//退出标志
	exitMutex     sync.RWMutex		//退出信号

	// state tracking
	clients        map[int64]Consumer    //因为不需要关注生产者，这里的 clients 全是 consumer.
	paused         int32				 //暂停接收消息
	ephemeral      bool					 //不备份标志
	deleteCallback func(*Channel)		 //删除 channel 回调
	deleter        sync.Once			 //删除锁

	// Stats tracking
	e2eProcessingLatencyStream *quantile.Quantile   //状态报告

	// TODO: these can be DRYd up
	deferredMessages map[MessageID]*pqueue.Item   //延迟消息
	deferredPQ       pqueue.PriorityQueue		  //延迟消息队列
	deferredMutex    sync.Mutex					  //延迟队列锁
	inFlightMessages map[MessageID]*Message		  //待发送的消息
	inFlightPQ       inFlightPqueue				  //待发送的队列
	inFlightMutex    sync.Mutex					  //发送队列锁
}
```

Channel 只负责接收数据并发出去，消息发出去前会备份到队列，出错会备份到文件。它只做消息的临时的存储地，不做任何其他的事情。所有它内部没有 goroutine 循环做轮询的事情。下面这张官方的图是描述 channel 的形态，这里的 goroutine 是指其他模块里面的 goroutine 循环。

![img](/images/nsq_server_channle.png)

在 nsq 与客户端建立连接后，每个客户端通过 sub 指令与 nsq 建立一条 channel 通道，每条消息通过 pub 指令投递进来，进来后会通过全局的 nsq instance 找到这条消息的 topic 并投递进来。从上面的 topic 的工作流我们可以看到消息最终还是进入到对应的 channel 里面。在 protocol 里面的 pummessage 这个 goroutine 里面通过从对应的客户端拿出 channel（在 sub 指令时建立的），并从 channel 里面的 memoryMsgChan 这个 go channel 里面拿消息出来，并最终通过投递给远程的 client。

以下是这一系列流程的代码：

```go
func (p *protocolV2) SUB(client *clientV2, params [][]byte) ([]byte, error) {
	var channel *Channel
	for {
        //...
		topic := p.ctx.nsqd.GetTopic(topicName) //获取对应的 topic
		channel = topic.GetChannel(channelName) //获取 topic 下面的 channel
		channel.AddClient(client.ID, client)
		//...
		break
	}
	//...
	client.Channel = channel
	// update message pump
	client.SubEventChan <- channel
}
func (p *protocolV2) messagePump(client *clientV2, startedChan chan bool) {
    //...
    var memoryMsgChan chan *Message
    //...
    for{
        //...
    	memoryMsgChan = subChannel.memoryMsgChan
        //...
        select {
        //...
        case subChannel = <-subEventChan:
        case msg := <-memoryMsgChan:
            //发送前先放到队列里面去。
            subChannel.StartInFlightTimeout(msg, client.ID, msgTimeout)
            client.SendingMessage()
            err = p.SendMessage(client, msg)
            if err != nil {
                goto exit
            }
            flushed = false
        }
    }
}
```

消息发给客户端后，并不会丢弃，因为有可能会发送失败，从上面的 StartInFlightTimeout 函数我们就能知道这一点。StartInFlightTimeout 就涉及了前面提到的同时用 hash 表和优先队列存储消息。这里使用优先队列的原因是因为在消息重发的机制里面需要根据时间上的优先级重新发送消息。另外就是消息需要 client 返回 FIN 指令才能删除并丢弃掉。 nsq 每当收到一条消息后会马上投递出去，或者是备份到文件里面（channel 和 topic 的Putmessage 都有对应的实现）。

这里有一点比较迷惑的是 pqueue.PriorityQueue 和 inFlightPqueue 是一摸一样的队列，功能上的实现也是一样的，为什么要用两个，我也要去问问官方是怎么回事。

以下是 StartInFlightTimeout 具体实现

```go
func (c *Channel) StartInFlightTimeout(msg *Message, clientID int64, timeout time.Duration) error {
	//...
	msg.pri = now.Add(timeout).UnixNano() //这里的 now 就是消息往客户端开始发送的时间点
	err := c.pushInFlightMessage(msg)	  //写到 hash 表
	if err != nil {
		return err
	}
	c.addToInFlightPQ(msg)				 //写到优先队列
	return nil
}
```

我们可以通过  FIN 指令看看如何从队列删除一条消息。

```go
//p.Exec(client, params)
switch {
	case bytes.Equal(params[0], []byte("FIN")):
		return p.FIN(client, params)
        //....
    }
//p.FIN(client, params)
client.Channel.FinishMessage(client.ID, *id)
//channel.FinishMessage
func (c *Channel) FinishMessage(clientID int64, id MessageID) error {
	msg, err := c.popInFlightMessage(clientID, id)
    //...
	c.removeFromInFlightPQ(msg)
	//...
}

func (c *Channel) popInFlightMessage(clientID int64, id MessageID) (*Message, error) {
	c.inFlightMutex.Lock()
	msg, ok := c.inFlightMessages[id]
	//...
	delete(c.inFlightMessages, id)
	c.inFlightMutex.Unlock()
	return msg, nil
}

func (c *Channel) removeFromInFlightPQ(msg *Message) {
	c.inFlightMutex.Lock()
	c.inFlightPQ.Remove(msg.index)
	c.inFlightMutex.Unlock()
}
```



## 清理过期消息

接下来我们在回到 NSQ 的 Main 里面看看另一个重要的 goroutine - queueScanLoop。如果你刚刚上手开始看 nsq 的源码，深入这个函数去读会有一些迷惑，因为它前面有一个复杂 **Redis's probabilistic expiration** 算法，这个算法我现在也解释不清楚，我们先跳过就把它当成一个定时器好了。

queueScanLoop 里面会每隔一段时间获取所有的 channel 并调用 resizePool，这里的 Pool 指的就是 channel 里面的消息队列。resizePool 实际上做的就是根据超时时间重新发送 channel 里面的 inFlightMessage 和 DeferredMessage 里面的消息。

```go
func (n *NSQD) queueScanLoop() {
    //...
    n.resizePool(len(channels), workCh, responseCh, closeCh)
    //...
}
func (n *NSQD) resizePool(num int, workCh chan *Channel, responseCh chan bool, closeCh chan int) {
    n.waitGroup.Wrap(func() {
        n.queueScanWorker(workCh, responseCh, closeCh)
    })
}
func (n *NSQD) queueScanWorker(workCh chan *Channel, responseCh chan bool, closeCh chan int) {
	for {
		select {
		case c := <-workCh:
			now := time.Now().UnixNano()
			dirty := false
			if c.processInFlightQueue(now) {
				dirty = true
			}
			if c.processDeferredQueue(now) {
				dirty = true
			}
			responseCh <- dirty
		case <-closeCh:
			return
		}
	}
}

func (c *Channel) processInFlightQueue(t int64) bool {
	c.exitMutex.RLock()
	defer c.exitMutex.RUnlock()
	dirty := false
	for {
		c.inFlightMutex.Lock()
		msg, _ := c.inFlightPQ.PeekAndShift(t) //根据超时拿出队列里面的消息。
		c.inFlightMutex.Unlock()

		if msg == nil {
			goto exit
		}
		dirty = true

		_, err := c.popInFlightMessage(msg.clientID, msg.ID) //删除 hash 表里面的消息
		if err != nil {
			goto exit
		}
		atomic.AddUint64(&c.timeoutCount, 1)
		c.RLock()
		client, ok := c.clients[msg.clientID]
		c.RUnlock()
		if ok {
			client.TimedOutMessage()
		}
		c.put(msg) //把消息投递出去。
	}

exit:
	return dirty 
}
//c.processDeferredQueue(now) 和 processInFlightQueue 做的事情是一样的。
```

这里要注意的是，如果客户端始终不回复消息是否处理完成，那么这个消息会不断的出队入队重复循环。



## 小结

以上就是整个 NSQ 服务端的源码分析，在这里可以稍微总结下它的工作流。

nsq server 启动监听一个端口，并且经此与 client 建立一个 TCP 长连接。每个连接都会有一个 goroutine 轮询 TCP 过来的消息，这时会为这个连接和客户端建立一个 ioloop，里面有两个独立的 goroutine 分别负责读和写。消息通过 pub 指令进入 nsq，其对应的 topic 会负责转发到相应 channel 的 memeryMsgChan(go-channel) 里面。每个 client 的 ioloop 通过 channel 的 memeryMsgChan(go-channel) 拿到消息并发出去。当客户端处理完消息回应 Fin 指令过来后，原来进入队列备份的消息就会被丢弃掉。



