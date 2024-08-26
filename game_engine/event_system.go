package game_engine

type EventSystem struct {
	listeners map[string][]func(map[string]interface{})
}

func NewEventSystem() *EventSystem {
	return &EventSystem{
		listeners: make(map[string][]func(map[string]interface{})),
	}
}

func (es *EventSystem) On(eventName string, listener func(map[string]interface{})) {
	es.listeners[eventName] = append(es.listeners[eventName], listener)
}

func (es *EventSystem) Emit(eventName string, data map[string]interface{}) {
	if listeners, ok := es.listeners[eventName]; ok {
		for _, listener := range listeners {
			listener(data)
		}
	}
}
