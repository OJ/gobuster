package libgobuster

import "sync"

type MessageLevel int

const (
	LevelDebug MessageLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Message struct {
	Level   MessageLevel
	Message string
}

type Progress struct {
	requestsExpectedMutex *sync.RWMutex
	requestsExpected      int
	requestsCountMutex    *sync.RWMutex
	requestsIssued        int
	ResultChan            chan Result
	ErrorChan             chan error
	MessageChan           chan Message
}

func NewProgress() *Progress {
	var p Progress
	p.requestsIssued = 0
	p.requestsExpectedMutex = new(sync.RWMutex)
	p.requestsCountMutex = new(sync.RWMutex)
	p.ResultChan = make(chan Result)
	p.ErrorChan = make(chan error)
	p.MessageChan = make(chan Message)
	return &p
}

func (p *Progress) RequestsExpected() int {
	p.requestsExpectedMutex.RLock()
	defer p.requestsExpectedMutex.RUnlock()
	return p.requestsExpected
}

func (p *Progress) RequestsIssued() int {
	p.requestsCountMutex.RLock()
	defer p.requestsCountMutex.RUnlock()
	return p.requestsIssued
}

func (p *Progress) incrementRequestsIssues(by int) {
	p.requestsCountMutex.Lock()
	defer p.requestsCountMutex.Unlock()
	p.requestsIssued += by
}

func (p *Progress) incrementRequests() {
	p.requestsCountMutex.Lock()
	defer p.requestsCountMutex.Unlock()
	p.requestsIssued++
}

func (p *Progress) IncrementTotalRequests(by int) {
	p.requestsCountMutex.Lock()
	defer p.requestsCountMutex.Unlock()
	p.requestsExpected += by
}
