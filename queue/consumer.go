package queue

type consumerStatus uint16

const (
	consWait consumerStatus = 0
	consBusy consumerStatus = 1
)

type Consumer[T any] struct {
	ID     string
	Sub    chan *Msg[T]
	offset uint64
	status consumerStatus
}
