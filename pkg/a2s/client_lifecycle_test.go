package a2s

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestClientLifecycle(t *testing.T) {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("listen UDP fixture: %v", err)
	}
	defer listener.Close()

	address := listener.LocalAddr().(*net.UDPAddr)
	client, err := NewWithAddr(address, WithBufferSize(2048), WithTimeout(2*time.Second))
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	gotAddress := client.Addr()
	if gotAddress == nil || gotAddress.String() != address.String() {
		t.Fatalf("client address = %v, want %v", gotAddress, address)
	}
	gotAddress.IP[0] = 0
	if client.Addr().String() != address.String() {
		t.Fatal("Addr returned the client's mutable internal address")
	}
	if got, want := client.BufferSize(), uint16(2048); got != want {
		t.Fatalf("buffer size = %d, want %d", got, want)
	}
	if got, want := client.Timeout(), 2*time.Second; got != want {
		t.Fatalf("timeout = %s, want %s", got, want)
	}

	if err := client.Close(); err != nil {
		t.Fatalf("first Close returned error: %v", err)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("second Close returned error: %v", err)
	}
	if _, _, _, err := client.Get(context.Background(), InfoRequest); !errors.Is(err, ErrClientClosed) {
		t.Fatalf("Get after Close error = %v, want ErrClientClosed", err)
	}
}

func TestClientRejectsInvalidAddressesAndOptions(t *testing.T) {
	if _, err := New("", 27015); !errors.Is(err, ErrInvalidAddress) {
		t.Fatalf("empty host error = %v, want ErrInvalidAddress", err)
	}
	if _, err := New("127.0.0.1", 0); !errors.Is(err, ErrInvalidAddress) {
		t.Fatalf("zero port error = %v, want ErrInvalidAddress", err)
	}

	if _, err := NewWithAddr(nil); !errors.Is(err, ErrInvalidAddress) {
		t.Fatalf("nil address error = %v, want ErrInvalidAddress", err)
	}

	invalidAddress := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)}
	if _, err := NewWithAddr(invalidAddress); !errors.Is(err, ErrInvalidAddress) {
		t.Fatalf("zero port error = %v, want ErrInvalidAddress", err)
	}

	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("listen UDP fixture: %v", err)
	}
	defer listener.Close()

	address := listener.LocalAddr().(*net.UDPAddr)
	if _, err := NewWithAddr(address, WithTimeout(0)); !errors.Is(err, ErrInvalidTimeout) {
		t.Fatalf("zero timeout error = %v, want ErrInvalidTimeout", err)
	}
	if _, err := NewWithAddr(address, WithBufferSize(0)); !errors.Is(err, ErrInvalidBufferSize) {
		t.Fatalf("zero buffer error = %v, want ErrInvalidBufferSize", err)
	}
}

func TestNewResolvesIPAndHostname(t *testing.T) {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("listen UDP fixture: %v", err)
	}
	defer listener.Close()

	port := listener.LocalAddr().(*net.UDPAddr).Port
	for _, host := range []string{"127.0.0.1", "localhost"} {
		t.Run(host, func(t *testing.T) {
			client, err := New(host, port)
			if err != nil {
				t.Fatalf("New(%q, %d) returned error: %v", host, port, err)
			}
			defer client.Close()

			if client.Addr() == nil {
				t.Fatal("New returned a client without an address")
			}
		})
	}
}
