package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/lucas-clemente/quic-go/h2quic"
	"github.com/lucas-clemente/quic-go/utils"
)

var indexQuic, indexH2, rtt, loss string

func testHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, `<body style="font-family: sans-serif; background-color: #5cb85c; color: white; text-align: center; padding-top: 30px;">Loaded</body>`)
}

func notQuicHandler(w http.ResponseWriter, req *http.Request) {
	_, port, _ := net.SplitHostPort(req.Host)
	setHeaders(port, w.Header())
	io.WriteString(w, "Not using QUIC, reload or see instructions.")
}

func tileHandler(w http.ResponseWriter, req *http.Request) {
	i := req.URL.Query().Get("i")

	b, err := Asset(fmt.Sprintf("html/tiles/tile_%s.png", i))
	if err != nil {
		println(err)
	}
	w.Write(b)
}

func notQuicTileHandler(w http.ResponseWriter, req *http.Request) {
	_, port, _ := net.SplitHostPort(req.Host)
	setHeaders(port, w.Header())
	w.WriteHeader(404)
}

func runH2Server(port string, handler http.Handler, closeAfterFirst bool) {
	h2server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}
	if closeAfterFirst {
		h2server.ConnState = func(c net.Conn, cs http.ConnState) {
			if cs == http.StateIdle {
				go func() {
					time.Sleep(100 * time.Millisecond)
					c.Close()
				}()
			}
		}
	}
	h2server.ListenAndServeTLS("/certs/cert.pem", "/certs/privkey.pem")
}

const nImages = 90
const nImagesPerLine = 10

func writeLoss(w io.Writer) {
	s := ""
	for i := 0; i < nImages; i++ {
		s += fmt.Sprintf(`<img src="https://quic.clemente.io:8010/tile?i=%d&cachebust=%d">`, i, time.Now().Nanosecond())
		if i%nImagesPerLine == nImagesPerLine-1 {
			s += "<br>"
		}
	}

	r := strings.Replace(loss, "TILES", s, 1)
	r = strings.Replace(r, "CURRENT", "TCP+TLS+HTTP/2", 1)

	io.WriteString(w, r)
}

func writeLossQuic(w io.Writer) {
	s := ""
	for i := 0; i < nImages; i++ {
		s += fmt.Sprintf(`<img src="https://quic.clemente.io:8015/tile-quic?i=%d&cachebust=%d">`, i, time.Now().Nanosecond())
		if i%nImagesPerLine == nImagesPerLine-1 {
			s += "<br>"
		}
	}

	r := strings.Replace(loss, "TILES", s, 1)
	r = strings.Replace(r, "CURRENT", "QUIC", 1)

	io.WriteString(w, r)
}

func runQuicServer(port string, handler http.Handler, closeAfterFirst bool) {
	server := &h2quic.Server{
		Server: &http.Server{
			Addr:      ":" + port,
			Handler:   handler,
			TLSConfig: GetTLSConfig(),
		},
		CloseAfterFirstRequest: closeAfterFirst,
	}
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func GetTLSConfig() *tls.Config {
	cert, err := tls.LoadX509KeyPair("/certs/fullchain.pem", "/certs/privkey.pem")
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
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
	data, _ = Asset("html/loss.html")
	loss = string(data)

	var err error

	quicRTTMux := http.NewServeMux()
	h2RTTMux := http.NewServeMux()

	h2RTTMux.HandleFunc("/test", testHandler)
	h2RTTMux.HandleFunc("/test-quic", notQuicHandler)

	quicRTTMux.HandleFunc("/test-quic", testHandler)

	go runH2Server("8000", h2RTTMux, true)
	go runH2Server("8001", h2RTTMux, true)
	go runH2Server("8002", h2RTTMux, true)
	go runH2Server("8003", h2RTTMux, true)

	go runH2Server("8005", h2RTTMux, true)
	go runH2Server("8006", h2RTTMux, true)
	go runH2Server("8007", h2RTTMux, true)
	go runH2Server("8008", h2RTTMux, true)

	go runQuicServer("8005", quicRTTMux, true)
	go runQuicServer("8006", quicRTTMux, true)
	go runQuicServer("8007", quicRTTMux, true)
	go runQuicServer("8008", quicRTTMux, true)

	quicLossMux := http.NewServeMux()
	h2LossMux := http.NewServeMux()

	h2LossMux.HandleFunc("/tile-quic", notQuicTileHandler)
	h2LossMux.HandleFunc("/tile", tileHandler)

	quicLossMux.HandleFunc("/tile-quic", tileHandler)

	go runH2Server("8010", h2LossMux, false)
	go runH2Server("8015", h2LossMux, false)

	go runQuicServer("8015", quicLossMux, false)

	quicMux := http.NewServeMux()
	h2Mux := http.NewServeMux()

	quicServer := &h2quic.Server{
		Server: &http.Server{
			Addr:      ":7000",
			Handler:   quicMux,
			TLSConfig: GetTLSConfig(),
		},
	}

	quicMux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) { io.WriteString(w, indexQuic) })
	quicMux.HandleFunc("/rtt", func(w http.ResponseWriter, req *http.Request) { io.WriteString(w, rtt) })
	quicMux.HandleFunc("/loss", func(w http.ResponseWriter, req *http.Request) { writeLoss(w) })
	quicMux.HandleFunc("/loss-quic", func(w http.ResponseWriter, req *http.Request) { writeLossQuic(w) })

	h2Mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		setHeaders("443", w.Header())
		io.WriteString(w, indexH2)
	})
	h2Mux.HandleFunc("/rtt", func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, rtt)
	})
	h2Mux.HandleFunc("/loss", func(w http.ResponseWriter, req *http.Request) {
		writeLoss(w)
	})

	go func() {
		err := quicServer.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()

	err = http.ListenAndServeTLS(":7000", "/certs/fullchain.pem", "/certs/privkey.pem", h2Mux)
	if err != nil {
		panic(err)
	}
}
