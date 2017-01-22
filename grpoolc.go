package grpoolc

import (
	"fmt"
	"google.golang.org/grpc"
	"math/rand"
	"sync"
	"time"
)

type newConn func() (*grpc.ClientConn, error)

type grpcPool struct {
	connections         []*grpc.ClientConn
	connectionGenerator newConn
	maxConnections      int
}

var m sync.RWMutex

var pools = make(map[string]*grpcPool)

func New(descriptor string, generatorFunc newConn, maxConn int) error {
	var e error
	m.Lock()
	defer m.Unlock()
	if _, ok := pools[descriptor]; ok != true {
		pools[descriptor] = &grpcPool{connectionGenerator: generatorFunc, connections: make([]*grpc.ClientConn, 0), maxConnections: maxConn}
	} else {
		e = fmt.Errorf("Pool %s already exists", descriptor)
	}
	return e
}

func Get(descriptor string) (*grpc.ClientConn, error) {
	var e error
	var c *grpc.ClientConn
	m.Lock()
	defer m.Unlock()
	if g, ok := pools[descriptor]; ok {
		c, e = g.get()
	} else {
		e = fmt.Errorf("Pool %s does not exist", descriptor)
	}
	return c, e
}

func Put(descriptor string, conn *grpc.ClientConn) error {
	var e error
	m.Lock()
	defer m.Unlock()
	if g, ok := pools[descriptor]; ok {
		g.put(conn)
	} else {
		e = fmt.Errorf("Pool %s does not exist", descriptor)
	}
	return e
}

func Close(descriptor string) error {
	var e error
	m.Lock()
	defer m.Unlock()
	if g, ok := pools[descriptor]; ok {
		g.close()
		delete(pools, descriptor)
	} else {
		e = fmt.Errorf("Pool %s does not exist", descriptor)
	}
	return e
}

func (g *grpcPool) get() (*grpc.ClientConn, error) {
	var c *grpc.ClientConn
	var e error
	if len(g.connections) > 0 {
		c, g.connections = g.connections[0], g.connections[1:]
		rand.Seed(time.Now().UnixNano())
		r := rand.Intn(100)
		if r < 5 {
			c.Close()
			c, e = g.connectionGenerator()
		}
	} else {
		c, e = g.connectionGenerator()
	}
	return c, e
}

func (g *grpcPool) put(conn *grpc.ClientConn) {
	if g.maxConnections == 0 || len(g.connections) < g.maxConnections {
		g.connections = append(g.connections, conn)
	} else {
		conn.Close()
	}

}

func (g *grpcPool) close() {
	for _, c := range g.connections {
		c.Close()
	}
	g.connections = make([]*grpc.ClientConn, 0)
}
