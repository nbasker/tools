package main

import (
	"container/list"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
)

// ConsuserInfo tracker
type consumerInfo struct {
	pcount   int
	lastnode *list.Element
	mch      chan MsgNode
}

// Constructor for consumerInfo
func newConsumerInfo(ch chan MsgNode) *consumerInfo {
	return &consumerInfo{
		pcount:   0,
		lastnode: nil,
		mch:      ch,
	}
}

// ProducerInfo tracker
type producerInfo struct {
	pcount int
	mch    chan MsgNode
}

// Constructor for producerInfo
func newProducerInfo(ch chan MsgNode) *producerInfo {
	return &producerInfo{
		pcount: 0,
		mch:    ch,
	}
}

// MsgQueue holding the messages
type MsgQueue struct {
	sync.Mutex
	msglist              *list.List
	retention, cleanfreq int
	bulkcount            int
	seqid, currcnt       uint64
	pinfo                map[string]*producerInfo
	cinfo                map[string]*consumerInfo
}

// NewMsgQueue is constructor for MsgQueue
func NewMsgQueue(retention, cleanfreq int) *MsgQueue {
	return &MsgQueue{
		msglist:   list.New(),
		retention: retention,
		cleanfreq: cleanfreq,
		bulkcount: 10,
		seqid:     0,
		currcnt:   0,
		pinfo:     make(map[string]*producerInfo),
		cinfo:     make(map[string]*consumerInfo),
	}
}

// Put function to add message
func (m *MsgQueue) Put(pid string, mn *MsgNode) bool {
	m.msglist.PushBack(mn)
	m.seqid++
	m.currcnt++
	return true
}

// Get function to read message
func (m *MsgQueue) Get(cid string, ci *consumerInfo) (*MsgNode, bool) {
	// For cid the lastnode is not filled
	if ci.lastnode == nil {
		e := m.msglist.Front()
		if e == nil {
			// log.Printf("Consumer %s, no message in queue\n", cid)
			return nil, false
		}
		ci.lastnode = e
		return e.Value.(*MsgNode), true
	}

	ne := ci.lastnode.Next()

	// If the next node is nil, return nothing
	if ne == nil {
		// log.Printf("Consumer %s, no new messages\n", cid)
		return nil, false
	}

	// Move the CI to next node and return data
	ci.lastnode = ne
	return ne.Value.(*MsgNode), true
}

// RecvFromProducers gets messages and adds to queue
func (m *MsgQueue) RecvFromProducers() {

	procid := os.Getpid()
	tstamp := time.Now().UTC().String()
	log.Printf("%s: Pid[%d] MsgQueue.RecvFromProducers() begins...\n", tstamp, procid)

	for {
		m.Lock()

		nump := len(m.pinfo)
		// log.Printf("Num producers = %d\n", nump)
		cases := make([]reflect.SelectCase, nump+1)
		pids := make([]string, nump+1)
		i := 0
		for pid, pi := range m.pinfo {
			cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(pi.mch), Send: reflect.ValueOf(nil)}
			pids[i] = pid
			i++
		}

		if nump > 0 {
			cases[i] = reflect.SelectCase{Dir: reflect.SelectDefault, Chan: reflect.ValueOf(nil), Send: reflect.ValueOf(nil)}
			pids[i] = "default_timeout"

			for i = 0; i < m.bulkcount; i++ {
				idx, v, ok := reflect.Select(cases)
				// log.Printf("reflect.Select(): idx=%d, ok=%t\n", idx, ok)
				if ok && idx != nump {
					var mnode MsgNode
					mnode = v.Interface().(MsgNode)
					m.Put(pids[idx], &mnode)
					m.pinfo[pids[idx]].pcount++
					// log.Printf("%s: receiving pcount=%d\n", pids[idx], m.pinfo[pids[idx]].pcount)
				}
			}
		}

		m.Unlock()
		// log.Printf("Pid[%d] MsgQueue.RecvFromProducers() msg count = %d\n", procid, m.seqid)
		// time.Sleep(1 * time.Second)
	}
}

// SendToConsumers reads messages from queue and sends to consumers
func (m *MsgQueue) SendToConsumers() {

	procid := os.Getpid()
	tstamp := time.Now().UTC().String()
	log.Printf("%s: Pid[%d] MsgQueue.SendToConsumers() begins...\n", tstamp, procid)

	for {
		m.Lock()
		// From message q, send to consumer channel
		for cid, ci := range m.cinfo {
			// log.Printf("%s: checking for msg to send to consumer\n", cid)
			mnode, ok := m.Get(cid, ci)
			if ok {
				ci.pcount++
				ci.mch <- *mnode
				// log.Printf("%s: sending ccount=%d\n", cid, ci.pcount)
			}
		}

		/***
		log.Printf("Pid[%d] MsgQueue.SendToConsumers() msg count = %d\n", procid, m.seqid)
		for cid, ci := range m.cinfo {
			log.Printf("%s: sent ccount=%d\n", cid, ci.pcount)
		}
		time.Sleep(1 * time.Second)
		***/
		m.Unlock()
	}
}

// RegisterConsumer to the message queue
func (m *MsgQueue) RegisterConsumer(cid string, ch chan MsgNode) bool {
	m.Lock()
	defer m.Unlock()
	_, ok := m.cinfo[cid]
	if ok {
		log.Printf("Consumer %s already registered\n", cid)
		return false
	}
	m.cinfo[cid] = newConsumerInfo(ch)
	log.Printf("Consumer %s registration successful\n", cid)
	return true
}

// UnregisterConsumer to the message queue
func (m *MsgQueue) UnregisterConsumer(cid string) {
	m.Lock()
	defer m.Unlock()
	ci, ok := m.cinfo[cid]
	if ok {
		close(ci.mch)
		delete(m.cinfo, cid)
		log.Printf("Consumer %s unregistration successful\n", cid)
	}
}

// RegisterProducer to the message queue
func (m *MsgQueue) RegisterProducer(pid string, ch chan MsgNode) bool {
	m.Lock()
	defer m.Unlock()
	_, ok := m.pinfo[pid]
	if ok {
		log.Printf("Producer %s already registered\n", pid)
		return false
	}
	m.pinfo[pid] = newProducerInfo(ch)
	log.Printf("Producer %s registration successful\n", pid)
	return true
}

// UnregisterProducer to the message queue
func (m *MsgQueue) UnregisterProducer(pid string) {
	m.Lock()
	defer m.Unlock()
	_, ok := m.pinfo[pid]
	if ok {
		delete(m.pinfo, pid)
		log.Printf("Producer %s unregistration successful\n", pid)
	}
}

// PrintStats gives MsgQueue status
func (m *MsgQueue) PrintStats() {
	log.Println("Msg Queue Status:")
	log.Printf("total = %d, current = %d\n", m.seqid, m.currcnt)
	/***
	for e := m.msglist.Front(); e != nil; e = e.Next() {
		log.Printf("%s\n", e.Value.(*MsgNode).PrintMsgNodeStr())
	}
	***/
	log.Printf("consumer count = %d\n", len(m.cinfo))
	/***
	for cid, ci := range m.cinfo {
		s := ""
		if ci.lastnode != nil {
			s = ci.lastnode.Value.(*MsgNode).PrintMsgNodeStr()
		}
		log.Printf("%s: count=%d %s\n", cid, ci.pcount, s)
	}
	***/
}

// isConsumed checks if all consumers processed a element
func (m *MsgQueue) isConsumed(dnode *list.Element) (bool, string) {
	var cids []string
	cstatus := true
	for cid, ci := range m.cinfo {
		eq := reflect.DeepEqual(ci.lastnode, dnode)
		if eq {
			cstatus = false
			cids = append(cids, cid)
		}
	}

	cidstr := strings.Join(cids, ", ")
	return cstatus, cidstr
}

// CleanQueue cleans consumed nodes from the list
func (m *MsgQueue) CleanQueue() {
	procid := os.Getpid()
	tstamp := time.Now().UTC().String()
	log.Printf("%s: Pid[%d] MsgQueue.CleanQueue() begins...\n", tstamp, procid)

	ticker := time.NewTicker(time.Second * time.Duration(m.cleanfreq))
	for ; true; <-ticker.C {

		tnano := time.Now().UTC().UnixNano()
		rtime := tnano - int64(m.retention)*1000000000
		rcount := 0

		m.Lock()
		for {

			e := m.msglist.Front()
			if e == nil {
				break
			}

			// node timestamp not beyond retention
			if e.Value.(*MsgNode).ts >= rtime {
				break
			}

			// check if node is consumed by all consumers
			allconsume, _ := m.isConsumed(e)
			if !allconsume {
				/*** No Print in lock section
				mnode := e.Value.(*MsgNode)
				log.Printf("%s yet to consume %s", cids, mnode.PrintMsgNodeStr())
				***/
				break
			}

			m.msglist.Remove(e)
			rcount++
			m.currcnt--
		}
		m.Unlock()
		// log.Printf("Removing %d nodes, retention time %d\n", rcount, rtime)
		m.PrintStats()
	}
}
