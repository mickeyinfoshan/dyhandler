package poolhandler

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/nsqio/go-nsq"
)

func TestPool_increase(t *testing.T) {
	type fields struct {
		handler        nsq.Handler
		msgChan        chan *nsq.Message
		items          []*poolItem
		size           uint64
		msgCount       uint64
		Mutex          sync.Mutex
		adjustDuration time.Duration
		adjustTicker   *time.Ticker
		exitChan       chan int
		exitHandler    sync.Once
	}
	type args struct {
		delta uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "increase should add goroutines",
			fields: fields{
				size:    20,
				handler: testHandler{},
			},
			args: args{
				delta: 10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.fields.handler, tt.fields.size)
			beforeIncrease := runtime.NumGoroutine()
			wg := sync.WaitGroup{}
			for i := 0; i < int(tt.args.delta); i++ {
				wg.Add(1)
				go func() { // test concurrencty
					p.increase(1)
					wg.Done()
				}()
			}
			wg.Wait()
			afterIncrease := runtime.NumGoroutine()
			if afterIncrease != beforeIncrease+int(tt.args.delta) {
				t.Errorf("increase doesn't add goroutine, before increase: %d, after increase: %d", beforeIncrease, afterIncrease)
			}
			if len(p.items) != int(tt.args.delta)+1 {
				t.Errorf("increase should append items slice, the length of items is %d, delta is %d", len(p.items), tt.args.delta)
			}
		})
	}
}
