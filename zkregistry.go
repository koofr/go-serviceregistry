package serviceregistry

import (
	"fmt"
	"git.koofr.lan/go-zkutils.git"
	zk "launchpad.net/gozk"
	"sync"
	"time"
)

var ZK_PERM = zk.WorldACL(zk.PERM_ALL)

type entry struct {
	service  string
	protocol string
	server   string
}

type ZkRegistry struct {
	zkServers    string
	zk           *zk.Conn
	zkSession    <-chan zk.Event
	zkMutex      sync.RWMutex
	closed       bool
	entries      []*entry
	entriesMutex sync.RWMutex
}

func NewZkRegistry(zkServers string) (r *ZkRegistry, err error) {
	r = &ZkRegistry{
		zkServers: zkServers,
		entries:   make([]*entry, 0),
	}

	err = r.zkConnect()

	// connected, make sure it stays connected
	if err == nil {
		go r.zkKeepAlive()
	}

	return
}

func (r *ZkRegistry) zkConnect() (err error) {
	z, session, err := zk.Dial(r.zkServers, 10*time.Second)

	if err != nil {
		err = fmt.Errorf("ZkRegistry connection failed: %s", err)
		return
	}

	// Wait for connection.
	event := <-session

	if event.State != zk.STATE_CONNECTED {
		err = fmt.Errorf("ZkRegistry connection failed: %v", event)
		return
	}

	r.zk = z
	r.zkSession = session

	return
}

func (r *ZkRegistry) zkReconnect() (err error) {
	r.zkMutex.Lock()
	defer r.zkMutex.Unlock()

	r.zk.Close()

	r.zkSession = nil

	err = r.zkConnect()

	return
}

func (r *ZkRegistry) zkKeepAlive() {
	for {
		// if reconnect failed, zkSession is nil
		if r.zkSession != nil {
			// wait for error/disconnect
			_ = <-r.zkSession
		}

		// stop if closed manually
		if r.closed {
			break
		}

		err := r.zkReconnect()

		if err != nil {
			// could not reconnect, maybe wait for a while?
			fmt.Println(err)
		} else {
			r.reregister()
		}
	}
}

func (r *ZkRegistry) pathParts(service string, protocol string) []string {
	return []string{"services", service, protocol}
}

func (r *ZkRegistry) register(e *entry) (err error) {
	r.zkMutex.RLock()
	defer r.zkMutex.RUnlock()

	pathParts := r.pathParts(e.service, e.protocol)

	err = zkutils.EnsurePath(r.zk, pathParts, ZK_PERM)

	if err != nil {
		err = fmt.Errorf("ZkRegistry register ensure path error: %s", err)
		return
	}

	path := zkutils.BuildPath(pathParts) + "/"

	_, err = r.zk.Create(path, e.server, zk.EPHEMERAL|zk.SEQUENCE, ZK_PERM)

	if err != nil {
		err = fmt.Errorf("ZkRegistry register create error: %s", err)
		return
	}

	return
}

func (r *ZkRegistry) reregister() {
	r.entriesMutex.RLock()
	defer r.entriesMutex.RUnlock()

	for _, e := range r.entries {
		err := r.register(e)

		if err != nil {
			fmt.Printf("ZkRegistry %s could not be reregistered\n", e)
		}
	}
}

func (r *ZkRegistry) Register(service string, protocol string, server string) (err error) {
	e := &entry{
		service:  service,
		protocol: protocol,
		server:   server,
	}

	err = r.register(e)

	if err == nil {
		r.entriesMutex.Lock()
		defer r.entriesMutex.Unlock()

		r.entries = append(r.entries, e)
	}

	return
}

func (r *ZkRegistry) Get(service string, protocol string) (servers []string, err error) {
	r.zkMutex.RLock()
	defer r.zkMutex.RUnlock()

	pathParts := r.pathParts(service, protocol)

	path := zkutils.BuildPath(pathParts)

	children, _, err := r.zk.Children(path)

	if err != nil {
		if zkErr, ok := err.(*zk.Error); ok && zkErr.Code == zk.ZNONODE {
			// path doesn't exist yet
			servers = []string{}
			err = nil
			return
		} else {
			err = fmt.Errorf("ZkRegistry get error: %s", err)
			return
		}
	}

	servers = make([]string, len(children))

	var data string

	for i, child := range children {
		data, _, err = r.zk.Get(path + "/" + child)

		if err != nil {
			return
		}

		servers[i] = data
	}

	return
}

func (r *ZkRegistry) Close() error {
	r.zkMutex.Lock()
	defer r.zkMutex.Unlock()

	r.closed = true

	return r.zk.Close()
}
