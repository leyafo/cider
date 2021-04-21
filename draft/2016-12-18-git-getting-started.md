#  "git 初级教程
## Introduce 

git 不是一个对新手友好的工具，大多数新手第一次面对 git 总会有点手无足措。这时当你开始怀疑是不是一个蠢货时请不要停止，继续保持怀疑。面对一大堆命令和魔法般的黑色 terminal 大多数新手都会手无足措，不知道该怎么去用 git 去提交自己的代码。更可怕的事情是，如果搞砸了还会导致你旁边的豹子头同事把你给撕了的危险。本文档的主要作用是就是可以让你顺利的把代码提交到 git 库，避免产生意外的人身伤害。



## Example

在被我忽悠 git 后你可能会偷偷的跑去网上找 git 相关的教程读。或者在看到我贱指如飞、神乎奇技地在键盘上狂按一通的时候你偷偷的拿笔在旁边记录我敲过的顺序。作为一个踏坑无数，身经百战的老司机有必要在这时告诉你一点儿人生经验，你这么做并没有什么卵用！！！因为你碰到的问题和网上那些教程说的是不一样的，老司机和你看到的世界也是不一样的（隔壁老王用的撩妹技能你直接拿去用会有什么后果？）。因此我们现在以平常工作中最简单的一个应用场景来解释 git 的基本用法。

现在假设有两个农民工张大锤，王大锤，以下简称老张和老王。老张和老王每天搬砖太累了，他们决定写一部电影剧本来改变世界。但两个人在同一个剧情时间点上写同一件事情的时候总是会产生冲突。老张本来在 1970 年10月24日把男主角写挂了。但老王由于刚在淘宝买了块机械键盘，打字打得太爽了。在1970年10月24日这一天写成男主角正在吃着火锅唱着歌。这个时候冲突产生了，老王和老张为了这个事情从相亲相爱变成了砖角遇到爱。下面我们开始详细描述这一起恶性暴力的流血事件。



## Commit

张大锤和王大锤已经写好了一些剧情了。他们的最开始的共同成果如下：

```bash
 1	从前有个人，名字叫赵铁柱。他人长得帅，又多金，还会写代码。
```

最前面那个 1 表示第一行。

老张刚和前女友分手，伤心难过，悲痛欲绝，以致于影响了日常严肃文学创作。所以他把接下来的剧本改成了这样：

```bash
 1	从前有个人，名字叫赵铁柱。他人长得帅，又多金，还会写代码。
 2	后来他死了。
```

写完后老张开始提交代码，他先输入 git status 查看 git 库给出的文件改动信息。(git status 是一个人畜无害的命令，你敲破桌子也不会有人理你)。

```bash
[老张]$ git status                                                                      
On branch master
Your branch is up-to-date with 'origin/master'.
Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
  (use "git checkout -- <file>..." to discard changes in working directory)

	modified:   scenario

no changes added to commit (use "git add" and/or "git commit -a")
```



这里给出的信息我们可以得到如下信息：

1.现在所在的分支在 master 上。（你可以先不用管什么是分支）

2.改动的文件是 scenario。

3.提示显示如果你要提交这个改动先使用 git add 和 git commit -a。

这里解释下 git add 和 git commit 这两个命令，这两个命令是要合用的。git add 是添加改动，git commit 是提交改动。git commit -a 表示的是添加并提交。现在看不懂 git commit -a 这个命令没关系，只要使用按照流程先用 git add 再用 git commit 就可以了。

老张按照流程输入了如下的命令进行 commit。

```bash
[老张]$ git add scenario
[老张]$ git commit -m "自古空情多余恨，此恨绵绵无绝期。"
```

其中 -m 后面双引号括起来的那一串表示你对提交的改动信息说明，主要为了让别人能快速理解你的修改。

老张输入了 git status 查看 git 库给出的信息。

```bash
[老张]$ git status                                                                       
On branch master
Your branch is ahead of 'origin/master' by 1 commit.
  (use "git push" to publish your local commits)
nothing to commit, working directory clean
```

这个时候提示我们可以往远程的 git 库进行提交了。

于是老张使用了 git push origin master (你只需要知道 origin 表示仓库，master 表示分支就可以了)

```bash
[老张]$ git push origin master
Counting objects: 3, done.
Delta compression using up to 8 threads.
Compressing objects: 100% (2/2), done.
Writing objects: 100% (3/3), 342 bytes | 0 bytes/s, done.
Total 3 (delta 1), reused 0 (delta 0)
To git@data-git.mana.com:yafo_bitch/git-tutorial.git
   84179ac..060cd28  master -> master
```

以上信息告诉我们提交到远程库成功了，master -> master 表示远程的与本地的 master 同步了。

张大锤的工作到此为止，幸福生活依然还很幸福。但以下事情开始让张大锤觉得生活并不是那么的一帆风顺。



## Conflict

王大锤鬼混了一天后开始改邪归正为改变世界写剧本了。他也从如下状态的剧本开始写。

```bash
 1	从前有个人，名字叫赵铁柱。他人长得帅，又多金，还会写代码。
```

老王自从用上了淘宝买的机械键盘后，打字快如狗。他使用了他那把逼格甚高的上等无刻 HHKB 添加如下的剧情：

```bash
 1	从前有个人，名字叫赵铁柱。他人长得帅，又多金，还会写代码。
 2	他最大的爱好是吃着火锅，听着 beatles 与高富帅们谈笑风声。
```

老王也按照流程开始提交他本地的改动

```bash
[老王]$ git add scenario
[老王]$ git commit -m "两情若是久长时，一枝红杏出墙来。"
```

然后他把本地的改动 push 到远程的 git 库。但是 push 不上去了，提示信息告诉他与远程的 git 库冲突了。

```bash
[老王]$ git push origin master 
To git@data-git.mana.com:yafo_bitch/git-tutorial.git
 ! [rejected]        master -> master (fetch first)
error: failed to push some refs to 'git@data-git.mana.com:yafo_bitch/git-tutorial.git'
hint: Updates were rejected because the remote contains work that you do
hint: not have locally. This is usually caused by another repository pushing
hint: to the same ref. You may want to first integrate the remote changes
hint: (e.g., 'git pull ...') before pushing again.
hint: See the 'Note about fast-forwards' in 'git push --help' for details.
```

提示信息显示远程库的改动和本地不一致，必须要先 pull 下来再 push。这个时候可以使用命令 pull（用 pull 命令一定要加 ```--rebase``` 参数） 先把远程的改动拉下来。拉下来后会出现两种情况。

一、你和别人代码没有冲突。pull 完了以后，这个时候你可以再一次 push 就好了。

二、你和别人的代码有冲突，这种情况一般是你和别人改了同一个位置的代码。就像这次王大锤和张大锤同时改了第二行的剧本一样。会产生如下的输出：

```bash
[老王]$ git pull --rebase origin master   #一定要加 --rebase 参数。
remote: Counting objects: 3, done.
remote: Compressing objects: 100% (2/2), done.
remote: Total 3 (delta 1), reused 0 (delta 0)
Unpacking objects: 100% (3/3), done.
From data-git.mana.com:yafo_bitch/git-tutorial
 * branch            master     -> FETCH_HEAD
   84179ac..060cd28  master     -> origin/master
First, rewinding head to replay your work on top of it...
Applying: 两情若是久长时，一枝红杏出墙来。
Using index info to reconstruct a base tree...
M	scenario
Falling back to patching base and 3-way merge...
Auto-merging scenario
CONFLICT (content): Merge conflict in scenario
Failed to merge in the changes.
Patch failed at 0001 两情若是久长时，一枝红杏出墙来。
The copy of the patch that failed is found in:
   /Users/老王/t/1/.git/rebase-apply/patch

When you have resolved this problem, run "git rebase --continue".
If you prefer to skip this patch, run "git rebase --skip" instead.
To check out the original branch and stop rebasing, run "git rebase --abort".
```

以上信息显示同时被修改的文件是 scenario（注意前面那个大写的 M，表示 modify 的意思），conflict 出现在了这个文件。我们先把注意力放到最下的三行。下面有请伟大的中文，英文，计算机文，三语使用者张全蛋先生为我们翻译。

```bash
When you have resolved this problem, run "git rebase --continue".
搞定所有的麻烦后请使用 "git rebase --continue" 命令.

If you prefer to skip this patch, run "git rebase --skip" instead.
如果你想挨揍请使用 "git rebase --skip" 命令。

To check out the original branch and stop rebasing, run "git rebase --abort".
如果你想被开除请使用 "git rebase --abort" 命令。
```

很明显，为了生命安全，生活质量我们要按照第一条规则来行事。为了更方便解决冲突我们可以使用 git status 来查看更具体的冲突信息。

```bash
[老王]$ git status
rebase in progress; onto 060cd28
You are currently rebasing branch 'master' on '060cd28'.
  (fix conflicts and then run "git rebase --continue")
  (use "git rebase --skip" to skip this patch)
  (use "git rebase --abort" to check out the original branch)

Unmerged paths:
  (use "git reset HEAD <file>..." to unstage)
  (use "git add <file>..." to mark resolution)

	both modified:   scenario

no changes added to commit (use "git add" and/or "git commit -a")
```

这样就可以更清楚的看到具体的 both modified 文件路径。现在我们打开 scenario 文件可以看到另一个作者张大锤先生写的剧本。

```bash
     1	从前有个人，名字叫赵铁柱。他人长得帅，又多金，还会写代码。
     2	<<<<<<< HEAD
     3	后来他死了。
     4	=======
     5	他最大的爱好是吃着火锅，听着 beatles 与高富帅们谈笑风声。
     6	>>>>>>> 两情若是久长时，一枝红杏出墙来。
```

现在 scenario 文件告诉我们具体冲突的地方（使用git diff scenario 可以快速查看冲突） 。

就是以 `<<<<<<< HEAD` 开始到 ` >>>>>>>` 结束的地方。中间的 `=======` 用来分割开两次的改动。

`<<<<<<< HEAD` 下面的是别人那边提交过来的代码

 `=======`  下面的是我们这次修改后的代码

决定好留低3行或第5行后，在文件里面其他的所有东西都要都要删除掉。

王大锤和张大锤后来打了很多场架，留过很多血后，终于决定了到底保留第3行还是第5行。

这时王大锤可以继续提交了。以下是修改好冲突后提交的流程：

```bash
[老王]$ git add scenario  #由于只是修改冲突，所以并不需要再使用 commit
[老王]$ git rebase --continue
```

这时可以使用 push 把本地的修改提交到远程去。

```bash
[老王]$ git push origin master
Counting objects: 3, done.
Delta compression using up to 8 threads.
Compressing objects: 100% (2/2), done.
Writing objects: 100% (3/3), 462 bytes | 0 bytes/s, done.
Total 3 (delta 0), reused 0 (delta 0)
To git@data-git.mana.com:yafo_bitch/git-tutorial.git
   060cd28..7f1794c  master -> master
```

Ok，Everything is all right. 幸福的生活可以继续了。



## Summary

为了在关键时刻不手忙脚乱的再从头开始读文章，我把所有的步骤全列在下面供你们快速查看。$ 后面表示命令行。千万别使用 ```Ctrl-C``` copy 任何命令，否则出现生命危险后果请自负！

```bash
#正常步骤
....  edit some files ....
$ git add edited_file
$ git commit -m "some changes message"
$ git push origin master

#冲突步骤
....  edit some files ....
$ git add edited_file
$ git commit -m "some changes message"
$ git push origin master
....  failed to push some refs to ....
$ git pull --rebase origin master

## 如果没有冲突？
$ git push origin master

## 如果有冲突？
... resolved all of conflicts ...
$ git add some_resolved_file
$ git rebase --contine
$ git push origin master
```

