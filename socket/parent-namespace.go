package socket

import (
	"errors"
	"strconv"
	"sync/atomic"

	"github.com/edelbrocken/socket.io/parser"
)

var count uint64 = 0

type ParentNamespace struct {
	*Namespace

	children *Set
}

func NewParentNamespace(server *Server) *ParentNamespace {
	p := &ParentNamespace{}
	p.Namespace = NewNamespace(server, "/_"+strconv.FormatUint(atomic.AddUint64(&count, 1), 10))
	p.children = NewSet()
	p._initAdapter()

	return p
}

func (p *ParentNamespace) _initAdapter() {
	broadcast := func(packet *parser.Packet, opts *BroadcastOptions) {
		for _, nsp := range p.children.Keys() {
			nsp.adapter.Broadcast(packet, opts)
		}
	}
	p.adapter.SetBroadcast(broadcast)
}

func (p *ParentNamespace) Emit(ev string, args ...interface{}) error {
	for _, nsp := range p.children.Keys() {
		nsp.Emit(ev, args...)
	}
	return nil
}

func (p *ParentNamespace) CreateChild(name string) *Namespace {
	namespace := NewNamespace(p.server, name)

	namespace._fns_mu.RLock()
	namespace._fns = append([]func(*Socket, func(*ExtendedError)){}, p._fns...)
	namespace._fns_mu.RUnlock()

	namespace.AddListener("connect", p.Listeners("connect")...)
	namespace.AddListener("connection", p.Listeners("connection")...)
	p.children.Add(namespace)
	p.server._nsps.Store(name, namespace)
	return namespace
}

func (p *ParentNamespace) FetchSockets() ([]*RemoteSocket, error) {
	// note: we could make the fetchSockets() method work for dynamic namespaces created with a regex (by sending the
	// regex to the other Socket.IO servers, and returning the sockets of each matching namespace for example), but
	// the behavior for namespaces created with a function is less clear
	// note²: we cannot loop over each children namespace, because with multiple Socket.IO servers, a given namespace
	// may exist on one node but not exist on another (since it is created upon client connection)
	return nil, errors.New("FetchSockets() is not supported on parent namespaces")
}
