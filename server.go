// Copyright 2018 Kyon Li. All rights reserved.
// Use of this source code is governed by a BSD-style

package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gorilla/websocket"
)

// 客户端信息结构体
type UserInfo struct {
	Room    string `json:"room"`
	Window  string `json:"window"`
	IP      string `json:"ip"`
	Joined  bool   `json:"joined"`
}

var clients = make(map[*websocket.Conn]UserInfo)  // 已连接客户端map
var broadcast = make(chan UserInfo)               // 消息队列

// 配置upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// GET请求升级到websocket
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		var info UserInfo
		// 读取JSON格式消息并转换为UserInfo
		err := c.ReadJSON(&info)
		if err != nil {
			log.Println(err)
			// 生成离开房间通知
			var info = clients[c]
			info.Joined = false
			log.Println(info)
			broadcast <- info

			delete(clients, c)
			break
		}

		// 断开非法连接
		if info.Room == "" || info.Window == "" {
			break
		}

		// 如果是主播端则发送给他所有客户端信息
		if info.Window == "0" {
			sendAllClientsInfo(c, info.Room)
		}

		// 获取客户端ip
		ips := r.Header.Get("X-Forwarded-For")
		if ips != "" {
			info.IP = strings.Split(ips, ", ")[0]
		} else {
			addrStr := c.RemoteAddr().String()
			host, _, err := net.SplitHostPort(addrStr)
			if err != nil {
				info.IP = addrStr
			} else {
				info.IP = host
			}
		}
		info.Joined = true
		log.Println(info)

		// 客户端连接保存到clients
		clients[c] = info
		// 将新消息加入消息队列
		broadcast <- info
	}
}

func handleMessages() {
	// 从消息队列取出UserInfo
	for info := range broadcast {
		// 向房间号相同的主播发消息
		for client := range clients {
			if info.Window != "0" && clients[client].Window == "0" && clients[client].Room == info.Room {
				err := client.WriteJSON(info)
				if err != nil {
					log.Printf("error: %v", err)
					client.Close()
					delete(clients, client)
				}
			}
		}
	}
}

func sendAllClientsInfo(c *websocket.Conn, room string) {
	// 房间号相同并且不是主播
	for client := range clients {
		if clients[client].Room == room && clients[client].Window != "0" {
			c.WriteJSON(clients[client])
		}
	}
}

func closeAll()  {
	for c := range clients {
		c.Close()
	}
	close(broadcast)
}

func startHttpServer() *http.Server {
	srv := &http.Server{Addr: ":7080"}

	http.HandleFunc("/", handleConnections)

	go func() {
		// returns ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// NOTE: there is a chance that next line won't have time to run,
			// as main() doesn't wait for this goroutine to stop. don't use
			// code with race conditions like these for production. see post
			// comments below on more discussion on how to handle this.
			log.Fatalf("ListenAndServe(): %s", err)
		}
	}()

	// returning reference so caller can call Shutdown()
	return srv
}

func main() {
	// 启动websocket服务器
	log.SetFlags(0)
	go handleMessages()
	srv := startHttpServer()

	{
		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM)
		<-osSignals

		closeAll()
		if err := srv.Shutdown(nil); err != nil {
			panic(err) // failure/timeout shutting down the server gracefully
		}
		os.Exit(0)
	}
}
