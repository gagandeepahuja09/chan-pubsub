package redis

import (
	"sync"

	"github.com/gagandeepahuja09/chanpubsub"
	"github.com/go-redis/redis"
)

// make sure Transport satisfies vice.Transport interface.
var _ chanpubsub.Transport = (*Transport)(nil)

type Transport struct {
	// Operations like Send and Receive should be behind locks
	sync.Mutex
	// we'll need to maintain a mapping of each topic/queue name with the
	// the corresponding channel
	sendChans    map[string]chan []byte
	receiveChans map[string]chan []byte

	errChan chan error

	client *redis.Client
}

func New(opts ...Option) *Transport {

}

func (t *Transport) makeSubscriber(name string) (chan []byte, error) {
	c, err := t.newConnection()
	if err != nil {
		return nil, err
	}

	// buffered channel, why?
	ch := make(chan []byte, 1024)
	go func() {
		// defer t.wg.Done()
		// for {
		// 	select {
		// 		case <-t.stopSub
		// 	}
		// }
	}()
}

// make publisher's job would be to constantly run a goroutine.
// where it will keep on reading from a channel.
func (t *Transport) makePublisher() (chan<- []byte, error) {

}

func (t *Transport) Send(name string) chan<- []byte {
	t.Lock()
	defer t.Unlock()

	ch, ok := t.sendChans[name]
	if ok {
		return ch
	}

	ch, err := t.makePublisher(name)
	return ch
}

func (t *Transport) Receive(name string) <-chan []byte {
	t.Lock()
	defer t.Unlock()

	ch, ok := t.receiveChans[name]

	ch, err := t.makeSubscriber(name)
}
