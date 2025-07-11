package main

import (
	lru "github.com/hashicorp/golang-lru"
)

type fragmentBuffer struct {
	cache *lru.Cache
}

type fragmentBufferEntry struct {
	sequenceNumber  int32
	fragmentsNeeded int
	fragments       map[int][]byte
}

func newFragmentBuffer() *fragmentBuffer {
	var f fragmentBuffer
	f.cache, _ = lru.New(128)
	return &f
}

func (buf *fragmentBuffer) offer(msg reliableFragment) *photonCommand {
	var entry fragmentBufferEntry

	if buf.cache.Contains(msg.sequenceNumber) {
		obj, _ := buf.cache.Get(msg.sequenceNumber)
		entry = obj.(fragmentBufferEntry)
		entry.fragments[int(msg.fragmentNumber)] = msg.data
	} else {
		entry.sequenceNumber = msg.sequenceNumber
		entry.fragmentsNeeded = int(msg.fragmentCount)
		entry.fragments = make(map[int][]byte)
		entry.fragments[int(msg.fragmentNumber)] = msg.data
	}

	if entry.finished() {
		command := entry.make()
		buf.cache.Remove(msg.sequenceNumber)
		return &command
	} else {
		buf.cache.Add(msg.sequenceNumber, entry)
		return nil
	}
}

func (buf fragmentBufferEntry) finished() bool {
	return len(buf.fragments) == buf.fragmentsNeeded
}

func (buf fragmentBufferEntry) make() photonCommand {
	var data []byte

	for i := 0; i < buf.fragmentsNeeded; i++ {
		data = append(data, buf.fragments[i]...)
	}

	return photonCommand{
		commandType:            sendReliableType,
		data:                   data,
		reliableSequenceNumber: buf.sequenceNumber,
	}
}
