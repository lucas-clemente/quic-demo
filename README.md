# QUIC Demo (Work in progress)

Ports:

- 7000 tcp/udp: Main page of the demo.
- 8000 tcp: RTT demo, 0ms
- 8001 tcp: RTT demo, 100ms
- 8002 tcp: RTT demo, 500ms
- 8003 tcp: RTT demo, 1s
- 8005 udp: RTT demo, 0ms
- 8006 udp: RTT demo, 100ms
- 8007 udp: RTT demo, 500ms
- 8008 udp: RTT demo, 1s

Important commands:

```sh
go-bindata html/... && \
go build -o main && \
sudo docker build -t lclemente/quic-demo . && \
sudo docker kill quic-demo; \
 sudo docker rm quic-demo; \
sudo docker run --name quic-demo -d --cap-add=NET_ADMIN \
-p 8000-8009:8000-8009 -p 8000-8009:8000-8009/udp \
-p 7000:7000 -p 443:7000/udp \
-v /etc/letsencrypt/live/quic.clemente.io:/certs \
lclemente/quic-demo

/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --user-data-dir=/tmp/chrome --origin-to-force-quic-on=quic.clemente.io:443,quic.clemente.io:8005,quic.clemente.io:8006,quic.clemente.io:8007,quic.clemente.io:8008 --enable-quic https://quic.clemente.io
```
