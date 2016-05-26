package gmvc

/*
Function used to bind to specified event
when the event triggered it will be executed
*/
type EventHandler func(args ...interface{})

/*
Event manages events and handlers
an event is just a collection of functions
*/
type Event struct {
    events map[string][]EventHandler
}

/*
NewEvent creates and initialize an Event struct
*/
func NewEvent() *Event {
    return &Event{make(map[string][]EventHandler)}
}

/*
AddHandler adds an EventHandler to an event by name
if there is no event with the name, it will be created automatically
*/
func (e *Event) AddHandler(name string, handler EventHandler) {
    handlers, has := e.events[name]
    if !has {
        handlers = make([]EventHandler, 0, 1)
    }
    e.events[name] = append(handlers, handler)
}

/*
ClearHandlers removes all handlers binds to event by name
*/
func (e *Event) ClearHandlers(name string) {
    delete(e.events, name)
}

/*
ClearHandlers removes all handlers in e
*/
func (e *Event) ClearAllHandlers() {
    e.events = make(map[string][]EventHandler)
}

/*
Trigger triggers event by name, handlers are executed in FIFO order
*/
func (e *Event) Trigger(name string, args ...interface{}) {
    if handlers, has := e.events[name]; has && len(handlers) > 0 {
        for _, handler := range handlers {
            handler(args...)
        }
    }
}


