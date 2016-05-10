# QUIC Demo (Work in progress)

Important commands:

```sh
go build -o main && sudo docker build -t lclemente/quic-demo . && sudo docker run --rm -it --d=NET_ADMIN -p 8000-8009:8000-8009 -p 8000-8009:8000-8009/udp -v /etc/letsencrypt/live/quic.clemente.io:/certs lclemente/quic-demo

/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome  --user-data-dir=/tmp/chrome --origin-to-force-quic-on=quic.clemente.io:8005,quic.clemente.io:8006,quic.clemente.io:8007,quic.clemente.io:8008,quic.clemente.io:8009 --enable-quic https://quic.clemente.io:8000
```
