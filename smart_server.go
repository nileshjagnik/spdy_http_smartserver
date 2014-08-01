// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"crypto/tls"
	"github.com/amahi/spdy"
	"log"
	"net"
	"net/http"
	"runtime"
	"time"
)

const start_bytes = 4

func ListenAndServeTLSSmart(addr string, certFile string, keyFile string, handler http.Handler, smart_handler Handler) error {
	server := Server{
		Addr:    addr,
		Handler: smart_handler,
		TLSConfig: &tls.Config{
			NextProtos: []string{"spdy/3.1", "spdy/3", "http/1.1"},
		},
		TLSNextProto: map[string]func(*Server, *tls.Conn, Handler){
			"spdy/3.1": nextproto3,
			"spdy/3":   nextproto3,
		},
	}
	if server.Handler == nil {
		server.Handler = DefaultServeMux
	}

	server.inner_h = handler
	return server.ServerListenAndServeTLSSmart(certFile, keyFile)
}

func (srv *Server) ServerListenAndServeTLSSmart(certFile, keyFile string) error {
	addr := srv.Addr
	if addr == "" {
		addr = ":https"
	}
	config := &tls.Config{}
	if srv.TLSConfig != nil {
		*config = *srv.TLSConfig
	}
	if config.NextProtos == nil {
		config.NextProtos = []string{"spdy/3.1", "spdy/3", "http/1.1"}
	}
	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	tlsListener := tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener)}, config)
	srv.TLSConfig = config
	return srv.serveSmart(tlsListener)
}

func (srv *Server) serveSmart(l net.Listener) error {
	defer l.Close()
	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		rw, e := l.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Printf("http: Accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0
		c, err := srv.newConn(rw)
		if err != nil {
			continue
		}
		go c.serveSmart()
	}
}
func nextproto3(s *Server, c *tls.Conn, h Handler) {
	z := new(http.Server)
	z.Addr = s.Addr
	z.Handler = s.inner_h
	z.ReadTimeout = s.ReadTimeout
	z.WriteTimeout = s.WriteTimeout
	z.MaxHeaderBytes = s.MaxHeaderBytes
	if z.Handler == nil {
		z.Handler = http.DefaultServeMux
	}
	server_session := spdy.NewServerSession(c, z)
	server_session.Serve()
}

func checkspdy(initial_bytes []byte) bool {
	if len(initial_bytes) < start_bytes {
		return false
	}
	if (int(initial_bytes[0]) == 128) && (int(initial_bytes[1]) == 3) && (int(initial_bytes[2]) == 0) && (int(initial_bytes[3]) == 1) {
		return true
	}
	return false
}

func (c *conn) serveSmart() {
	defer func() {
		if err := recover(); err != nil {
			const size = 4096
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("http: panic serving %v: %v\n%s", c.remoteAddr, err, buf)
		}
		if !c.hijacked() {
			c.close()
		}
	}()

	if tlsConn, ok := c.rwc.(*tls.Conn); ok {
		if d := c.server.ReadTimeout; d != 0 {
			c.rwc.SetReadDeadline(time.Now().Add(d))
		}
		if d := c.server.WriteTimeout; d != 0 {
			c.rwc.SetWriteDeadline(time.Now().Add(d))
		}
		if err := tlsConn.Handshake(); err != nil {
			return
		}
		c.tlsState = new(tls.ConnectionState)
		*c.tlsState = tlsConn.ConnectionState()
		if proto := c.tlsState.NegotiatedProtocol; validNPN(proto) {
			if fn := c.server.TLSNextProto[proto]; fn != nil {
				h := initNPNRequest{tlsConn, serverHandler{c.server}}
				fn(c.server, tlsConn, h)
			}
			return
		}
	}
	initial_bytes := make([]byte, start_bytes)
	// add smart detection
	n, err := c.rwc.Read(initial_bytes)
	if err != nil || n != start_bytes {
		return
	}

	if checkspdy(initial_bytes) {
		s := c.server
		z := new(http.Server)
		z.Addr = s.Addr
		z.Handler = s.inner_h
		z.ReadTimeout = s.ReadTimeout
		z.WriteTimeout = s.WriteTimeout
		z.MaxHeaderBytes = s.MaxHeaderBytes
		server_session := spdy.NewServerSession(NewSmartConn(c.rwc, initial_bytes), z)
		server_session.Serve()
		return
	}

	c2, err := c.server.newConn(NewSmartConn(c.rwc, initial_bytes))
	if err != nil {
		return
	}
	c2.serve()
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
