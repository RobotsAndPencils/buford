package push

import "sync"

// Queue up notifications without waiting for the response.
type Queue struct {
	service       *Service
	notifications chan notification
	responses     chan response
	wg            sync.WaitGroup
}

// NewQueue wraps a service with a queue for sending notifications asynchronously.
func NewQueue(service *Service, workers uint) *Queue {
	// unbuffered channels
	q := &Queue{
		service:       service,
		notifications: make(chan notification),
		responses:     make(chan response),
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

// Response blocks waiting for a response. Responses may be received in any order.
func (q *Queue) Response() (id string, deviceToken string, err error) {
	resp := <-q.responses
	q.wg.Done()
	return resp.ApnsID, resp.DeviceToken, resp.Err
}

// Wait for all responses to be handled and shutdown workers to stop accepting notifications.
func (q *Queue) Wait() {
	close(q.notifications)
	q.wg.Wait()
}

// notification to send.
type notification struct {
	DeviceToken string
	Headers     *Headers
	Payload     []byte
}

// response from sending a notification.
type response struct {
	ApnsID      string
	Err         error
	DeviceToken string
}

func worker(q *Queue) {
	for {
		n, more := <-q.notifications
		if !more {
			return
		}
		id, err := q.service.Push(n.DeviceToken, n.Headers, n.Payload)
		q.responses <- response{ApnsID: id, Err: err, DeviceToken: n.DeviceToken}
	}
}
