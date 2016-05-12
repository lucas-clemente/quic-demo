package main

import (
	"io"
	"net/http"
	"time"

	"github.com/lucas-clemente/quic-go/h2quic"
	"github.com/lucas-clemente/quic-go/testdata"
	"github.com/lucas-clemente/quic-go/utils"
)

func testHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, `<body style="font-family: sans-serif; background-color: #5cb85c; color: white; text-align: center; padding-top: 30px;">Loaded</body>`)
}

func notQuicHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Not using QUIC, see instructions.")
}

const timeout = 4 * time.Second

func runH2Server(port string, handler http.Handler) {
	h2server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}
	h2server.ReadTimeout = timeout
	h2server.ListenAndServeTLS("/certs/cert.pem", "/certs/privkey.pem")
}

func startServer(port string) {
	server := &http.Server{
		Addr:    ":" + port,
		Handler: nil,
	}
	server.ReadTimeout = timeout
	server.ListenAndServeTLS("/certs/cert.pem", "/certs/privkey.pem")
}

var quicServer *h2quic.Server

func runQuicServer(port string, handler http.Handler) {
	err := quicServer.ListenAndServe(":"+port, handler)
	if err != nil {
		panic(err)
	}
}

func main() {
	utils.SetLogLevel(utils.LogLevelInfo)

	tlsConfig := testdata.GetTLSConfig()
	var err error
	quicServer, err = h2quic.NewServer(tlsConfig)
	if err != nil {
		panic(err)
	}

	quicMux := http.NewServeMux()
	h2Mux := http.NewServeMux()

	h2Mux.HandleFunc("/using", func(w http.ResponseWriter, req *http.Request) { io.WriteString(w, "not using QUIC :(") })
	h2Mux.HandleFunc("/test", testHandler)
	h2Mux.HandleFunc("/test-quic", notQuicHandler)
	h2Mux.Handle("/", http.FileServer(assetFS()))

	quicMux.HandleFunc("/using", func(w http.ResponseWriter, req *http.Request) { io.WriteString(w, "using QUIC :)") })
	quicMux.HandleFunc("/test-quic", testHandler)
	quicMux.Handle("/", http.FileServer(assetFS()))

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
	go runQuicServer("8000", quicMux)

	go runQuicServer("7000", quicMux)
	http.ListenAndServeTLS(":7000", "/certs/cert.pem", "/certs/privkey.pem", h2Mux)
}
