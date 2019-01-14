var MJPEGStream = function () {
    
    /**********************************************************/
    /*                                                        */
    /*                       事件处理                          */
    /*                                                        */
    /**********************************************************/
    function EventEmitter() {
        this.events = {};
    }
    //绑定事件函数
    EventEmitter.prototype.on = function (eventName, callback) {
        this.events[eventName] = this.events[eventName] || [];
        this.events[eventName].push(callback);
    };
    //触发事件函数
    EventEmitter.prototype.emit = function (eventName, _) {
        var events = this.events[eventName],
            args = Array.prototype.slice.call(arguments, 1),
            i, m;

        if (!events) {
            return;
        }
        for (i = 0, m = events.length; i < m; i++) {
            events[i].apply(null, args);
        }
    };

    /**********************************************************/
    /*                                                        */
    /*                   建立WebSocket连接                     */
    /*                                                        */
    /**********************************************************/


    /*******************基础部分*********************/
    function mjpegstream() {
        //所在房间
        this.room = "";
        //所在窗口
        this.windowId = "";
        //本地WebSocket连接
        this.socket = null;
        //本地流
        this.localMediaStream = null;
    }
    //继承自事件处理器，提供绑定事件和触发事件的功能
    mjpegstream.prototype = new EventEmitter();

    /*************************服务器连接部分***************************/
    mjpegstream.prototype.connect = function (server, room, windowId) {
        this.room = room || "";
        this.windowId = windowId || "";
        var socket,
            that = this;
        socket = this.socket = new WebSocket(server);
        socket.onopen = function (evt) {
            console.log("Connection open ...");
            socket.send(
                JSON.stringify({
                    room: that.room,
                    window: that.windowId
                })
            );
            that.emit("socket_opened", socket);
        };

        socket.onmessage = function (evt) {
            console.log("Received Message: " + evt.data);
            var userInfo = JSON.parse(evt.data);
            var streamUrl = "http://" + userInfo.ip + ":8000/camera/mjpeg"
            if (userInfo.joined) {
                that.emit("add_stream", userInfo.window, streamUrl);
            } else {
                that.emit("remove_stream", userInfo.window);
            }
        };

        socket.onclose = function (evt) {
            console.log("Connection closed.");
            that.emit('socket_closed', socket);
        };

        window.onbeforeunload = function () {
            socket.onclose = function () { }; // disable onclose handler first
            socket.close();
        };

    }

    /*************************流处理部分*******************************/

    //创建本地流
    mjpegstream.prototype.createStream = function (options) {
        var that = this;

        options.video = !!options.video;
        options.audio = !!options.audio;

        if (navigator.mediaDevices.getUserMedia) {
            navigator.mediaDevices.getUserMedia(options)
                .then(function (stream) {
                    that.localMediaStream = stream;
                    that.emit("stream_created", stream);
                })
                .catch(function (err) {
                    /* 处理error */
                    that.emit("stream_create_error", err);
                });
        } else {
            that.emit("stream_create_error", new Error('WebRTC is not yet supported in this browser.'));
        }
    }

    //关闭本地流
    mjpegstream.prototype.stopStream = function () {
        if (this.localMediaStream) {
            this.localMediaStream.getTracks().forEach(track => track.stop());
            this.localMediaStream = null;
        }
    }

    return new mjpegstream();
}