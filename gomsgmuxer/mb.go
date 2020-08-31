package main

import (
	"container/list"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MsgNode container
type MsgNode struct {
	id   int64
	ts   int64
	data string
}

// NewMsgNode is constructor for MsgNode
func NewMsgNode(ident int64, msg string) *MsgNode {
	tstamp := time.Now().UTC().UnixNano()
	return &MsgNode{
		id:   ident,
		ts:   tstamp,
		data: msg,
	}
}

// PrintMsgNodeStr prints the contents in string
func (mn *MsgNode) PrintMsgNodeStr() string {
	s := fmt.Sprintf("Id:%d ts=%d msg='%s'\n", mn.id, mn.ts, mn.data)
	return s
}

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
			// fmt.Printf("Consumer %s, no message in queue\n", cid)
			return nil, false
		}
		ci.lastnode = e
		return e.Value.(*MsgNode), true
	}

	ne := ci.lastnode.Next()

	// If the next node is nil, return nothing
	if ne == nil {
		// fmt.Printf("Consumer %s, no new messages\n", cid)
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
	fmt.Printf("%s: Pid[%d] MsgQueue.RecvFromProducers() begins...\n", tstamp, procid)

	for {
		m.Lock()
		// From producer channel, read and add to message q
		for pid, pi := range m.pinfo {
			// fmt.Printf("%s: checking for msg from producer...\n", pid)
			var mnode MsgNode
			ok := true
			timedout := false
			select {
			case mnode, ok = <-pi.mch:
			default:
				timedout = true
			}
			if ok {
				if !timedout {
					m.Put(pid, &mnode)
					pi.pcount++
					// fmt.Printf("%s: found, pcount=%d\n", pid, pi.pcount)
				}
			}
		}
		m.Unlock()
		//fmt.Printf("Pid[%d] MsgQueue.RecvFromProducers() msg count = %d\n", procid, m.seqid)
		//time.Sleep(1 * time.Second)
	}
}

// SendToConsumers reads messages from queue and sends to consumers
func (m *MsgQueue) SendToConsumers() {

	procid := os.Getpid()
	tstamp := time.Now().UTC().String()
	fmt.Printf("%s: Pid[%d] MsgQueue.SendToConsumers() begins...\n", tstamp, procid)

	for {
		m.Lock()
		// From message q, send to consumer channel
		for cid, ci := range m.cinfo {
			// fmt.Printf("%s: checking for msg to send to consumer\n", cid)
			mnode, ok := m.Get(cid, ci)
			if ok {
				ci.pcount++
				ci.mch <- *mnode
				// fmt.Printf("%s: sending ccount=%d\n", cid, ci.pcount)
			}
		}

		/***
		fmt.Printf("Pid[%d] MsgQueue.SendToConsumers() msg count = %d\n", procid, m.seqid)
		for cid, ci := range m.cinfo {
			fmt.Printf("%s: sent ccount=%d\n", cid, ci.pcount)
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
		fmt.Printf("Consumer %s already registered\n", cid)
		return false
	}
	m.cinfo[cid] = newConsumerInfo(ch)
	fmt.Printf("Consumer %s registration successful\n", cid)
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
		fmt.Printf("Consumer %s unregistration successful\n", cid)
	}
}

// RegisterProducer to the message queue
func (m *MsgQueue) RegisterProducer(pid string, ch chan MsgNode) bool {
	m.Lock()
	defer m.Unlock()
	_, ok := m.pinfo[pid]
	if ok {
		fmt.Printf("Producer %s already registered\n", pid)
		return false
	}
	m.pinfo[pid] = newProducerInfo(ch)
	fmt.Printf("Producer %s registration successful\n", pid)
	return true
}

// UnregisterProducer to the message queue
func (m *MsgQueue) UnregisterProducer(pid string) {
	m.Lock()
	defer m.Unlock()
	_, ok := m.pinfo[pid]
	if ok {
		delete(m.pinfo, pid)
		fmt.Printf("Producer %s unregistration successful\n", pid)
	}
}

// PrintStats gives MsgQueue status
func (m *MsgQueue) PrintStats() {
	fmt.Println("Msg Queue Status:")
	fmt.Printf("total = %d, current = %d\n", m.seqid, m.currcnt)
	/***
	for e := m.msglist.Front(); e != nil; e = e.Next() {
		fmt.Printf("%s\n", e.Value.(*MsgNode).PrintMsgNodeStr())
	}
	***/
	fmt.Printf("consumer count = %d\n", len(m.cinfo))
	for cid, ci := range m.cinfo {
		s := ""
		if ci.lastnode != nil {
			s = ci.lastnode.Value.(*MsgNode).PrintMsgNodeStr()
		}
		fmt.Printf("%s: count=%d %s\n", cid, ci.pcount, s)
	}
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
	fmt.Printf("%s: Pid[%d] MsgQueue.CleanQueue() begins...\n", tstamp, procid)

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
				fmt.Printf("%s yet to consume %s", cids, mnode.PrintMsgNodeStr())
				***/
				break
			}

			m.msglist.Remove(e)
			rcount++
			m.currcnt--
		}
		m.Unlock()
		fmt.Printf("Removing %d nodes, retention time %d\n", rcount, rtime)
		m.PrintStats()
	}
}

// Consumer definition
type Consumer struct {
	instid   int
	msgcount int
	cname    string
	mq       *MsgQueue
	mch      chan MsgNode
	wg       *sync.WaitGroup
}

// NewConsumer creates a Consumer
func NewConsumer(id int, mq *MsgQueue, wg *sync.WaitGroup) *Consumer {
	return &Consumer{
		instid:   id,
		msgcount: 0,
		cname:    "consumer_" + strconv.Itoa(id),
		mq:       mq,
		mch:      make(chan MsgNode, 2),
		wg:       wg,
	}
}

// Consume reads the msgs channel
func (c *Consumer) Consume() {
	defer c.wg.Done()
	procid := os.Getpid()
	tstamp := time.Now().UTC().String()
	fmt.Printf("%s: Pid[%d] %s begins...\n", tstamp, procid, c.cname)
	c.mq.RegisterConsumer(c.cname, c.mch)

	f, err := os.Create(c.cname)
	if err != nil {
		fmt.Println("Error opening file ", err)
	}

	for m := range c.mch {
		fmt.Fprintf(f, "%s: %s", c.cname, m.PrintMsgNodeStr())
	}
	fmt.Printf("consume for %s : Done\n", c.cname)
}

// Close and unregister with queue
func (c *Consumer) Close() {
	c.mq.UnregisterConsumer(c.cname)
}

// Producer definition
type Producer struct {
	instid    int
	totalmsgs int
	pname     string
	mq        *MsgQueue
	mch       chan MsgNode
	wg        *sync.WaitGroup
}

// NewProducer creates a Producer
func NewProducer(id, nmsgs int, mq *MsgQueue, wg *sync.WaitGroup) *Producer {
	return &Producer{
		instid:    id,
		totalmsgs: nmsgs,
		pname:     "producer_" + strconv.Itoa(id),
		mq:        mq,
		mch:       make(chan MsgNode),
		wg:        wg,
	}
}

// Produce creates and sends the message through msgs channel
func (p *Producer) Produce() {
	defer p.wg.Done()
	procid := os.Getpid()
	tstamp := time.Now().UTC().String()
	fmt.Printf("%s: Pid[%d] %s begins...\n", tstamp, procid, p.pname)
	p.mq.RegisterProducer(p.pname, p.mch)

	for i := 0; i < p.totalmsgs; i++ {
		utctime := time.Now().UTC()
		utctimestr := utctime.String()
		utctimens := utctime.UnixNano()
		msgstr := "hello from " + p.pname + " [" + strconv.Itoa(i) + "] @ " + utctimestr
		mnode := NewMsgNode(utctimens, msgstr)
		p.mch <- *mnode
	}
	fmt.Printf("produce for %s : Done\n", p.pname)
}

// Close and unregister with queue
func (p *Producer) Close() {
	close(p.mch)
	p.mq.UnregisterProducer(p.pname)
}

func main() {
	var pwg sync.WaitGroup
	var cwg sync.WaitGroup

	mq := NewMsgQueue(30, 10)
	pwg.Add(1)
	pro1 := NewProducer(1, 25, mq, &pwg)
	pwg.Add(1)
	pro2 := NewProducer(2, 15, mq, &pwg)
	pwg.Add(1)
	pro3 := NewProducer(3, 5, mq, &pwg)

	cwg.Add(1)
	con1 := NewConsumer(1, mq, &cwg)
	cwg.Add(1)
	con2 := NewConsumer(2, mq, &cwg)
	cwg.Add(1)
	con3 := NewConsumer(3, mq, &cwg)

	fmt.Println("Starting MsgQueue go routines...")
	go mq.RecvFromProducers()
	go mq.SendToConsumers()
	go mq.CleanQueue()

	fmt.Println("Starting regular producer/consumer go routines...")
	go pro1.Produce()
	go pro2.Produce()
	go pro3.Produce()
	go con1.Consume()
	go con2.Consume()
	go con3.Consume()

	fmt.Println("Waiting for all producers to complete...")
	pwg.Wait()
	fmt.Println("Producers done, waiting for 20 seconds for consumers process...")
	pro1.Close()
	pro2.Close()
	pro3.Close()
	time.Sleep(60 * time.Second)
	con1.Close()
	con2.Close()
	con3.Close()
	cwg.Wait()
	fmt.Println("Consumers done...")

	fmt.Println("Printing queue status and ending...")
	time.Sleep(2 * time.Second)
	mq.PrintStats()
	fmt.Println("Ending message broker")
}
