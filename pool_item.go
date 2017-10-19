package poolhandler

import (
	"sync"

	"github.com/nsqio/go-nsq"
)

type poolItem struct {
	msgChan        *chan *nsq.Message
	destroyChan    chan int
	handler        nsq.Handler
	destroyHandler sync.Once
}

func newPoolItem(msgChan *chan *nsq.Message, handler nsq.Handler) *poolItem {
	return &poolItem{
		msgChan:     msgChan,
		destroyChan: make(chan int),
		handler:     handler,
	}
}

func (p *poolItem) destroy() {
	p.destroyChan <- 1
}

func (p *poolItem) loop() {
	for {
		select {
		case msg := <-*p.msgChan:
			(p.handler).HandleMessage(msg)
		case <-p.destroyChan:
			goto DESTROY
		}
	}
DESTROY:
	p.destroyHandler.Do(func() {
		close(p.destroyChan)
	})
}
