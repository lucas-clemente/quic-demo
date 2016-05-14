package main

import (
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/lucas-clemente/quic-go/h2quic"
	"github.com/lucas-clemente/quic-go/testdata"
	"github.com/lucas-clemente/quic-go/utils"
)

var quicServer *h2quic.Server

var indexQuic, indexH2, rtt string

func testHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, `<body style="font-family: sans-serif; background-color: #5cb85c; color: white; text-align: center; padding-top: 30px;">Loaded</body>`)
}

func notQuicHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Not using QUIC, see instructions.")
}

func runH2Server(port string, handler http.Handler) {
	h2server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}
	h2server.ConnState = func(c net.Conn, cs http.ConnState) {
		if cs == http.StateIdle {
			go func() {
				time.Sleep(100 * time.Millisecond)
				c.Close()
			}()
		}
	}
	h2server.ListenAndServeTLS("/certs/cert.pem", "/certs/privkey.pem")
}

func runQuicServer(port string, handler http.Handler) {
	err := quicServer.ListenAndServe(":"+port, handler)
	if err != nil {
		panic(err)
	}
}

func main() {
	utils.SetLogLevel(utils.LogLevelInfo)

	data, _ := Asset("html/rtt.html")
	rtt = string(data)
	data, _ = Asset("html/index.html")
	indexQuic = strings.Replace(string(data), "QUIC_STATUS", "yes", -1)
	indexH2 = strings.Replace(string(data), "QUIC_STATUS", "no", -1)

	tlsConfig := testdata.GetTLSConfig()
	var err error
	quicServer, err = h2quic.NewServer(tlsConfig)
	if err != nil {
		panic(err)
	}
	quicServer.CloseAfterFirstRequest = true

	quicMux := http.NewServeMux()
	h2Mux := http.NewServeMux()

	h2Mux.HandleFunc("/test", testHandler)
	h2Mux.HandleFunc("/test-quic", notQuicHandler)
	h2Mux.HandleFunc("/rtt", func(w http.ResponseWriter, req *http.Request) { io.WriteString(w, rtt) })
	h2Mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) { io.WriteString(w, indexH2) })

	quicMux.HandleFunc("/test-quic", testHandler)
	quicMux.HandleFunc("/rtt", func(w http.ResponseWriter, req *http.Request) { io.WriteString(w, rtt) })
	quicMux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) { io.WriteString(w, indexQuic) })

	go runH2Server("8000", h2Mux)
	go runH2Server("8001", h2Mux)
	go runH2Server("8003", h2Mux)
	go runH2Server("8004", h2Mux)

	go runH2Server("8005", h2Mux)
	go runH2Server("8006", h2Mux)
	go runH2Server("8008", h2Mux)
	go runH2Server("8009", h2Mux)

	go runQuicServer("8005", quicMux)
	go runQuicServer("8006", quicMux)
	go runQuicServer("8008", quicMux)
	go runQuicServer("8009", quicMux)

	mainQuicServer, err := h2quic.NewServer(tlsConfig)
	if err != nil {
		panic(err)
	}
	go mainQuicServer.ListenAndServe(":7000", quicMux)

	err = http.ListenAndServeTLS(":7000", "/certs/cert.pem", "/certs/privkey.pem", h2Mux)
	if err != nil {
		panic(err)
	}
}
