package watcher

// EventType defines the type of event being broadcast.
type EventType string

const (
	EventPriceUpdated        EventType = "price_updated"
	EventChainDataUpdated    EventType = "chain_data_updated"
	EventGasPriceUpdated     EventType = "gas_price_updated"
	EventTransactionsUpdated EventType = "transactions_updated"
	EventStatusUpdated       EventType = "status_updated"
)

// Event represents a monitoring event.
type Event struct {
	Type EventType
	Data interface{}
}

// Subscriber is a channel that receives events.
type Subscriber chan Event
