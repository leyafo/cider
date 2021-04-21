# Understanding the dependence in Makefile
## Introduction  

`make` is a simple and useful tool which controls the building of your source code. It can be used in any programming language. It's also can be used for reducing the repetitive tasks. For example, if you have multiple shell tasks that listed one by one. You can use `make` to organize the workflow.

If you've downloaded an open-source project or joined a new project as a developer. To run the entire project, you have to start building it first. Nearly all open-source projects have a `Makefile` for organizing code. But the `Makefile` is not easy to understand. Its official manual is so long and hard to understand. It's difficult to modify a `Makefile` for you want if you're a newbie. Indeed, you can Google the `Makefile` tutorial and then copy&paste some codes within your `Makefile`. Even though you don't know how `Makefile` works, your code running.

This post will take you 10 minutes to understand the `make` core conception. Afterwards, you will create your own `Makefile` to run some basic tasks and figured out how `make` works.

## Core conception

**Dependence** is the `make`'s core conception. Understanding the dependence is the most important principle in `make`. In the `Makefile` every task depends on other prerequisites, which can be files, tasks, or some other form of rules. If you want to run a task, you must define the target and prerequisites. We call this a dependence. A simple `Makefile` consists of "dependencies" with the following shape:
```makefile
foo: bar
```

The above `Makefile` do nothing because the `bar` target is undefined. Target should be depend on a defined prerequisite or be defined how it works. The following target defined to run `echo` command.

```makefile
foo:
    echo "hello world"  #noticed: you must use 'tab' to start this line.
```

## A realistic example

Next, we will give you a realistic example. Suppose we have a simple C project, which contains the following files:

```bash
.
├── http.c
├── http.h
├── main.c
├── socket.c
├── socket.h
```

In our project, we have the following dependent relationships:  
`http.c` include `http.h` and `socket.h`  
`socket.c` include `socket.h`  
`main.c` include `socket.h` and `http.h`.

To build an executing file with `gcc`, we must compile the source code for the object and then linked them to an executing file. Here are our rules:

```Makefile
http.o: http.c http.h socket.h
    gcc -c http.c
socket.o: socket.h
    gcc -c socket.c
main.o: main.c http.h socket.h
    gcc -c main.c
```

Link all object files to an execute file:
```
exe: http.o socket.o main.o
    gcc -o exe http.o socket.o main.o
```

Now, typing `make exe` to build your own project.
