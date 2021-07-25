package main

import (
	"log"
	"sync"
	"time"
)

func main() {
	var pwg sync.WaitGroup // producer wait group
	var cwg sync.WaitGroup // consumer wait group

	mq := NewMsgQueue(30, 10)
	pwg.Add(1)
	pro1 := NewInMemProducer(1, 25, mq, &pwg)
	pwg.Add(1)
	pro2 := NewInMemProducer(2, 15, mq, &pwg)
	pwg.Add(1)
	pro3 := NewInMemProducer(3, 5, mq, &pwg)

	cwg.Add(1)
	con1 := NewConsumerFileWriter(1, mq, &cwg)
	cwg.Add(1)
	con2 := NewConsumerFileWriter(2, mq, &cwg)
	cwg.Add(1)
	con3 := NewConsumerFileWriter(3, mq, &cwg)

	log.Println("Starting MsgQueue go routines...")
	go mq.RecvFromProducers()
	go mq.SendToConsumers()
	go mq.CleanQueue()

	log.Println("Starting regular producer/consumer go routines...")
	go con1.Consume()
	go con2.Consume()
	go con3.Consume()
	go pro1.Produce()
	go pro2.Produce()
	go pro3.Produce()

	log.Println("Waiting for all producers to complete...")
	pwg.Wait()
	pro1.Close()
	pro2.Close()
	pro3.Close()
	log.Println("Producers closed, waiting for for consumers to process...")
	time.Sleep(60 * time.Second)
	con1.Close()
	con2.Close()
	con3.Close()
	cwg.Wait()
	log.Println("Consumers done...")

	log.Println("Printing queue status and ending...")
	time.Sleep(2 * time.Second)
	mq.PrintStats()
	log.Println("Ending message broker")
}
