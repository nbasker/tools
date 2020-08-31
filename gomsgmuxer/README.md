# Go Message Multiplexer

The mb.go implements a message queue using go routines and List data type
The message queue is implemented as a List data type and cleans up after a retention period.
It is possible to have multiple producers and consumers operating at the sametime.
The consumers and producers can come alive and leave subscriptions at randome times.
The message queue serializes messages from multiple producers and replicates to all consumers.

Locking is applied to protect the list from multiple go routines

## Usage

This can be run as a service or within a service by exposing API to push and pull data.

## Example

Work In Progress
