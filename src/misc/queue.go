package misc

import (
    "sync"

	"github.com/bwmarrin/discordgo"
)

type PinQueue struct {
    queue []*PinRequest
    lock *sync.Mutex
    cond *sync.Cond
}
var Queue *PinQueue

func NewQueue() *PinQueue {
    q := &PinQueue{}
    q.queue = make([]*PinRequest, 0, 10)
    q.lock = &sync.Mutex{}
    q.cond = sync.NewCond(q.lock)
    return q
}

func (q *PinQueue) Push(req *PinRequest) {
    q.lock.Lock()
    defer q.lock.Unlock()

    // Append to queue
    q.queue = append(q.queue, req)

    // Signal new change
    q.cond.Signal()
}

func (q *PinQueue) Execute(discord *discordgo.Session) (string, string, error)  {
    q.lock.Lock()

    // Block if queue is empty
    for len(q.queue) == 0 {
        q.cond.Wait()
    }

    // Pop from queue
    top, rest := q.queue[0], q.queue[1:]
    q.queue = rest

    q.lock.Unlock()

    // Execute pin request
    return top.Execute(discord)
}
