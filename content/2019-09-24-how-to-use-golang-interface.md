# golang interface 要怎么用？

## 面向对象的思考
我们学习面向对象经常讨论的三大毒瘤是：封装，继承，多态。我们先来分析和定义这三大特性，最后再来介绍为什么 golang 一个 interface 就能实现所有的这三大特性。  

### 封装
我们先来思考封装的本质是什么？封装的本质就是把你的代码做成一个可装卸的组件，组件暴露一些接口，用户使用组件时毋需关注内部的实现，只关心接口的输入输出就可以把这个组件集成到代码里面。为了把这一系列的过程更加便捷的描述出来，我们可以把它命名成一个具体的对象，对象里面包含一系列操作。

### 继承，多态
golang 里面实际上没有继承这个概念。golang 里面把一个 struct 包含到另一个 struct 辅以一定的编译的语法糖看起来像继承一样的抽象实际上不是继承，它还是封装这个概念的升级版。在 golang 里面把这叫做**组合式继承**。  
继承和多态并不能各自分开独立的来看。多态如果脱离继承它将变成另一抽象的概念：duck type。  
在其他面向对象的语言里面我们继承一个实体，就拥有父类的所有方法。如果需要对父类同一方法进行覆盖，我们就实现一个同样的方法进行覆盖。这样的目的是让我们为两个不同的对象调用同一个方法。但实现这一点我们为什么需要继承？如果一个对象有一个我们需要的方法，那么我们并不需要关心这具体是个什么对象。这就是 duck type 的概念（如果一个物体会像鸭子一样叫，那么它就是一个鸭子）。 dock type 相对于继承式的多态要更加灵活。如下代码所示：
```golang
import (
	"fmt"
)

type Speaker interface{
   Speak()
}
type Bird struct{

}
func (b *Bird)Speak(){
    fmt.Println("zzzzzzzz")
}
type People struct{
}
func (p *People)Speak(){
    fmt.Println("hello world")
}

func main() {
	var speakers []Speaker
	speakers = append(speakers, new(Bird), new(People))
	for _, s := range(speakers){
		s.Speak()
	}
}
```
如上所示我们不需要关心 bird 和 people 是否有同一个共同的父类，或者是否都继承自 Object 这个祖先类。我们不需要关心对象，只需要关心是否能方便的把代码组织到一起。封装实际上是所有编程语言的基础要素，它本质上就是**代码复用**和你的代码是否**面向对象**没有任何的关系。而 interface 是我们代码抽象的一种方式，它只关心我们要做到事情，并不关心它属于一个什么**对象**。所以在 golang 里面我们没有 **has-a** 和 **is-a** 这两个概念的困惑。在 golang 里面网络传输，文件读写，编解码用 reader 和 writer 这两个接口抽象起来非常自然。

## interface{}
golang 经常被人诟病的一点就是它没有泛型，但泛型反对者们说 ```interface{}``` 就是泛型，可以拿来做为泛型用。这种说法是不对的。 ```interface{}``` 它不是泛型，它是一个接口，这个接口可以适配任何类型。它与 void 又不太一样，它会保存传入对象的 value 和 type 信息。  
使用 ```interface{}``` 的时候我们重点需要关注的怎样传递 value，而不是用来抽象接口的设计。当一个接口没有明确的输入输出的信息，那这是个糟糕的接口，它会让使用接口的人非常困惑。

## interface receiver
在 struct 上 bind 一个 function 有 pointer 和 value 的区别。在调用的时候没有区别，但在 interface 的实现上，需要明确的指定传递的对象是一个 pointer 还是非 pointer。这也是使用 interface 有点迷惑的地方。如下代码所示：
```golang
import (
	"fmt"
)

type Speaker interface{
   Speak()
}

type People struct{
   Age int
}
func (p *People)Speak(){
    fmt.Println("hello world")
}

func main() {
	p1 := new(People)
	p1.Speak()
	p2 := &People{}
	p2.Speak()

	var speakers []Speaker
	//Can't compile
	//speakers = append(speakers, new(People), People{})
	speakers = append(speakers, new(People), &People{})
	for _, s := range(speakers){
		s.Speak()
	}
}
```

## interface 现实应用案例分析
从 java 转过来的同学容易对 interface 进行滥用，在不必要使用 interface 的地方使用 interface。如下代码就是一位写 java 的同学转过来写 golang 的代码。
```golang
    type RemoteClienter interface{
        func Start(addr string)
        func Send([]byte)(int, error)
        func AsyncSend([]byte)(int, error)
    }
```
这段代码的目的是抽象不同的 client，它可以通过 http，tcp，或者其他协议进行同步或异步网络数据发送。单纯的从这个目的推断这个接口的实现是没有问题的。但我们为什么要去抽象一个 client？我们为什么需要关心接口的数据发送是同步还是异步的？这个 client 我们实际上只需要关心数据是否可以发出去，至于用什么协议，同步还是异步发送这些问题根本不重要。因此这个接口更适合用 golang 的 writer 接口抽象。如下所示：
```golang
    type Wirter interface{
        func Write([]byte)(int, error)
    }
```
