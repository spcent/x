package eventbus

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Event interface {
	Key() string
	Payload() []byte
	ID() string
	At() time.Time
}

type BaseEvent struct {
	key     string
	payload []byte
	id      string
	at      time.Time
}

func (e *BaseEvent) Key() string     { return e.key }
func (e *BaseEvent) Payload() []byte { return e.payload }
func (e *BaseEvent) ID() string      { return e.id }
func (e *BaseEvent) At() time.Time   { return e.at }

// 这里只使用json格式进行序列化
func NewJSONEvent(key string, data any) Event {
	b, _ := json.Marshal(data)
	return &BaseEvent{
		key:     key,
		payload: b,
		id:      uuid.NewString(),
		at:      time.Now(),
	}
}

// DecodeJSON: 将事件负载解为目标类型
func DecodeJSON[T any](evt Event) (T, error) {
	var zero T
	var v T
	if err := json.Unmarshal(evt.Payload(), &v); err != nil {
		return zero, err
	}
	return v, nil
}
