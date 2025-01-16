package misc

import (
    "sync"

	"github.com/bwmarrin/discordgo"
)

type PinQueue struct {
    queue []*PinRequest
    lock sync.Mutex
}
var Queue *PinQueue

func (q *PinQueue) Push(req *PinRequest) {
    q.lock.Lock()
    defer q.lock.Unlock()

    // Append to queue
    q.queue = append(q.queue, req)
}

func (q *PinQueue) Execute(discord *discordgo.Session) (string, string, error)  {
    q.lock.Lock()
    defer q.lock.Unlock()

    // Pop from queue
    top, rest := q.queue[0], q.queue[1:]
    q.queue = rest

    // Execute pin request
    return top.Execute(discord)
}
