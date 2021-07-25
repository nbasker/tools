package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

// ConsumerInterface definition
type ConsumerInterface interface {
	Consume() bool
}

// ConsumerFileWriter definition
type ConsumerFileWriter struct {
	instid   int
	msgcount int
	cname    string
	mq       *MsgQueue
	mch      chan MsgNode
	wg       *sync.WaitGroup
}

// NewConsumerFileWriter creates a Consumer
func NewConsumerFileWriter(id int, mq *MsgQueue, wg *sync.WaitGroup) *ConsumerFileWriter {
	return &ConsumerFileWriter{
		instid:   id,
		msgcount: 0,
		cname:    "consumer_" + strconv.Itoa(id),
		mq:       mq,
		mch:      make(chan MsgNode, 2),
		wg:       wg,
	}
}

// Consume reads the msgs channel
func (c *ConsumerFileWriter) Consume() bool {
	defer c.wg.Done()
	procid := os.Getpid()
	tstamp := time.Now().UTC().String()
	log.Printf("%s: Pid[%d] %s begins...\n", tstamp, procid, c.cname)
	c.mq.RegisterConsumer(c.cname, c.mch)

	f, err := os.Create(c.cname)
	if err != nil {
		log.Println("Error opening file ", err)
	}

	for m := range c.mch {
		fmt.Fprintf(f, "%s: %s", c.cname, m.PrintMsgNodeStr())
	}
	log.Printf("consume for %s : Done\n", c.cname)
	return true
}

// Close and unregister with queue
func (c *ConsumerFileWriter) Close() {
	c.mq.UnregisterConsumer(c.cname)
}
