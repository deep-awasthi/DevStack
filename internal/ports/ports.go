package ports

import (
	"fmt"
	"net"
)

type Allocator struct {
	used map[int]bool
}

func NewAllocator() Allocator {
	return Allocator{used: map[int]bool{}}
}

func (a *Allocator) Reserve(preferred int) (int, error) {
	if preferred <= 0 {
		return 0, fmt.Errorf("invalid preferred port %d", preferred)
	}
	port := preferred
	for attempts := 0; attempts < 1000; attempts++ {
		if !a.used[port] && available(port) {
			a.used[port] = true
			return port, nil
		}
		port++
	}
	return 0, fmt.Errorf("could not find an available port near %d\nSolution: stop conflicting services or choose explicit ports in devstack.yml", preferred)
}

func (a *Allocator) Mark(port int) {
	if port > 0 {
		a.used[port] = true
	}
}

func available(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	_ = listener.Close()
	return true
}
