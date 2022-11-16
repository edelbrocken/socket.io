package socket

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"engine.io/types"
	"engine.io/utils"
	"socket.io/parser"
)

type BroadcastOperator struct {
	adapter     Adapter
	rooms       *types.Set
	exceptRooms *types.Set
	flags       *BroadcastFlags
}

func NewBroadcastOperator(adapter Adapter, rooms *types.Set, exceptRooms *types.Set, flags *BroadcastFlags) *BroadcastOperator {
	b := &BroadcastOperator{}
	b.adapter = adapter
	if rooms == nil {
		b.rooms = types.NewSet()
	} else {
		b.rooms = rooms
	}
	if exceptRooms == nil {
		b.exceptRooms = types.NewSet()
	} else {
		b.exceptRooms = exceptRooms
	}
	if flags == nil {
		b.flags = &BroadcastFlags{}
	} else {
		b.flags = flags
	}

	return b
}

// Targets a room when emitting.
func (b *BroadcastOperator) To(room ...string) *BroadcastOperator {
	tmp := make([]string, 0, len(room))
	for _, v := range room {
		tmp = append(tmp, string(v))
	}
	rooms := types.NewSet(b.rooms.Keys()...)
	rooms.Add(tmp...)
	return NewBroadcastOperator(b.adapter, rooms, b.exceptRooms, b.flags)
}

// Targets a room when emitting.
func (b *BroadcastOperator) In(room ...string) *BroadcastOperator {
	return b.To(room...)
}

// Excludes a room when emitting.
func (b *BroadcastOperator) Except(room ...string) *BroadcastOperator {
	tmp := make([]string, 0, len(room))
	for _, v := range room {
		tmp = append(tmp, string(v))
	}
	exceptRooms := types.NewSet(b.exceptRooms.Keys()...)
	exceptRooms.Add(tmp...)
	return NewBroadcastOperator(b.adapter, b.rooms, exceptRooms, b.flags)
}

// Sets the compress flag.
func (b *BroadcastOperator) Compress(compress bool) *BroadcastOperator {
	flags := *b.flags
	flags.Compress = compress
	return NewBroadcastOperator(b.adapter, b.rooms, b.exceptRooms, &flags)
}

// Sets a modifier for a subsequent event emission that the event data may be lost if the client is not ready to
// receive messages (because of network slowness or other issues, or because they’re connected through long polling
// and is in the middle of a request-response cycle).
func (b *BroadcastOperator) Volatile() *BroadcastOperator {
	flags := *b.flags
	flags.Volatile = true
	return NewBroadcastOperator(b.adapter, b.rooms, b.exceptRooms, &flags)
}

// Sets a modifier for a subsequent event emission that the event data will only be broadcast to the current node.
func (b *BroadcastOperator) Local() *BroadcastOperator {
	flags := *b.flags
	flags.Local = true
	return NewBroadcastOperator(b.adapter, b.rooms, b.exceptRooms, &flags)
}

// Adds a timeout in milliseconds for the next operation
//
// <pre><code>
//
//	io.Timeout(1000 * time.Millisecond).Emit("some-event", func(args ...any) {
//	  // ...
//	});
//
// </pre></code>
func (b *BroadcastOperator) Timeout(timeout time.Duration) *BroadcastOperator {
	flags := *b.flags
	flags.Timeout = &timeout
	return NewBroadcastOperator(b.adapter, b.rooms, b.exceptRooms, &flags)
}

// Emits to all clients.
func (b *BroadcastOperator) Emit(ev string, args ...interface{}) error {
	if SOCKET_RESERVED_EVENTS.Has(ev) {
		return errors.New(fmt.Sprintf(`"%s" is a reserved event name`, ev))
	}
	// set up packet object
	data := append([]interface{}{ev}, args...)
	data_len := len(data)

	packet := &parser.Packet{
		Type: parser.EVENT,
		Data: data,
	}

	ack, withAck := data[data_len-1].(func(error, []interface{}))

	if !withAck {
		b.adapter.Broadcast(packet, &BroadcastOptions{
			Rooms:  b.rooms,
			Except: b.exceptRooms,
			Flags:  b.flags,
		})

		return nil
	}

	packet.Data = data[:data_len-1]

	timedOut := false
	responses := []interface{}{}
	var responsesMu sync.RWMutex
	var timeout time.Duration

	if time := b.flags.Timeout; time != nil {
		timeout = *time
	}

	timer := utils.SetTimeOut(func() {
		timedOut = true
		responsesMu.RLock()
		defer responsesMu.RUnlock()

		ack(errors.New("operation has timed out"), responses)
	}, timeout)

	expectedServerCount := int64(-1)
	actualServerCount := int64(0)
	expectedClientCount := uint64(0)

	checkCompleteness := func() {
		responsesMu.RLock()
		defer responsesMu.RUnlock()

		if !timedOut && expectedServerCount == atomic.LoadInt64(&actualServerCount) && uint64(len(responses)) == atomic.LoadUint64(&expectedClientCount) {
			utils.ClearTimeout(timer)
			ack(nil, responses)
		}
	}

	b.adapter.BroadcastWithAck(packet, &BroadcastOptions{
		Rooms:  b.rooms,
		Except: b.exceptRooms,
		Flags:  b.flags,
	}, func(clientCount uint64) {
		// each Socket.IO server in the cluster sends the number of clients that were notified
		atomic.AddUint64(&expectedClientCount, clientCount)
		atomic.AddInt64(&actualServerCount, 1)
		checkCompleteness()
	}, func(clientResponse ...interface{}) {
		// each client sends an acknowledgement
		responsesMu.Lock()
		responses = append(responses, clientResponse...)
		responsesMu.Unlock()
		checkCompleteness()
	})
	expectedServerCount = b.adapter.ServerCount()
	checkCompleteness()
	return nil
}

// Gets a list of clients.
func (b *BroadcastOperator) AllSockets() (*types.Set, error) {
	if b.adapter == nil {
		return nil, errors.New("No adapter for this namespace, are you trying to get the list of clients of a dynamic namespace?")
	}
	return b.adapter.Sockets(b.rooms), nil
}

// Returns the matching socket instances
func (b *BroadcastOperator) FetchSockets() (remoteSockets []*RemoteSocket) {
	for _, socket := range b.adapter.FetchSockets(&BroadcastOptions{
		Rooms:  b.rooms,
		Except: b.exceptRooms,
		Flags:  b.flags,
	}) {
		if s, ok := socket.(*RemoteSocket); ok {
			remoteSockets = append(remoteSockets, s)
		} else if sd, sd_ok := socket.(SocketDetails); sd_ok {
			remoteSockets = append(remoteSockets, NewRemoteSocket(b.adapter, sd))
		}
	}
	return remoteSockets
}

// Makes the matching socket instances join the specified rooms
func (b *BroadcastOperator) SocketsJoin(room ...string) {
	b.adapter.AddSockets(&BroadcastOptions{
		Rooms:  b.rooms,
		Except: b.exceptRooms,
		Flags:  b.flags,
	}, room)
}

// Makes the matching socket instances leave the specified rooms
func (b *BroadcastOperator) SocketsLeave(room ...string) {
	b.adapter.DelSockets(&BroadcastOptions{
		Rooms:  b.rooms,
		Except: b.exceptRooms,
		Flags:  b.flags,
	}, room)
}

// Makes the matching socket instances disconnect
func (b *BroadcastOperator) DisconnectSockets(status bool) {
	b.adapter.DisconnectSockets(&BroadcastOptions{
		Rooms:  b.rooms,
		Except: b.exceptRooms,
		Flags:  b.flags,
	}, status)
}

type RemoteSocket struct {
	id        string
	handshake *Handshake
	rooms     *types.Set
	data      interface{}

	operator *BroadcastOperator
}

func (r *RemoteSocket) Id() string {
	return r.id
}

func (r *RemoteSocket) Handshake() *Handshake {
	return r.handshake
}

func (r *RemoteSocket) Rooms() *types.Set {
	return r.rooms
}

func (r *RemoteSocket) Data() interface{} {
	return r.data
}

func NewRemoteSocket(adapter Adapter, details SocketDetails) *RemoteSocket {
	r := &RemoteSocket{}

	r.id = details.Id()
	r.handshake = details.Handshake()
	r.rooms = types.NewSet(details.Rooms().Keys()...)
	r.data = details.Data()
	r.operator = NewBroadcastOperator(adapter, types.NewSet(string(r.id)), nil, nil)

	return r
}

func (r *RemoteSocket) Emit(ev string, args ...interface{}) error {
	return r.operator.Emit(ev, args...)
}

// Joins a room.
func (r *RemoteSocket) Join(room ...string) {
	r.operator.SocketsJoin(room...)
}

// Leaves a room.
func (r *RemoteSocket) Leave(room ...string) {
	r.operator.SocketsLeave(room...)
}

// Disconnects this client.
func (r *RemoteSocket) Disconnect(status bool) *RemoteSocket {
	r.operator.DisconnectSockets(status)
	return r
}
