package publisher

import (
	"context"
	"log"
	"sync"
	"time"

	zmq "github.com/go-zeromq/zmq4"
	"github.com/vmihailenco/msgpack/v5"
)

type Message struct {
	Content string
}

type publisher struct {
	socket       zmq.Socket
	message      Message
	closeChannel chan bool
	ListenLocal  func()
	Listen       func(string)
	Update       func(string)

	routine func()
	Start   func()
	Play    func()
	Pause   func()
	Wait    func()
	Quit    func()
}

var Publisher *publisher
var once sync.Once

var (
	chWork       <-chan struct{}
	chWorkBackup <-chan struct{}
	chControl    chan struct{}
	wg           sync.WaitGroup
)

func GetInstance() *publisher {
	once.Do(func() {
		Publisher = &publisher{
			socket:       zmq.NewPub(context.Background()),
			closeChannel: make(chan bool),
			ListenLocal:  listenLocal5555,
			Listen:       listen,
			Update:       update,

			routine: routine,
			Start:   start,
			Play:    play,
			Pause:   pause,
			Wait:    wait,
			Quit:    quit,
		}
	})

	return Publisher
}

func Send() {
	byteArray, err := msgpack.Marshal(&Publisher.message)
	if err != nil {
		panic(err)
	}

	msg := zmq.NewMsgFrom([]byte(byteArray))
	err = Publisher.socket.Send(msg)
	if err != nil {
		log.Fatal(err)
	}
}

func Close() {
	close(Publisher.closeChannel)
	Publisher.socket.Close()
}

func listenLocal5555() {
	listen("tcp://*:5555")
}

func listen(endpoint string) {
	if "" == endpoint {
		endpoint = "tcp://*:5555"
	}

	err := Publisher.socket.Listen(endpoint)
	if err != nil {
		log.Fatalf("could not listen: %v", err)
	}
}

func update(data string) {
	Publisher.message = Message{Content: data}
}

func routine() {
	defer wg.Done()

	for {
		select {
		case <-chWork:
			Send()
			// time.Sleep(250 * time.Millisecond)
			time.Sleep(1 * time.Second)
		case _, ok := <-chControl:
			if ok {
				continue
			}
			return
		}
	}
}

func start() {
	ch := make(chan struct{})
	close(ch)
	chWork = ch
	chWorkBackup = ch

	chControl = make(chan struct{})

	wg = sync.WaitGroup{}
	wg.Add(1)
	go Publisher.routine()
}

func play() {
	chWork = chWorkBackup
	chControl <- struct{}{}
}

func pause() {
	chWork = nil
	chControl <- struct{}{}
}

func wait() {
	wg.Wait()
}

func quit() {
	chWork = nil
	close(chControl)
}
