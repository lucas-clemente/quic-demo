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
		body {
			font-family: sans-serif;
			margin: 1em;
		}

		code {
			display: block;
			margin: 2em;
		}

    .demo {
      width: 100px;
      height: 100px;
			text-align: center;
    }

		footer {
			margin-top: 2em;
		}
  </style>
</head>
<body>
	<h1>QUIC Test Page</h1>

	<p>
		This webpage uses an experimental <a href="https://github.com/lucas-clemente/quic-go">Go implemention</a> of the novel <a href="https://en.wikipedia.org/wiki/QUIC">QUIC protocol</a> proposed by Google.
	</p>
	<p>
		To use it, you need to use a current Chrome version, and launch it using a special command from the console. For OS X, try:
		<code>
			/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --user-data-dir=/tmp/chrome --origin-to-force-quic-on=quic.clemente.io:8000,quic.clemente.io:8005,quic.clemente.io:8006,quic.clemente.io:8007,quic.clemente.io:8008,quic.clemente.io:8009 --enable-quic https://quic.clemente.io:8000
		</code>

		For linux, try:
		<code>
			chrome --user-data-dir=/tmp/chrome --origin-to-force-quic-on=quic.clemente.io:8000,quic.clemente.io:8005,quic.clemente.io:8006,quic.clemente.io:8007,quic.clemente.io:8008,quic.clemente.io:8009 --enable-quic https://quic.clemente.io:8000
		</code>
	</p>

	<h2>Simulated Round-Trip-Time (RTT)</h2>

	<p>
		QUIC features a custom crypto stack that is able to achieve connection in 0-1 RTTs. TCP+TLS usually needs 3-4 RTTs. If something doesn't work, just try reloading :)
	</p>

  <table>
    <tr>
      <th>Simulated RTT</th>
      <th>0 ms</th>
      <th>100 ms</th>
      <th>500 ms</th>
      <th>1 s</th>
    </tr>
    <tr>
      <td>TCP+TLS+HTTP2</td>
      <td><iframe class="demo" src="https://quic.clemente.io:8000/test"></iframe></td>
      <td><iframe class="demo" src="https://quic.clemente.io:8001/test"></iframe></td>
      <td><iframe class="demo" src="https://quic.clemente.io:8003/test"></iframe></td>
      <td><iframe class="demo" src="https://quic.clemente.io:8004/test"></iframe></td>
    </tr>
    <tr>
      <td>QUIC</td>
      <td><iframe class="demo" src="https://quic.clemente.io:8005/test"></iframe></td>
      <td><iframe class="demo" src="https://quic.clemente.io:8006/test"></iframe></td>
      <td><iframe class="demo" src="https://quic.clemente.io:8008/test"></iframe></td>
      <td><iframe class="demo" src="https://quic.clemente.io:8009/test"></iframe></td>
    </tr>
    <tr>
    </tr>
  </table>


	<footer>
		<hr/>
		Built by Lucas Clemente and Marten Seemann, see <a href="https://github.com/lucas-clemente/quic-go">GitHub</a>.
	</footer>
</body>
</html>
`

func indexHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, index)
}

func testHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, `<body style="font-family: sans-serif; background-color: #5cb85c; color: white; text-align: center; padding-top: 30px;">Loaded</body>`)
}

const timeout = 4 * time.Second

func runH2serverWithTimeout(port string, handler http.Handler) {
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
var quicServeMux *http.ServeMux

func startQuicServer(port string) {
	go runH2serverWithTimeout(
		port,
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Alternate-Protocol", port+":quic")
			io.WriteString(w, "Not using QUIC, see instructions above.")
		}),
	)
	err := quicServer.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}

func main() {
	var err error
	tlsConfig := testdata.GetTLSConfig()

	quicServer, err = h2quic.NewServer(tlsConfig)
	if err != nil {
		panic(err)
	}

	quicServeMux = http.NewServeMux()
	quicServeMux.HandleFunc("/test", testHandler)
	quicServeMux.HandleFunc("/", indexHandler)

	utils.SetLogLevel(utils.LogLevelInfo)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/test", testHandler)
	go runH2serverWithTimeout("8001", nil)
	go runH2serverWithTimeout("8002", nil)
	go runH2serverWithTimeout("8003", nil)
	go runH2serverWithTimeout("8004", nil)
	go startQuicServer("8005")
	go startQuicServer("8006")
	go startQuicServer("8007")
	go startQuicServer("8008")
	go startQuicServer("8009")
	go startQuicServer("8000")
	runH2serverWithTimeout("8000", nil)
}
