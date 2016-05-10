package main

import (
	"io"
	"net/http"
	"time"

	"github.com/lucas-clemente/quic-go/h2quic"
	"github.com/lucas-clemente/quic-go/testdata"
	"github.com/lucas-clemente/quic-go/utils"
)

const index = `
<html>
<head>
  <title>QUIC Demo</title>
  <style>
    iframe {
      width: 100px;
      height: 100px;
    }
  </style>
</head>
<body>
  <table>
    <tr>
      <th></th>
      <th>0 ms</th>
      <th>100 ms</th>
      <th>500 ms</th>
      <th>1 s</th>
    </tr>
    <tr>
      <td>TCP+TLS</td>
      <td><iframe src="https://quic.clemente.io:8000/test"></iframe></td>
      <td><iframe src="https://quic.clemente.io:8001/test"></iframe></td>
      <td><iframe src="https://quic.clemente.io:8003/test"></iframe></td>
      <td><iframe src="https://quic.clemente.io:8004/test"></iframe></td>
    </tr>
    <tr>
      <td>QUIC</td>
      <td><iframe src="https://quic.clemente.io:8005/test"></iframe></td>
      <td><iframe src="https://quic.clemente.io:8006/test"></iframe></td>
      <td><iframe src="https://quic.clemente.io:8008/test"></iframe></td>
      <td><iframe src="https://quic.clemente.io:8009/test"></iframe></td>
    </tr>
    <tr>
    </tr>
  </table>
</body>
</html>
`

func indexHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, index)
}

func testHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "<body bgcolor=green>Loaded</body>")
}

func startServer(port string) {
	server := &http.Server{
		Addr:    ":" + port,
		Handler: nil,
	}
	server.ReadTimeout = 3 * time.Second
	server.ListenAndServeTLS("/certs/cert.pem", "/certs/privkey.pem")
}

var quicServer *h2quic.Server

func startQuicServer(port string) {
	go func() {
		err := quicServer.ListenAndServe(":"+port, nil)
		if err != nil {
			panic(err)
		}
	}()

	h2server := &http.Server{
		Addr: ":" + port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Alternate-Protocol", port+":quic")
			io.WriteString(w, "Not QUIC, refresh")
		}),
	}
	h2server.ReadTimeout = 3 * time.Second
	h2server.ListenAndServeTLS("/certs/cert.pem", "/certs/privkey.pem")
}

func main() {
	tlsConfig := testdata.GetTLSConfig()
	var err error
	quicServer, err = h2quic.NewServer(tlsConfig)
	if err != nil {
		panic(err)
	}

	utils.SetLogLevel(utils.LogLevelInfo)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/test", testHandler)
	go startServer("8001")
	go startServer("8002")
	go startServer("8003")
	go startServer("8004")
	go startQuicServer("8005")
	go startQuicServer("8006")
	go startQuicServer("8007")
	go startQuicServer("8008")
	go startQuicServer("8009")
	startServer("8000")
}
