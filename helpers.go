package bluerpc

import (
	"bytes"
	"net"
)

type testConn struct {
	r bytes.Buffer
	w bytes.Buffer
}

func (c *testConn) Read(b []byte) (int, error)  { return c.r.Read(b) }  //nolint:wrapcheck // This must not be wrapped
func (c *testConn) Write(b []byte) (int, error) { return c.w.Write(b) } //nolint:wrapcheck // This must not be wrapped
func (*testConn) Close() error                  { return nil }

func (*testConn) LocalAddr() net.Addr  { return &net.TCPAddr{Port: 0, Zone: "", IP: net.IPv4zero} }
func (*testConn) RemoteAddr() net.Addr { return &net.TCPAddr{Port: 0, Zone: "", IP: net.IPv4zero} }
