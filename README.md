# web网站clone

递归实现自定义的多级网址clone


# Window下使用

```
go build

webclone.exe http://baidu.com

```

# 编译Linux可执行文件

``` shell
set GOOS = linux
go build
./webclone http://baidu.com

```

# 示例

```shell

-r #递归次数
-p #代理IP
-u #useragent请求头
-c #cookie


```
