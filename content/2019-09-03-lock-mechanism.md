# 各种锁的原理，你要了解的都在这里。
在并发编程中，锁是非常见的防止 race condition 情况的发生。race condtion 是计算机运行中一种常见的状态。例如我们写下表达式 `i = i+1` 时， CPU 执行会首先把变量 i 从寄存器里取出来，再进行加 1 的操作，最后再把结果放回到寄存器中。这个操作需要三步，会产生资源访问的状态不一致。以下是 `i = i + 1` 并发时产生状态不一致的情况。
```C
T0: producer execute register1 = count {register1 = 5}
T1: producer execute register1 = register1 + 1 {register1 = 6}
T2: consumer execute register2 = count {register2 = 5}
T3: consumer execute register2 = register2 − 1 {register2 = 4}
T4: producer execute count = register1 {count = 6}
T5: consumer execute count = register2 {count = 4}
```
当并发执行以上代码时，我们会得到两个截然不同的结果。这就是 race condition。要解决这个问题我们需要保证两个并发执行的单元最终得到一致的结果，解决方法就是让一个执行单元执行完所有相关后再让另一个执行单元执行。我们把这块执行的单元称作临界区，以下就是 `i = i+1` 的临界区。
```C
T0: producer execute register1 = count {register1 = 5}
T1: producer execute register1 = register1 + 1 {register1 = 6}
T4: producer execute count = register1 {count = 6}
```
一个临界区通常又如下几个特点：  
1. 只能由一个执行单元在里面执行代码。  
2. 一个执行单元要进入临界区必须要等其他的执行单元执行完。 
3. 执行单元不能一直呆在临界区，必须要出来，否者其他的执行单元就无法进入临界区。  
以上的三个特点我们用锁来保证，让一段需要 n 步的操作合并成无可拆分的一步操作。我们把这个操作称为原子操作。  

## 算法锁
为了保证临界区所需要的这三个特性，我们可以由算法来实现一个锁，作为临界区的资源保护用。下面我们有请 Peterson 老哥为我们演示一下如何徒手不靠任何硬件指令撸出一把锁。这位老哥的思路非常清晰，主要就是设定一个标志，然和再设定一个互斥的标志用来防止两个执行单元去争抢一个临界区。下面是 peterson 老师的徒手算法锁。
```C
    int turn; //turn can be 0 or 1 those are execute sectors
    bool flag[2]; 
    while (true) {
        flag[i] = true;
        turn = j;  
        while (flag[j] && turn == j)
            ;
        /* critical section */
        flag[i] = false;
        /*remainder section */
    }
```
以上的算法我们主要关注 flag 和 trun。其中 flag 表示当前谁准备进入，turn 表示当前谁正在临界区里面，主要用于互斥访问。flag 和 turn 保证临界区的互斥访问和等待。最后的 `flag[i]=false` 让临界区的资源让位。

## 硬件指令锁
硬件指令锁主要是 TAS(test and set) 和 CAS(compare and swap)。TAS 还有个名字叫 (TSL) test and lock 它的操作具有原子性，通过在进入临界区之前把一个变量值设为 true 来实现保护临界区的效果。我们单纯的看 TAS 指令并不那么好理解，如果我们把它放到具体的应用场景里就能很清晰的明白 TAS 指令的作用。
```C
    bool test_and_set(bool *target) {
        bool rv = *target;
        *target = true;
        return rv;
    }
    void foo(){
        bool locked;
        while(test_and_set(locked))
            ;
        /* critical secion */
        locked = false;    
    }
```

另一个指令是 CAS(compare and swap)，它通过原子操作交换两个变量的值来达到对变量的修改。我们可以把它看作是 `i = i+1 ` 这个表达式的原子操作版本。它的实现类似如下：
```C
int compare and swap(int *value, int expected, int new_value) {
    int temp = *value;
    if (*value == expected)
        *value = new_value;
    return temp;
}
```
CAS 指令是我们对变量做修改非常常见的，尤其是在需要对变量进行修改的场景，例如：`increment(int *i)` 这样对 i 进行自增的操作种。

接下来我们介绍几种基于 CAS 和 TAS 实现的锁。
# Mutex
在一些线程库里面，mutex 是最简单的一种锁，也是最常见的锁。它的作用主要是对一段临界区上锁，对其他试图访问已经上锁的资源禁止访问。因此 mutex 也叫互斥锁，它的是英文单词 mutual exclustion 的缩写。mutex 可以使用 SAS 或 CAS 来实现。实现的细节主要有两种类型：
1. spin lock。如果一个执行单元锁住了一块资源，另一个执行单元试图进入会一直轮询知道获取到锁为止。它的实现很像我们上面使用 TAS 的场景。
```C
    bool locked;
    void lock(){
        while(test_and_set(locked))
            ;
    }

    void unlock(){
        locked = false
    }
```
2. 另一种类型的 mutex 是通过让试图获取的锁的执行单元进入到一个等待队列里排队，当锁用完了以后再把等待的执行单元拿出来获取锁并继续执行。
```C
    bool locked;
    void lock(){
        if(test_and_set(locked)){
            //把当前的执行单元加入到等待队列
            add_process_to_waitlist(current_pocess)
            sleep()
        }
    }

    void unlock(){
        locked = false
        //唤醒休眠的线程
        wakeup_process()
    }
```
3. 另外 mutex 还有一种非常简单的实现，就是在加锁时关闭中断，这种方式在单核心的系统中式有效的，对于多核心的操作系统来说是无效的，因为两个核心并不依赖中断并发。
```C
    void lock(){
        //disable interrupts;
    }

    void unlock(){
        //enable interrupts;
    }
```

# Semaphores
semaphore 和 mutex 非常像。他们唯一的不同之处就在于 semaphore 可以让多个执行单元进入临界区，而 mutex 只能让一个单元进入临界区。
```C
    int S
    void lock(){
        while( S <= 0)
            ;  //busy wait
    }

    void unlock(){
        S++
    }
```
由于以上的实现的等待属于空转的忙等待，因此一般的 semphore 会像 mutex 那样加入一个 wait list，让等待的执行单元进入等待，解锁时再唤醒等待的执行单元。
```C
typedef struct {
    int value;
    struct process *list;
} semaphore;

void lock(semaphore *S) {
    S->value--;
    if (S->value < 0) {
    //add this process to S->list;
    sleep();
    }
}

void unlock(semaphore *S){
    S->value++;
    if (S->value <= 0) {
    //remove a process P from S->list;
    wakeup(P);
    }
}
```
我们可以看到，如果我们把 sempahore 里面的 value 换成一个 bool 的二元值就是一个 mutex 的实现。
