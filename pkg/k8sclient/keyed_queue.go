/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package k8sclient

import (
	"sync"
)

// Queue is an interface which abstracts a queue.
type Queue interface {
	// Add enqueues a key and its associated item.
	Add(key interface{}, item interface{})
	// Get removes an item from the queue and insets the item to the currently processing key set.
	Get() (key interface{}, items []interface{}, shutdown bool)
	// Done removes the item under processing.
	Done(key interface{})
	// ShutDown shuts down the queue.
	ShutDown()
	// ShuttingDown tests if the queue is shutting down.
	ShuttingDown() bool
}

type tk interface{}

// NewKeyedQueue initializes a queue.
func NewKeyedQueue() *Type {
	return &Type{
		items:        map[tk][]interface{}{},
		toQueue:      map[tk][]interface{}{},
		processing:   set{},
		shuttingDown: false,
		cond:         sync.NewCond(&sync.Mutex{}),
	}
}

// Type implements the Queue interface.
type Type struct {
	// Queue of keys to be processed.
	queue []tk
	// Items to be processed associated with each key.
	items map[tk][]interface{}
	// Items to be queued once the keys are processed.
	toQueue map[tk][]interface{}
	// Set of keys currently under processing.
	processing set
	// shuttingDown is the flag representing if the queue is shutting down.
	shuttingDown bool
	cond         *sync.Cond
}

type empty struct{}
type set map[tk]empty

func (s set) has(item tk) bool {
	_, exists := s[item]
	return exists
}

func (s set) insert(item tk) {
	s[item] = empty{}
}

func (s set) delete(item tk) {
	delete(s, item)
}

// Add enqueues a key and its associated item.
func (q *Type) Add(key interface{}, item interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	if q.shuttingDown {
		return
	}
	if q.processing.has(key) {
		// Key is under processing. Can not add it to the queue.
		q.toQueue[key] = append(q.toQueue[key], item)
		return
	}
	if items, ok := q.items[key]; ok {
		// Key already exists in the queue. Don't have to signal.
		q.items[key] = append(items, item)
	} else {
		// New key in the queue. Send signal.
		q.items[key] = append(q.items[key], item)
		q.queue = append(q.queue, key)
		q.cond.Signal()
	}
}

// Get removes an item from the queue and inserts the item to the currently processing key set.
func (q *Type) Get() (key interface{}, items []interface{}, shutdown bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	for len(q.queue) == 0 && !q.shuttingDown {
		q.cond.Wait()
	}
	if len(q.queue) == 0 {
		// We must be shutting down.
		return nil, nil, true
	}
	key, q.queue = q.queue[0], q.queue[1:]
	// Add key to the processing set.
	q.processing.insert(key)
	items = q.items[key]
	delete(q.items, key)
	return key, items, false
}

// Done removes the item under processing and put the queued item into the to-be-processed set.
func (q *Type) Done(key interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	q.processing.delete(key)
	items, ok := q.toQueue[key]
	if ok {
		q.queue = append(q.queue, key)
		q.items[key] = items
		delete(q.toQueue, key)
		q.cond.Signal()
	}
}

// ShutDown shuts down the queue.
// After ShutDown is called new items will not be appended to the queue. Only
// already appended items will be drained.
func (q *Type) ShutDown() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	q.shuttingDown = true
	q.cond.Broadcast()
}

// ShuttingDown tests if the queue is shutting down.
func (q *Type) ShuttingDown() bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return q.shuttingDown
}
