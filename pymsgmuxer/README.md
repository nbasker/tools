# Python Asynchronous Message Multiplexer

This is a python based simple Multi Input Multi Output message queue implementation.
* mq.py defines the Message Node that encapsulates the message and Message Queue that implements a singly linked list
* consumer.py defines a sample consumer that uses Message Queue
* producer.py defines a sample producer using Message Queue
* requirements.txt specifies the dependent python packages to be implement in a virtual environment


The message queue (mq.py) is implemented as a singly linked list. The list grows and shrinks based on the production and consumption rate of the messages. All consumed messages at a given time that is beyond the configurable retention time is deleted. Linked list is a suitable data structure for insertion and deletion. It is done in constant time. Every consumer's last consumed message is tracked independently.

The producers and consumers are given the message queue that they need to use. Upon completion consumers unregister and clean up the tracking pointers. The message queue serializes the messages and ensures that ordering is maintained for a given producer. Every message produced shall be replicated and given to all consumers.

Each producer, consumer operates as an async-coroutine. The message queue cleaner is also a separate coroutine. The coroutines does cooperative execution by calling await operations. Since there is coperative scheduling, explicit lock protection of the message queue is not required.

The mbrok.py implements a test asyncio session that installs producers, consumers and message-queue. The producers and consumers are shuffled and started. Producer push messages into the queue and the consumers read them from the queue. The producers and consumers can operate independently and in this case implemented as a coroutine. 

## Usage

The mbrok.py shows an example of how the message queue can be used. It is recommended that virtual environment and python3 are used to execute. Once in the virutual environment, the command "python3 mbrok.py" should start the execution.
