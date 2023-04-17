package nats

import (
	"fmt"
	"log"
	"os"

	"github.com/nats-io/nats.go"
)

var (
	natsUrl        string
	metricsSubject string
	reqSubject     string
)

func init() {
	var ok bool
	natsUrl, ok = os.LookupEnv("NATS_URL")
	if !ok {
		log.Fatal("$NATS_URL not set")
	}
	metricsSubject, ok = os.LookupEnv("METRICS_SUBJECT")
	if !ok {
		log.Fatal("$METRICS_SUBJECT not set")
	}
	reqSubject, ok = os.LookupEnv("REQ_SUBJECT")
	if !ok {
		log.Fatal("$REQ_SUBJECT not set")
	}

}

func Publish(msg []byte) {

	nc, err := nats.Connect(natsUrl)
	if err != nil {
		errMsg := fmt.Sprintf("Cannot connect to nats: %s", err)
		log.Fatal(errMsg)
	}
	defer nc.Close()

	log.Printf("Publishing %d bytes to: %q", len(msg), metricsSubject)

	err = nc.Publish(metricsSubject, msg)
	if err != nil {
		log.Fatal(err.Error())
	}
}

/*
	func Subscribe(ctx context.Context, event []byte) ([]byte, error) {
		nc, err := nats.Connect(natsUrl)
		if err != nil {
			errMsg := fmt.Sprintf("Cannot connect to nats: %s", err)
			log.Fatal(errMsg)
		}
		defer nc.Close()
		ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()

}
*/
type Message struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

/*
func Handle() {
	// 建立 NATS 连接
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// 创建一个取消上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 定义一个计数器和切片，用于记录读取到的消息
	count := 0
	messages := make([]Message, 0)

	// 启动一个 goroutine，持续接收消息
	go func() {
		for {
			// 从队列中接收消息
			msg, err := nc.Subscribe(reqSubject, func(m *nats.Msg) {
				// 解析消息
				var message Message
				fmt.Printf("NATS message type: %T", m.Data)
				err := json.Unmarshal(m.Data, &message)
				if err != nil {
					log.Printf("Failed to unmarshal message: %v", err)
					return
				}

				// 记录消息并增加计数器
				messages = append(messages, message)
				count++

				// 如果计数器达到 batchSize，则打包发送给另一个容器处理
				if count == 4 {
					sendBatch(messages)
					count = 0
					messages = make([]Message, 0)
				}
			})
			if err != nil {
				log.Fatal(err)
			}

			select {
			case <-ctx.Done():
				// 取消订阅并退出 goroutine
				msg.Unsubscribe()
				return
			default:
				// 继续等待接收消息
			}
		}
	}()

	// 持续运行，每隔一段时间检查是否需要发送未处理的消息
	for {
		select {
		case <-ctx.Done():
			// 取消订阅并退出程序
			return
		default:
			// 继续执行
		}

		// 等待一段时间，检查是否需要发送未处理的消息
		time.Sleep(500 * time.Millisecond)

		if len(messages) > 0 {
			sendBatch(messages)
			count = 0
			messages = make([]Message, 0)
		}
	}
}

// Message 是一个简单的消息结构体，用于演示

func sendBatch(messages []Message) {
	// 打包 messages 切片并发送给另一个容器处理
	data, err := json.Marshal(messages)
	if err != nil {
		log.Printf("Failed to marshal messages: %v", err)
		return
	}

	log.Printf("Sending %d messages to another container", len(messages), data)

	// ...
	// 发送 data 给另一个容器处理
	// ...
}
*/
