# Python Message Multiplexer

The mbrok.py implements a message queue using asyncio constructs.
The message queue is implemented as a linked list and cleans up after a retention period.
It is possible to have multiple producers and consumers operating at the sametime.
The consumers and producers can come alive and leave subscriptions at randome times.
The message queue serializes messages from multiple producers and replicates to all consumers.

Need to add details on lockless feature using asyncio await Python feature.

## Usage

This can be run as a service or within a service by exposing API to push and pull data.

## Example

Work In Progress
