package graph

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"graphql-go/graph/model"
	"sync"
)

//go:generate go run github.com/99designs/gqlgen generate

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DB               *gorm.DB
	subscribers      map[string]chan *model.BurgerBellEvent
	subscribersMutex sync.Mutex
	BurgerBellChan   chan *model.BurgerBellEvent // Keep for backward compatibility
}

func NewResolver(db *gorm.DB) *Resolver {
	r := &Resolver{
		DB:             db,
		subscribers:    make(map[string]chan *model.BurgerBellEvent),
		BurgerBellChan: make(chan *model.BurgerBellEvent, 1),
	}

	// Start a goroutine to handle publishing events to all subscribers
	go r.broadcastEvents()

	return r
}

// RegisterSubscriber adds a new subscriber channel and returns its ID
func (r *Resolver) RegisterSubscriber() (string, chan *model.BurgerBellEvent) {
	r.subscribersMutex.Lock()
	defer r.subscribersMutex.Unlock()

	// Generate a unique ID for this subscriber
	id := "sub-" + uuid.New().String()

	// Create a buffered channel for this subscriber
	ch := make(chan *model.BurgerBellEvent, 5) // Buffer size of 5 should be enough

	// Store the channel in the map
	r.subscribers[id] = ch

	return id, ch
}

// UnregisterSubscriber removes a subscriber
func (r *Resolver) UnregisterSubscriber(id string) {
	r.subscribersMutex.Lock()
	defer r.subscribersMutex.Unlock()

	if ch, ok := r.subscribers[id]; ok {
		close(ch)
		delete(r.subscribers, id)
	}
}

// PublishBurgerBellEvent sends an event to all subscribers
func (r *Resolver) PublishBurgerBellEvent(event *model.BurgerBellEvent) {
	// Send to main channel for backward compatibility
	select {
	case r.BurgerBellChan <- event:
		// Event sent successfully
	default:
		// Channel is full, could log this or handle it differently
	}
	
	// Directly send to all subscribers
	r.subscribersMutex.Lock()
	defer r.subscribersMutex.Unlock()
	
	for _, ch := range r.subscribers {
		// Non-blocking send to each subscriber
		select {
		case ch <- event:
			// Event sent successfully
		default:
			// Channel is full, could log this or handle it differently
		}
	}
}

// broadcastEvents listens for events on the main channel and broadcasts them to all subscribers
func (r *Resolver) broadcastEvents() {
	for event := range r.BurgerBellChan {
		r.subscribersMutex.Lock()
		for _, ch := range r.subscribers {
			// Non-blocking send to each subscriber
			select {
			case ch <- event:
				// Event sent successfully
			default:
				// Channel is full, could log this or handle it differently
			}
		}
		r.subscribersMutex.Unlock()
	}
}
