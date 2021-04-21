# go 语言错误处理
go 语言错误处理非常简单清晰且有效。你只需要记住一个原则：“Whatever you do, always check your errors!”。这条原则很简单，我们只要在有 error 值返回的函数外处理所有的 error 即可。但这带来一个问题，我们的代码里面会大量充斥着如下代码：
```go
func Foo() (err error){
    err = Bar()
    if err != nil{
        return err
    }
    err = Bar()
    if err != nil{
       return err
    }
    err = Bar()
    if err != nil{
       return err
    }
    //....
}
```
对于这个问题，rob pike 在他的那篇 errors are values 文章里建议是写一个同样的接口去包装返回 error 的接口，然后每次重复调用时就不用再反复的检查 error 值。类似如下的方式
```go
func Foo() error{
    var err error
    func newBar(){
        if err != nil{
           return
        }
        err = Bar()
        return 
    }
    newBar()
    newBar()
    newBar()
    ...
    return err
}
```
这个方式有效且简单，省略非常多繁琐的 `if err != nil` 的判断。（请忽略函数调用开销）。实际上 bufio 里面的 Scan 接口就是这么实现的。如果我们的函数里面分别调用不同的函数，并且它们除 error 以外的返回值是不同的类型，这个方法就会失效。因此新版的 golang 里面的 [error handle](https://dev.to/deanveloper/go-2-draft-error-handling-3loo) 给出如下类似 try catch 的实现机制：
```go
func ParseJson(name string) (Parsed, error) {
    handle err {
        return fmt.Errorf("parsing json: %s %v", name, err)
    }

    // Open the file
    f := check os.Open(name)
    defer f.Close()

    // Parse json into p
    var p Parsed
    check json.NewDecoder(f).Decode(&p)

    return p
}
```
它就是为了解决频繁检查 error 而发明的一个机制。它与 try catch 不一样的是只将错误检查固定在函数的内部并只针对 error 类型做判断。这种方式更像是一个语法糖，至于以后会不会加入到 golang 里面我们拭目以待吧。

在 golang 的标准库里面，我们能看到 errors 这个库，其中里面有一个 `New(text string) error` 这样的接口，它帮助我们构造一个错误，并以传入的 text 做为错误信息。而 fmt 里面也有一个 `Errorf(format string, a ...interface{}) error` 接口，以 fmt 形式帮助我们构造一个错误。这里我们会误以为这两个接口实现的功能是重叠的，看起来 errors.New 的接口好像没有必要存在一样。如果你与我有同样疑惑时，证明你并没有完全理解 golang 里面的错误处理机制。即使 rob pike 一直在不同的地方反复强调 "errors are values" 这条原则，但 golang 里面的 error 是有可以类型的。golang 没有 try catch 那样复杂的异常类型判断，为了解决现实中我们需要对不同类型做出不同处理的问题，我需要使用 error.New 构造一个错误类型。  
这种应用场景典型的就是 io.Reader 这个接口。当我们需要判断一个数据 stream 是否已经读到 EOF(end of File) 的状态时，我们需要为其构造一个 `errors.New("EOF")` 的错误接口，如果没有这个错误类型标志 io.Reader 接口无法判断数据是否已经读到结束。  
我们可以通过 reader 接口的实现想象一下如果没有 EOF 这个错误标志。我们该如何判断数据已经读完？
```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

func (f Foo)Reader(p []byte)(n int, err error){
    //已经没有数据可读
}
```
试想一下，上面的 `reader` 这个接口已经没有数据可读，但我们需要让外界知道这个信息，我们该如何返回？
1. 如果返回 `n = -1` 做为标志，那么是不是表示 n 同时作为一个状态值返回？如果这样，n 这个返回值就会产生二义性。  
2. 如果返回 `EOF` 标志，它并不是一个错误状态，而是一个标志状态。
以上两种方式并不 100% 完美，但第二种方式比第一种方式更符合逻辑直觉。  
如果我们在 pkg 层面暴露一个错误的标志给外界作为判断，使用这个错误标志的用户会需要依赖我们的 pkg 才能使用，并且是强依赖。因此我们在设计模块时需要考虑把错误标志做为 API 接口的一部分。为了只将错误类型与我们的模块关联，我们可以选择不内部包装 error 标志，不暴露给外界去做错误类型判断。
```go
pkg foo
fooError = errors.New("foo")
func IsFooError(err error)bool{
    _, ok = err.(fooError)
    return ok
}
```
现在我们可以很清晰的知道，`errors.New` 是用来构造一个错误类型，`fmt.Errorf` 只是用来构造一个错误信息。在构造一个类型信息的时候它的错误值是固定的，因此只能用 `errors.New` 来进行构造。从这里我们可以看出 `errors.New` 只是一个常量的 string 值，它是在编译时确定的，`fmt.Errorf` 是一个运行时确定的变量值。现在我们还是回到了 "errors are values" 这一原则。

提到错误处理我们不能忽略 panic 和 restore 这一对函数。它俩长得非常像 try catch，看起来也是互相配对使用，这也是我们会让我们的 panic 形成错误的理解。在 golang 里面 panic 表示无法恢复的失败，比如空指针访问，程序奔溃，数组越界，这样严重的错误。restore 可以捕获程序的 panic 错误。在严格的意义上来说，我们不应该去恢复 panic 的错误。restore 的作用是为了记录 panic 的现场信息而准备的，它的作用更像是黑匣子的作用。另外对于需要长期运行不能停的服务端应用，我们可以使用 restore 来程序局部的 bug 造成的以外奔溃。比如我们可以在 http 的入口处去设置 restore 函数，确保当某一个 API 发生错误时不影响其他的 API 正常运行。  
记住： **panic 和 restore 并不是成对使用，也不能用来捕获错误。**

参考文章：
[Don’t just check errors, handle them gracefully](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)
[Errors are values](https://blog.golang.org/errors-are-values)
