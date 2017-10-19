package poolhandler

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/nsqio/go-nsq"
)

// Pool 处理池
type Pool struct {
	handler  nsq.Handler
	msgChan  chan *nsq.Message
	items    []*poolItem
	size     uint64
	msgCount uint64
	sync.Mutex

	adjustDuration time.Duration
	adjustTicker   *time.Ticker

	exitChan    chan int
	exitHandler sync.Once
}

// New will return a pool.
func New(handler nsq.Handler, size uint64) *Pool {
	p := Pool{
		handler:        handler,
		msgChan:        make(chan *nsq.Message),
		size:           size,
		exitChan:       make(chan int),
		adjustDuration: time.Second * 5,
	}
	p.increase(1)
	p.adjustTicker = time.NewTicker(p.adjustDuration)
	go p.loop()
	return &p
}

// HandleMessage 实现handler接口
func (p *Pool) HandleMessage(msg *nsq.Message) error {
	atomic.AddUint64(&(p.msgCount), 1)
	p.msgChan <- msg
	return nil
}

func (p *Pool) increase(delta uint64) {
	if uint64(len(p.items))+delta > p.size {
		return
	}
	items := make([]*poolItem, int(delta))
	for i := 0; uint64(i) < delta; i++ {
		item := newPoolItem(&p.msgChan, p.handler)
		items[i] = item
		go item.loop()
	}
	p.Lock()
	defer p.Unlock()
	p.items = append(p.items, items...)
}

func (p *Pool) decrease(delta uint64) {
	n := len(p.items)
	if uint64(len(p.items)) < delta {
		return
	}
	items := p.items[n-int(delta) : n-1]
	for _, i := range items {
		go i.destroy()
	}
	p.Lock()
	defer p.Unlock()
	p.items = p.items[0 : n-int(delta)]
}

func (p *Pool) adjust() {
	count := p.msgCount
	atomic.AddUint64(&p.msgCount, ^uint64(count-1))
	// TODO: adjust items count
	return
}

func (p *Pool) loop() {
	for {
		select {
		case <-p.adjustTicker.C:
			p.adjust()
		case <-p.exitChan:
			goto EXIT
		}
	}
EXIT:
	p.exitHandler.Do(func() {
		close(p.exitChan)
		wg := sync.WaitGroup{}
		for index := range p.items {
			wg.Add(1)
			go func(index int) {
				p.items[index].destroy()
				wg.Done()
			}(index)
		}
		wg.Wait()
		close(p.msgChan)
		p.adjustTicker.Stop()
	})
}
