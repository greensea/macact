# macact
当局域网上主机的 IP 地址改变的时候，执行特定的命令

## 用法
```
macact <MAC 地址> <命令>
```
<命令>中的 “%h” 会被替换成 MAC 对应的 IP 地址


## 使用场景
正常情况下，局域网上的主机的 IP 地址基本上是不会变化的，然而在特定情况下，一个主机的 IP 地址仍有可能改变。但无论 IP 地址如何改变，主机的 MAC 地址是不会变的。这时候这个工具就派上用场了。

### 场景 1
局域网中有 RTSP 摄像头，其 MAC 地址是 00:11:22:33:44:55. 我们在服务器上用 ffmpeg 将 RTSP 流推到其他远程主机上。

```
macact 00:11:22:33:44:55 ffmpeg -rtsp_transport tcp -i rtsp://%h/stream.sdp -c:a copy -c:v copy -f flv rtmp://remote-host.com/feed
```

当摄像头的 IP 地址改变的时候（比如改变为 192.168.1.3），macact 就会执行这条命令

```
ffmpeg -rtsp_transport tcp -i rtsp://192.168.1.3/stream.sdp -c:a copy -c:v copy -f flv rtmp://remote-host.com/feed
```
如果摄像头的 IP 地址再次改变，macact 就会结束之前的 ffmpeg 进程，重新启动一个新的 ffmpeg 进程。
