package redis

import (
	"sync"
	"time"

	"github.com/gagandeepahuja09/chanpubsub"
	"github.com/go-redis/redis"
)

// make sure Transport satisfies vice.Transport interface.
var _ chanpubsub.Transport = (*Transport)(nil)

// newConnection, Send and Receive: all 3 of them ensure single pattern
// by storing the client, sendChans and receiveChans respectively.

// separate publish method goroutine would be running infinitely for each channel.
// it will keep on reading for writes to it

type Transport struct {
	// Operations like Send and Receive should be behind locks
	sync.Mutex
	// we'll need to maintain a mapping of each topic/queue name with the
	// the corresponding channel
	sendChans    map[string]chan []byte
	receiveChans map[string]chan []byte

	wg sync.WaitGroup

	errChan     chan error
	stopChan    chan struct{}
	stopPubChan chan struct{}
	stopSubChan chan struct{}

	client *redis.Client
}

func New(opts ...Option) *Transport {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	return &Transport{
		sendChans:    make(map[string]chan []byte),
		receiveChans: make(map[string]chan []byte),
		errChan:      make(chan error),
		stopChan:     make(chan struct{}),
		stopPubChan:  make(chan struct{}),
		stopSubChan:  make(chan struct{}),
		client:       options.Client,
	}
}

func (t *Transport) newConnection() (*redis.Client, error) {
	var err error
	if t.client != nil {
		return t.client, nil
	}

	t.client = redis.NewClient(&redis.Options{
		Network:    "tcp",
		Addr:       "127.0.0.1:6379",
		Password:   "",
		DB:         0,
		MaxRetries: 0,
	})

	_, err = t.client.Ping().Result()
	return t.client, err
}

// make publisher's job would be to constantly run a goroutine.
// where it will keep on reading from a channel and will push the message in
// the list created by Redis.
// it will return this channel, so that Send method can return it
// and be used by the client to write on this channel.
func (t *Transport) makePublisher(name string) (chan []byte, error) {
	c, err := t.newConnection()
	if err != nil {
		return nil, err
	}

	ch := make(chan []byte, 1024)
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		for {
			select {
			case <-t.stopPubChan:
				if len(ch) != 0 {
					_, err := t.client.Ping().Result()
					// if ch is not empty, it will ensure that all messages get
					// consumed first.
					if err == nil {
						continue
					}
				}
				return
			case msg := <-ch:
				err := c.RPush(name, string(msg)).Err()
				if err != nil {
					t.errChan <- &chanpubsub.TransportErr{Message: msg, Name: name, Err: err}
				}
			}
		}
	}()
	return ch, nil
}

func (t *Transport) Send(name string) chan<- []byte {
	t.Lock()
	defer t.Unlock()

	ch, ok := t.sendChans[name]
	if ok {
		return ch
	}

	ch, err := t.makePublisher(name)
	if err != nil {
		t.errChan <- &chanpubsub.TransportErr{Name: name, Err: err}
		return make(chan<- []byte)
	}
	t.sendChans[name] = ch
	return ch
}

// make subscriber will read the message from redis list and will write to the channel
// how is this ensuring that it keeps on reading from the list infinitely without
// using a for loop?
func (t *Transport) makeSubscriber(name string) (chan []byte, error) {
	c, err := t.newConnection()
	if err != nil {
		return nil, err
	}

	// buffered channel, why?
	ch := make(chan []byte, 1024)
	go func() {
		// blocking list pop
		// pop from tail(r => right) of the list
		data, err := c.BRPop(0*time.Second, name).Result()
		if err != nil {
			select {
			case <-t.stopSubChan:
				return
			default:
				t.errChan <- &chanpubsub.TransportErr{Err: err, Name: name}
			}
		}

		ch <- []byte(data[len(data)-1])
	}()
	return ch, nil
}

func (t *Transport) Receive(name string) <-chan []byte {
	t.Lock()
	defer t.Unlock()

	ch, ok := t.receiveChans[name]
	if ok {
		return ch
	}

	ch, err := t.makeSubscriber(name)
	if err != nil {
		t.errChan <- &chanpubsub.TransportErr{Err: err, Name: name}
		return make(<-chan []byte)
	}
	return ch
}

func (t *Transport) Stop() {
	close(t.stopSubChan)
	close(t.stopPubChan)
	t.wg.Wait()
	t.client.Close()
	close(t.stopChan)
}

func (t *Transport) ErrChan() <-chan error {
	return t.errChan
}

func (t *Transport) Done() chan struct{} {
	return t.stopChan
}
