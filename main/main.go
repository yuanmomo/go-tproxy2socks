package main

import (
	"flag"
	"golang.org/x/net/proxy"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	tproxy "go-tproxy2socks/transocks"
	"time"
)

var (
	server = flag.String("server", "0.0.0.0:1080", "local server.")
	socks  = flag.String("socks", "socks5://10.10.1.2:1086", "socks5 proxy address, like: socks5://10.20.30.40:1086.")
	debug  = flag.Bool("debug", false, "debug mode.")
	test  = flag.Bool("test", false, "debug mode.")
)

func main() {
	// parse arguments
	flag.Parse()
	log.Printf("Start go-tproxy2socks with params: server=[%s], socks=[%s], debug:[%v]\n", *server, *socks, *debug)

	if *socks == "" {
		log.Fatalf("ERROR: socks url is nil\n")
		return
	}

	// listen the servers
	listen, err := net.Listen("tcp", *server)
	if err != nil {
		log.Fatalf("ERROR: listen on local server error, %s\n", err.Error())
		return
	}
	log.Printf("local server started:[%s]", *server)
	defer listen.Close()

	// check socks
	ProxyURL, err := url.Parse(*socks)
	if err != nil {
		log.Fatalf("ERROR: invalid socks address, %s\n", err.Error())
		return
	}
	proxyDialer, err := proxy.FromURL(ProxyURL,&net.Dialer{ } )
	if err != nil {
		log.Fatalf("ERROR: connect socks proxy server error , %s\n", err.Error())
		return
	}
	log.Printf("Dial to proxy server success:[%s]", ProxyURL.String())

	running := true
	go func(){
		for running {
			conn, err := listen.Accept()
			if err != nil {
				if running {
					log.Printf("ERROR: Listener.Accept error: %s\n", err.Error())
				}
				continue
			}
			go handleConnection(*test,conn,&proxyDialer)
		}
	}()
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-c
	log.Println("[INFO] Exiting go-tproxy2socks...")
	running = false
	listen.Close()
}

// copyBuffer returns any write errors or non-EOF read errors, and the amount
// of bytes written.
func copyBuffer(dst io.Writer, src io.Reader, buf []byte)  error {
	for {
		nr, rerr := src.Read(buf)
		if nr > 0 {
			nw, werr := dst.Write(buf[:nr])
			if werr != nil {
				return werr
			}
			if nr != nw {
				return io.ErrShortWrite
			}
		}
		if rerr != nil {
			if rerr == io.EOF {
				rerr = nil
			}
			return rerr
		}
	}
}

func handleConnection(test bool, conn net.Conn,socksDialer *proxy.Dialer) {
	if test {
		log.Printf("handleConnection ..... ")
	}
	// 设置超时？
	conn.(*net.TCPConn).SetKeepAlive(true)
	conn.(*net.TCPConn).SetKeepAlivePeriod(5 * time.Second)
	conn.(*net.TCPConn).SetNoDelay(false)

	tc, ok := conn.(*net.TCPConn)
	if !ok {
		log.Printf("non-TCP connection \n")
		return
	}

	origAddr, err := tproxy.GetOriginalDST(tc)
	destConn, err := (*socksDialer).Dial("tcp", origAddr.String())
	if err != nil {
		log.Printf("ERROR: failed to connect to dst:[%s] through proxy server, %s\n", origAddr.String(), err.Error())
		return
	}
	destConn.(*net.TCPConn).SetKeepAlive(true)
	destConn.(*net.TCPConn).SetKeepAlivePeriod(20 * time.Second)
	destConn.(*net.TCPConn).SetNoDelay(false)

	if test {
		log.Printf("[INFO] Connect to %s.\n", origAddr.String())
	}
	defer destConn.Close()

	// do proxy
	copyEnd := false;
	go func() {
		buf := leakyBuf.Get()
		err := copyBuffer(destConn, tc, buf)
		if err != nil && !copyEnd {
			log.Printf("ERROR: Copy client to proxy error: %s.\n", err.Error())
		}
		leakyBuf.Put(buf)
		copyEnd = true
		conn.Close()
		destConn.Close()
	}()

	buf := leakyBuf.Get()
	err = copyBuffer(conn, destConn, buf)
	if err != nil && !copyEnd {
		log.Printf("ERROR: Copy proxy to client error: %s.\n", err.Error())
	}
	leakyBuf.Put(buf)
	copyEnd = true

	destConn.Close()
	conn.Close()
}
