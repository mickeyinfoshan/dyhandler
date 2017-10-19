package poolhandler

import (
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/nsqio/go-nsq"
)

type testHandler struct {
}

func (t testHandler) HandleMessage(msg *nsq.Message) error {
	return nil
}

func Test_newPoolItem(t *testing.T) {
	msgChan := make(chan *nsq.Message)
	defer close(msgChan)
	handler := testHandler{}
	type args struct {
		msgChan *chan *nsq.Message
		handler nsq.Handler
	}
	tests := []struct {
		name string
		args args
		want *poolItem
	}{
		// TODO: Add test cases.
		{
			name: "create new item",
			args: args{
				handler: handler,
				msgChan: &msgChan,
			},
			want: &poolItem{
				handler: handler,
				msgChan: &msgChan,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newPoolItem(tt.args.msgChan, tt.args.handler); !reflect.DeepEqual(got.msgChan, tt.want.msgChan) && !reflect.DeepEqual(got.handler, tt.want.handler) {
				t.Errorf("newPoolItem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_poolItem_destroy(t *testing.T) {
	msgChan := make(chan *nsq.Message)
	handler := testHandler{}
	pooItem := newPoolItem(&msgChan, handler)
	beforeLoop := runtime.NumGoroutine()
	go pooItem.loop()
	tests := []struct {
		name string
		p    *poolItem
	}{
		// TODO: Add test cases.
		{
			name: "destroy item and relase resources",
			p:    pooItem,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.destroy()
			time.Sleep(time.Millisecond * 200)
			afterDestroy := runtime.NumGoroutine()
			// t.Run will add a goroutine.
			if afterDestroy != beforeLoop+1 {
				t.Errorf("destroy didn't release the goroutine, before loop: %d, after destroy: %d", beforeLoop, afterDestroy)
			}
		})
	}
}

func Test_poolItem_loop(t *testing.T) {
	msgChan := make(chan *nsq.Message)
	handler := testHandler{}
	pooItem := newPoolItem(&msgChan, handler)
	beforeLoop := runtime.NumGoroutine()
	tests := []struct {
		name string
		p    *poolItem
	}{
		// TODO: Add test cases.
		{
			name: "loop for receive messages",
			p:    pooItem,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timer := time.NewTimer(time.Second * 1)
			go func() {
				<-timer.C
				timer.Stop()
				afterLoop := runtime.NumGoroutine()
				// t.Run and this goroutine
				if afterLoop != beforeLoop+2 {
					t.Errorf("loop doesn't create a goroutine, before loop: %d, after loop: %d", beforeLoop, afterLoop)
				}
				tt.p.destroy()
			}()
			tt.p.loop()
		})
	}
}
