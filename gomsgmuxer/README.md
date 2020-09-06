# Go Message Multiplexer

The mb.py implements the main function that spawns go routines for producers, consumers and message-queue. The producers push messages into the queue and the consumers read from the queue. The producers and consumers can operate independently and in this case implemented as go-routine. 

The message queue is implemented as a golang's container/list package which implements doubly linked list. The list grows and shrinks based on the production and consumption rate of the messages. At a given point of time, all consumed messages that is beyond a configurable retention time is deleted. Linked list is a suitable data structure for insertion and deletion. Every consumer's last consumed message is tracked independently.

The producers and consumers can come alive and leave by registering and unregistering with the message queue. The message queue serializes the messages and ensures that ordering is maintained for a given producer. Every message produced shall be replicated and given to all consumers.

Each producer, consumer operates as a go-routine. The message queue cleaner is also a separate go-routine. The go-routines are scheduled by go scheduler and can be scheduled whenever the currently executing go-routine blocks. Explicitly locking is implemented for register, unregister, message-put, message-get and message-cleanup operations. 

The communication between go-routines are implemented using channels. The producer register with message-queue and produces messages. If the producer wants to close, it closes the channel and un-registers with the message-queue. Similarly a consumer if it wants to close, it closes the channel to the message-queue and unregisters.

## Usage

The producer's produce() and consumer's consume() functions are implemented interfaces. These are implemented by every producer and consumer.
The implementation can vary and can receive or send data from/to memory, file-system, socket etc..
