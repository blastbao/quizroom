package main

import (
	"quizroom/libs/proto"
	"sync"
)

type Room struct {
	id     int32
	rLock  sync.RWMutex
	next   *Channel
	drop   bool
	Online int // dirty read is ok
}

// NewRoom new a room struct, store channel room info.
func NewRoom(id int32) (r *Room) {
	r = new(Room)
	r.id = id
	r.drop = false
	r.next = nil
	r.Online = 0
	return
}

// Put put channel into the room.
func (r *Room) Put(ch *Channel) (err error) {
	r.rLock.Lock()
	if !r.drop {
		if r.next != nil {
			r.next.Prev = ch
		}
		ch.Next = r.next
		ch.Prev = nil
		r.next = ch // insert to header
		r.Online++
	} else {
		err = ErrRoomDroped
	}
	r.rLock.Unlock()
	return
}

// Del delete channel from the room.
func (r *Room) Del(ch *Channel) bool {
	r.rLock.Lock()
	if ch.Next != nil {
		// if not footer
		ch.Next.Prev = ch.Prev
	}
	if ch.Prev != nil {
		// if not header
		ch.Prev.Next = ch.Next
	} else {
		r.next = ch.Next
	}
	r.Online--
	r.drop = (r.Online == 0)
	r.rLock.Unlock()
	return r.drop
}

// Push push msg to the room, if chan full discard it.
func (r *Room) Push(p *proto.Proto) {
	r.rLock.RLock()
	//log.Info("start push msg: %v", string(p.Body))
	//num := 0
	//t1 := time.Now()
	for ch := r.next; ch != nil; ch = ch.Next {
		ch.Push(p)
	}
	//t2 := time.Since(t1)
	//log.Info("end push msg: %v, count: %v, speed: %v", string(p.Body), num, t2)
	r.rLock.RUnlock()
	return
}

// Close close the room.
func (r *Room) Close() {
	r.rLock.RLock()
	for ch := r.next; ch != nil; ch = ch.Next {
		ch.Close()
	}
	r.rLock.RUnlock()
}
