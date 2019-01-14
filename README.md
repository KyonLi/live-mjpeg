# live-mjpeg

MJPEG连接摄像头方案

## 简介

这是一个基于MJPEG-Stream实现主播、刮票口远程连接摄像头的Demo

## 安装和使用

1. 安装go及配置环境
2. 下载源码到本地，根据实际情况修改`html/caster.html`和`html/operator.html`中websocket地址
3. 使用命令`go get github.com/gorilla/websocket`安装websocket库
4. 运行命令`go run server.go`
5. 刮票口安装并运行[cam2web](https://github.com/cvsandbox/cam2web)，使用Caddy反代到8080端口
6. 将`html`放至Web目录，使用浏览器打开即可