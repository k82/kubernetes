/*
Copyright 2017 The Kubernetes Authors.

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

package util

import "container/heap"

type PriorityQueueValue interface {
	Priority() float64
}

type PriorityQueue struct {
	queue priorityQueue
}

func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		queue: priorityQueue{},
	}
}

func (pq *PriorityQueue) Push(x PriorityQueueValue) {
	item := &item{
		Value:    x,
		Priority: x.Priority(),
	}

	heap.Push(&pq.queue, item)
}

func (pq *PriorityQueue) Pop() PriorityQueueValue {
	item := heap.Pop(&pq.queue).(*item)
	return item.Value
}

func (pq *PriorityQueue) Len() int {
	return pq.queue.Len()
}

// ----> Internal data structure <-------

type item struct {
	Value    PriorityQueueValue // The value of the item; arbitrary.
	Priority float64            // The priority of the item in the queue.
	Index    int                // The index of the item in the heap.
}

type priorityQueue []*item

func (pq priorityQueue) Len() int {
	return len(pq)
}

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].Priority < pq[j].Priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*item)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.Index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}
