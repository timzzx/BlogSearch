# BlogSearch
用golang walk写的一款博客搜索查看的windows下的GUI软件。


功能比较简单。

1.有五个博客网站可以查询
2.翻页功能
3.有收藏功能
 

-**界面** 

![输入图片说明](https://git.oschina.net/uploads/images/2017/0825/114137_8a935026_462123.png "blogsearch.png")

![输入图片说明](https://git.oschina.net/uploads/images/2017/0825/114331_16277e83_462123.png "blogseach2.png")


上面是编译好的软件界面。

之前写过一个简书查询的软件过于简陋，有一些bug，使用起来不是很方便。
我就想到写个集合几个blog网站带收藏功能的软件，这就有了博客查询这个小软件。

题外话：写这个软件就是为了更好的学习golang，用之编写第一个GUI的应用，不做商业用途，只为学习golang，并且是写一个有 **一点点** 价值的软件。（以前一直写web /(ㄒoㄒ)/~~）

 **运行需要** 

golang版本需要1.8以上
go get github.com/lxn/walk
go get github.com/akavel/rsrc

 **执行** 

go build -ldflags="-H windowsgui"

 **或者** 

go build


