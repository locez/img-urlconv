img-urlconv
---
img-urlconv 是一个用于替换 [Linux 中国](https://linux.cn/) 与 [LCTT/TranslateProject](https://github.com/LCTT/TranslateProject) 文章图片 url 的工具

背景
---
由于历史原因，`LCTT/TranslateProject` 中一些图片 url 已经失效，当读者从仓库直接阅读文章时体验会不好，因此该项目主要是为了解决此问题而诞生的

使用
---

``` bash
Usage of img-urlconv:
  -b int
    	the begbin ID of article (default -1)
  -e int
    	the end ID of article (default -1)
  -f string
    	the published dir
```

`-b` `-e` 指定从 [Linux 中国](https://linux.cn/) 遍历的范围，用 `-f` 参数指定 `published dir`

example:

```
./img-urlconv -b 10000 -e 10100 -f ~/LCTT/TranslateProject/published/
```

说明
---

本工具从 [Linux 中国](https://linux.cn/) 中提取 via 再结合 ack 对仓库进行搜索文件，从而确定要对哪个文件进行操作，由于早期文章格式较不规则，特例比较多，因此本工具输出比较多，需要人为干预的情况也比较多，以下是对场景的说明：

### 场景一

文章中未含有 via，直接控制台打印：

```
2018/10/27 15:29:02 No via found : https://linux.cn/article-10010-1.html?pr
```
此种情况如有误报，则是不规则的 via，目前工具已经支持 `原文` `英文原文` `via` 三个关键词

### 场景二

能找到 via，但是无法找到文章，直接控制台输出：

```
2018/10/27 15:25:48 No file found
    url:https://linux.cn/article-10068-1.html?pr
    via:https://www.eyrie.org/%7Eeagle/reviews/books/1-62779-037-3.html
```
此种情况可能确实是找不到文章，也就是文件不存在，其次就是文件存在，但是由于 url 被转码导致搜索不出来，该例输出便是这种情况

### 场景三

对于图片数量能一一对应的，则直接进行一一替换（包含仓库 md 文件不含有图片，但是官网含有题图的场景），此场景直接对源文件进行操作

### 场景四

图片数量不一致，但是官网图片数量较多，此时假定多出是题图，尝试进行替换操作，但会同时在相同目录下生成一个 `.backup.md` 备份文件，可人工确定无误后删除备份文件或者修改版本

### 场景五

该场景是能找到文件，但是图片数量不一致或者其它情况，在判定之外的，将会在你执行 `img-urlconv` 的当前目录生成一个 `failurls.log` 文件，该文件包含官网文章 url 以及对应的 md 文件，方便人工干预修复

```
2018/10/27 15:29:08

    url: https://linux.cn/article-10015-1.html?pr
    md: /home/locez/LCTT/TranslateProject/published/201809/20180802 Distrochooser Helps Linux Beginners To Choose A Suitable Linux Distribution.md
2018/10/27 15:29:15

    url: https://linux.cn/article-10045-1.html?pr
    md: /home/locez/LCTT/TranslateProject/published/201809/20180308 What is open source programming.md
```

