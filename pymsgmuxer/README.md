# Python Asynchronous Message Multiplexer

The mbrok.py implements an asyncio session that installs producers, consumers and message-queue. The producers push messages into the queue and the consumers read from the queue. The producers and consumers can operate independently and in this case implemented as a coroutine. 

The message queue (mq.py) is implemented as a singly linked list. The list grows and shrinks based on the production and consumption rate of the messages. At a given point of time, all consumed messages at given time that is beyond a configurable retention time is deleted. Linked list is a suitable data structure for insertion and deletion. Every consumer's last consumed message is tracked independently.

The producers and consumers can come alive and leave by registering and unregistering with the message queue. The message queue serializes the messages and ensures that ordering is maintained for a given producer. Every message produced shall be replicated and given to all consumers.

Each producer, consumer operates as an async-coroutine. The message queue cleaner is also a separate coroutine. The coroutines does cooperative execution by calling await operations. Since there is coperative scheduling, explicit lock protection of the message queue is not required.

## Usage

The producer's produce() function can be implemented by each subclass and can fetch the data from memory, file-system, socket or any other method. Similarly the consumer's consume() function is independent to the operation needed. 

The producer's, consumer's and message-queue can be used in an async session to multiplex the messages.
