# go 并发
go 语言的并发非常简洁，只要写一个 goroutine 就可以了。

```go
go func(){
  println("foo")
}()
```

goroutine 会立即返回，这个匿名函数会并发执行。

如果需要和 goroutine 传递变量可以使用 go channels。从传递的角度来说，channels 可以分为发送和接收。使用 *<-* 符号作为传递，*channel <- a* 表示把变量 a 发送到 channel 里， *<- channel* 表示接收 *channel* 过来的数据。

从 channels 的属性角度来说，go channels 的类型分为两种：1.buffered channel. 2.Unbufered channel. 从字面上理解这两者的关系就是 buffer 和非 buffer，这个 buffer 可以理解为数组。如果只需要传递单个变量使用 unbuffered channel 就行了。

下面的代码是 unbufered channels 的传递

```go
a := make(chan int)
go func(){
  a <- 1
}()
println(<-a)
//do another things
```

unbufered channels 一个重要特点就是它是同步的，上面代码中 channel a 会一直等待到数据 1 发送过来才会继续往下执行。

channels 也可以像 Linux 管道一样，作为输入输出在不同的 goroutine 之间传递。

```go
a := make(chan int)
b := make(chan int)
go func() {
	a <- 1
}()

go func() {
	a1 := <-a
	b <- a1 + 1
}()
println(<-b)
```

bufered channels 是一个序列，大小是预先分配好的。每当这个序列有空余空间的时候，发送端的 goroutine 可以往里面发送数据。而接收端从序列里面拿掉一个数据，这个时候会空出一个空间出来。作为发送端如果 buffer 占满，程序会 block 住，直到有剩余空间为止。而接收端会在 buffer 为空的时候 block 住，直到有数据进入到 buffer 为止。

```go
package main
import "time"
func main() {
	buf := make(chan int, 3)
	go func() {
		for i := 0; ; i++ {
			buf <- i
			time.Sleep(100 * time.Millisecond)
		}
	}()
	go func() {
		for {
			println(<-buf)
		}
	}()
	time.Sleep(10 * time.Second)
}
```

上面的代码两个 goroutine 会同步一直跑，每当 buf 里面有数据就会直接 print 出来。

channels buffer 和普通的 slice 是有很大的区别的。channels 内部会和 goroutine 的 schedule 相连，如果 channels 在同一个 goroutine 里面做 receive 和 send，会互相等待的从而产生死锁。所以 go channel 一定只能作为 goroutine 间的信息传递，不要拿去做普通的变量或 buffer 用。

我们有时候需要在一个 goroutine 里面响应多个 channel 事件，对于不同的 channel 需要有不同的操作。这时候我们可以使用 select 来对 channel 做类似多路复用的事情。

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan int, 1)
	go func() {
		for i := 0; i < 10; i++ {
			ch <- i
			fmt.Printf("send value %d to channel\n", i)
		}
	}()

	break_timing := time.After(5 * time.Second)
loop:
	for {
		select {
		case x := <-ch:
			fmt.Printf("received value %d\n", x)
		case <-break_timing:
			fmt.Println("break now!")
			break loop
		default:
			fmt.Println("main goroutine are sleeping")
			time.Sleep(500 * time.Millisecond)
		}
	}
}
```

如上所示，select 会随机选一个被触发的事件执行。default 是当没有任何的事件需要响应的时候执行。select 的行为很像 switch case 子句。select 还有另外的一个作用就是可以防止被 channel block 住。上面的代码不会一直等 channel 过来的值，它只会在 channel 有响应的时候才会执行对应的 case。

由于 goroutine 是异步的，创建完成后会立即返回。我们没法非常准确的知道 goroutine 是什么时候执行完成的，尤其是在创建多个 goroutine 的情况下。这个时候我们就需要用到 sync 的 waitgroup 来判断 gouroutine 何时执行完成。

```go
package main
import "sync"
func main() {
  var wg sync.WaitGroup 
  for i:=0; i<10; i++{
    wg.Add(1)
    go func() {
	   defer wg.Done()
      //do something
	}()
  }
  wg.Wait();
  println("all goroutine are finshed.")
}
```

如上面代码所示，waitgroup 在 goroutine 创建前会增加一个计数，在完成时会调用 Done 减少一个计数。Wait() 会一直等到计数为零为止。Add 函数传入的是一个 int 值，也就是说可以为负数，这个细节可以忽略掉。因为 Done 函数内部调用的就是 Add(-1)。所以不要随意的往 Add 函数里面传参，配对使用 Add(1), Done() 就好。 
