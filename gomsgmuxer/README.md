# Go Message Multiplexer

This is a simiple Multi-Input-Multi-Output message queue implementation.
* msgnode.go defines the structure for each Message
* mq.go is the MessageQueue implementation supporting multiple producers and consumers
* producer.go is a sample message producer
* consumer.go is a sample message consumer
* mb.go provides a sample of how the message queue can be used with multiple producers and consumers

The mb.go is the function that puts together all the modules and performs the test of message queue data structure. It spawns go routines for producers, consumers and message-queue. The producers push messages into the queue and the consumers read from the queue. The producers and consumers can operate independently. The message queue has a channel with each consume and producer and processes them indepdently. All these are implemented as go-routine. 

The message queue is built using golang's container/list package which implements doubly linked list. The list grows and shrinks based on the production and consumption rate of the messages. At a given point of time, all consumed messages that is beyond a configurable retention time is deleted. Linked list is a suitable data structure for insertion and deletion. Every consumer's last consumed message is tracked independently.

The producers and consumers can come dynamically register and unregister with the message queue. The message queue serializes the messages and ensures that ordering is maintained for a given producer. Every message produced is replicated and given to all consumers. Only one copy of the message is maintained in memory within the message queue, when consumer is ready to consume a copy is made and sent to its respective channel.

Each producer, consumer operates as a go-routine. The message queue cleaner is an independent and separate go-routine. The go-routines are scheduled by go scheduler and can be scheduled whenever the currently executing go-routine blocks. Explicitly locking is implemented for register, unregister, message-put, message-get and message-cleanup operations.

The communication between go-routines are implemented using channels. The producer register with message-queue and produces messages. If the producer wants to stop and close, it closes the channel and un-registers from the message-queue. Similarly a consumer if it wants to close, it closes the channel to the message-queue and unregisters.

## Usage

The mb.qo gives an example of how the message queue can be used in conjunction with producer and consumers.
This can be executed using the following command, upon execution there are messages put out that trace the execution flow.

go build -o msgbroker
./msgbroker
