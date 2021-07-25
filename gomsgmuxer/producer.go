package main

import (
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// ProducerInterface definition
type ProducerInterface interface {
	Produce() bool
}

// InMemProducer definition
type InMemProducer struct {
	instid    int
	totalmsgs int
	pname     string
	mq        *MsgQueue
	mch       chan MsgNode
	wg        *sync.WaitGroup
}

// NewInMemProducer creates a InMemProducer
func NewInMemProducer(id, nmsgs int, mq *MsgQueue, wg *sync.WaitGroup) *InMemProducer {
	return &InMemProducer{
		instid:    id,
		totalmsgs: nmsgs,
		pname:     "producer_" + strconv.Itoa(id),
		mq:        mq,
		mch:       make(chan MsgNode),
		wg:        wg,
	}
}

// Produce creates and sends the message through msgs channel
func (p *InMemProducer) Produce() bool {
	defer p.wg.Done()
	procid := os.Getpid()
	tstamp := time.Now().UTC().String()
	log.Printf("%s: Pid[%d] %s begins...\n", tstamp, procid, p.pname)
	p.mq.RegisterProducer(p.pname, p.mch)

	for i := 0; i < p.totalmsgs; i++ {
		utctime := time.Now().UTC()
		utctimestr := utctime.String()
		utctimens := utctime.UnixNano()
		msgstr := "hello from " + p.pname + " [" + strconv.Itoa(i) + "] @ " + utctimestr
		mnode := NewMsgNode(utctimens, msgstr)
		p.mch <- *mnode
		if i%10 == 0 {
			// log.Printf("%s giving up\n", p.pname)
			runtime.Gosched()
		}
	}
	log.Printf("produce for %s : Done\n", p.pname)
	return true
}

// Close and unregister with queue
func (p *InMemProducer) Close() {
	close(p.mch)
	p.mq.UnregisterProducer(p.pname)
}
