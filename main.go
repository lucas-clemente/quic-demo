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

var indexQuic, indexH2, rtt string

func testHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, `<body style="font-family: sans-serif; background-color: #5cb85c; color: white; text-align: center; padding-top: 30px;">Loaded</body>`)
}

func notQuicHandler(w http.ResponseWriter, req *http.Request) {
	_, port, _ := net.SplitHostPort(req.Host)
	setHeaders(port, w.Header())
	io.WriteString(w, "Not using QUIC, reload or see instructions.")
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
	server := &h2quic.Server{
		Server: &http.Server{
			Addr:      ":" + port,
			Handler:   handler,
			TLSConfig: testdata.GetTLSConfig(),
		},
		CloseAfterFirstRequest: true,
	}
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func setHeaders(port string, header http.Header) {
	header.Add("Alternate-Protocol", port+":quic")
	header.Add("Alt-Svc", `quic=":`+port+`"; ma=2592000; v="33,32,31,30"`)
}

func main() {
	utils.SetLogLevel(utils.LogLevelInfo)

	data, _ := Asset("html/rtt.html")
	rtt = string(data)
	data, _ = Asset("html/index.html")
	indexQuic = strings.Replace(string(data), "QUIC_STATUS", "yes", -1)
	indexH2 = strings.Replace(string(data), "QUIC_STATUS", "no", -1)

	var err error

	quicRTTMux := http.NewServeMux()
	h2RTTMux := http.NewServeMux()

	h2RTTMux.HandleFunc("/test", testHandler)
	h2RTTMux.HandleFunc("/test-quic", notQuicHandler)

	quicRTTMux.HandleFunc("/test-quic", testHandler)

	go runH2Server("8000", h2RTTMux)
	go runH2Server("8001", h2RTTMux)
	go runH2Server("8003", h2RTTMux)
	go runH2Server("8004", h2RTTMux)

	go runH2Server("8005", h2RTTMux)
	go runH2Server("8006", h2RTTMux)
	go runH2Server("8008", h2RTTMux)
	go runH2Server("8009", h2RTTMux)

	go runQuicServer("8005", quicRTTMux)
	go runQuicServer("8006", quicRTTMux)
	go runQuicServer("8008", quicRTTMux)
	go runQuicServer("8009", quicRTTMux)

	quicMux := http.NewServeMux()
	h2Mux := http.NewServeMux()

	quicServer := &h2quic.Server{
		Server: &http.Server{
			Addr:      ":7000",
			Handler:   quicMux,
			TLSConfig: testdata.GetTLSConfig(),
		},
	}

	quicMux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) { io.WriteString(w, indexQuic) })
	quicMux.HandleFunc("/rtt", func(w http.ResponseWriter, req *http.Request) { io.WriteString(w, rtt) })

	h2Mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		setHeaders("443", w.Header())
		io.WriteString(w, indexH2)
	})
	h2Mux.HandleFunc("/rtt", func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, rtt)
	})

	go func() {
		err := quicServer.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()

	err = http.ListenAndServeTLS(":7000", "/certs/cert.pem", "/certs/privkey.pem", h2Mux)
	if err != nil {
		panic(err)
	}
}
