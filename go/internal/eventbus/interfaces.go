package eventbus

type EventEmitter interface {
	EmitEvent(eventType SystemEventType, source string, payload interface{})
}
