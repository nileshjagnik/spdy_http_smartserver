// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

// a smart connection

import (
	"bytes"
	"io"
	"net"
	"time"
)

type smart_conn struct {
	inner_conn   net.Conn
	smart_reader io.Reader
}

func (s smart_conn) Read(b []byte) (n int, err error) {
	return s.smart_reader.Read(b)
}
func (s smart_conn) Write(b []byte) (n int, err error) {
	return s.inner_conn.Write(b)
}
func (s smart_conn) Close() (err error) {
	return s.inner_conn.Close()
}
func (s smart_conn) LocalAddr() net.Addr {
	return s.inner_conn.LocalAddr()
}
func (s smart_conn) RemoteAddr() net.Addr {
	return s.inner_conn.RemoteAddr()
}
func (s smart_conn) SetDeadline(t time.Time) error {
	return s.inner_conn.SetDeadline(t)
}
func (s smart_conn) SetWriteDeadline(t time.Time) error {
	return s.inner_conn.SetWriteDeadline(t)
}
func (s smart_conn) SetReadDeadline(t time.Time) error {
	return s.inner_conn.SetReadDeadline(t)
}

func NewSmartConn(c net.Conn, buf []byte) net.Conn {
	return smart_conn{inner_conn: c, smart_reader: io.MultiReader(io.LimitReader(bytes.NewReader(buf), int64(len(buf))), c)}
}
