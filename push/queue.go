package push

import "sync"

// Queue up notifications without waiting for the response.
type Queue struct {
	service       *Service
	notifications chan notification
	Responses     chan Response
	wg            sync.WaitGroup
}

// notification to send.
type notification struct {
	DeviceToken string
	Headers     *Headers
	Payload     []byte
}

// Response from sending a notification.
type Response struct {
	DeviceToken string
	ID          string
	Err         error
}

// NewQueue wraps a service with a queue for sending notifications asynchronously.
func NewQueue(service *Service, workers uint) *Queue {
	// unbuffered channels
	q := &Queue{
		service:       service,
		notifications: make(chan notification),
		Responses:     make(chan Response),
	}
	// startup workers to send notifications
	for i := uint(0); i < workers; i++ {
		go worker(q)
	}
	return q
}

// Push queues a notification to the APN service.
func (q *Queue) Push(deviceToken string, headers *Headers, payload []byte) {
	n := notification{
		DeviceToken: deviceToken,
		Headers:     headers,
		Payload:     payload,
	}
	q.wg.Add(1)
	q.notifications <- n
}

// Wait for all responses to be handled and then close channels.
func (q *Queue) Wait() {
	// Stop accepting new notifications and shutdown workers after existing notifications
	// are processed:
	close(q.notifications)
	// Wait for all responses to be handled:
	q.wg.Wait()
	// Close responses channel to clean up:
	close(q.Responses)
}

func worker(q *Queue) {
	for n := range q.notifications {
		id, err := q.service.Push(n.DeviceToken, n.Headers, n.Payload)
		q.Responses <- Response{DeviceToken: n.DeviceToken, ID: id, Err: err}
		q.wg.Done()
	}
}
