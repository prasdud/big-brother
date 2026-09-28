package check

import (
	"context"
	"net"
	"strconv"
	"testing"

	"github.com/prasdud/big-brother/internal/store"
)

func TestTCPCheck(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()
	_, portStr, _ := net.SplitHostPort(listener.Addr().String())
	port, _ := strconv.ParseInt(portStr, 10, 32)

	c := New()
	res := c.Check(context.Background(), store.Service{
		Type: "tcp", Hostname: "127.0.0.1", Port: port, TimeoutSeconds: 2,
	})
	if !res.Up {
		t.Fatalf("expected up, got %+v", res)
	}

	// A closed port refuses the connection.
	closed, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, closedPortStr, _ := net.SplitHostPort(closed.Addr().String())
	closedPort, _ := strconv.ParseInt(closedPortStr, 10, 32)
	closed.Close()

	res = c.Check(context.Background(), store.Service{
		Type: "tcp", Hostname: "127.0.0.1", Port: closedPort, TimeoutSeconds: 2,
	})
	if res.Up || res.Error == "" {
		t.Fatalf("expected down, got %+v", res)
	}
}

func TestDNSCheck(t *testing.T) {
	c := New()

	res := c.Check(context.Background(), store.Service{
		Type: "dns", Hostname: "localhost", TimeoutSeconds: 2,
	})
	if !res.Up {
		t.Fatalf("expected up for localhost, got %+v", res)
	}

	res = c.Check(context.Background(), store.Service{
		Type: "dns", Hostname: "does-not-exist.invalid", TimeoutSeconds: 2,
	})
	if res.Up || res.Error == "" {
		t.Fatalf("expected down for invalid host, got %+v", res)
	}
}
