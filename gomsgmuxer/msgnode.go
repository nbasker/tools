package main

import (
	"fmt"
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
