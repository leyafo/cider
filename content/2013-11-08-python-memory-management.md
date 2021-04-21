# Python内存管理

在 Python 的内部系统中，它的内存管理结构是以金子塔结构呈现的.如下图所示:  
![pymemory-1.png](/images/pymemory-1.png)
- 其中-1和-2这两层是跟操作系统来负责的,这里我们略过不表.  
- 第0层就是我们平常在 C 中使用的 malloc, Python 不会直接使用它,而是会在此基础上做一个内存池.
- 第1层是 Python 自己在基于 malloc 的基础上构造的一个内存池.
- 第2和第3层是基于第1层的.每当 Python 内部需要使用内存时,会使用第1层做好的分配器来分配内存.
因此第1层是 Python 内部管理内存的主要地方.

## 作用
在 C 中如果频繁的调用 malloc 与 free 时,是会产生性能问题的.再加上频繁的分配与释放小块的内存会产生内存碎片. Python 在这里主要干的工作有:
1. 如果请求分配的内存在1~256字节之间就使用自己的内存管理系统,否则直接使用 malloc.
2. 这里还是会调用 malloc 分配内存,但每次会分配一块大小为256k的大块内存.
3. 经由内存池登记的内存到最后还是会回收到内存池,并不会调用 C 的 free 释放掉.以便下次使用.

## 内存池结构

![pymemory-2.png](/images/pymemory-2.png)
如上图所示,整个黑框格子代表内存池(usedpools).每个单元格存放一个双向链表,每个链表的节点是一个大小为4k的内存块.在这个池中,每个单元格负责管理不同的小块内存.为了便于管理,每个单元格管理的内存块总是以8的倍数为单位.以如下代码为例: 
```C
PyObject_Malloc(3)
``` 

这里我们需要一块大小为3个字节的内存.它将定位到管理大小为8字节的单元格.然后返回大小8字节的内存.在这里 usedpools 有一个令人蛋疼的 ticky. usedpools 在初始化时用了如下代码:
```C
#define PTA(x)  ((poolp )((uchar *)&(usedpools[2*(x)]) - 2*sizeof(block *)))
#define PT(x)   PTA(x), PTA(x)
static poolp usedpools[2 * ((NB_SMALL_SIZE_CLASSES + 7) / 8) * 8] = {
    PT(0), PT(1), PT(2), PT(3), PT(4), PT(5), PT(6), PT(7)
    ......
};
```

如上面所示使用了两个一组的指针来初始化 usedpools, 每次定位到单元格他使用的是 usedpools[idx+idx] 这样来定位的.我也不知道它为什么会使用这么蛋疼且令人费解的设计,连注释都这样写着:
>It's unclear why the usedpools setup is so convoluted.

## 分配  
PyObject\_Malloc 函数首先会判断进来申请分配的字节数是否在 1\<x\<256 bytes 这个范围内.然后再从 usedpools  中的管理对应大小的 pool 拿到一块 block, 每个 pool 的大小是4k.每当使用完 pool 中的最后一个 block 时,它会将这个 pool 从 usedpools 剥离出去.
在调用 malloc 获取内存时,这里做了一层缓存(arenas).每次调用 malloc 会分配一块大小为256k的内存,然后将这块内存分解为一块一块大小为4k的 pool,每当 pool 中 block 用完后,就会重新从 arenas 拿一块 pool 并放入到 usedpools.  
在获取空闲 block 时,这里使用了一个 ticky. pool 中的 freeblock 成员是指向一块空闲的 block. 但这个 freeblock 在空闲时,它里面存了一个地址这个地址指向下一块空闲的 block. 下一块空闲的 block 里同样也存放了一个空闲 block 的地址,以此往下推.直到最后的 block 指向 NULL 为止.  
```C
bp = pool->freeblock;
if ((pool->freeblock = *(block **)bp) != NULL) {
	UNLOCK();
	return (void *)bp;
}
```
如上面所示,拿到一块 block 后,直接获取 freeblock 里面存的地址,并指向它.

## 释放
在释放时会判断将要释放的内存是否属于 usedpools 管理.通常情况下它会直接将这块内存放到的 usedpools 对应 pools 中.如果发现这个 pools 中的 block 全部是 free 状态,它将会返到 arenas .如果 arenas 中的所有的 pool 都为 free 状态的话,则会直接调用 C 中的 free 函数将内存归还与操作系统.否则将这块 arenas 链接到 usable\_arenas 正确的位置中.
