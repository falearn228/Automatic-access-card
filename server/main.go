package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type State int

const (
	WaitingForCard State = iota
	CardInserted
	DoorOpen
)

func (s State) String() string {
	return [...]string{"WaitingForCard", "CardInserted", "DoorOpen"}[s]
}

type Event int

const (
	EventCardInserted Event = iota
	EventCodeCorrect
	EventCodeIncorrect
	EventCardRemoved
)

type StateMachine struct {
	state State
	mu    sync.Mutex
}

func NewStateMachine() *StateMachine {
	return &StateMachine{state: WaitingForCard}
}

func (sm *StateMachine) GetState() State {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.state
}

func (sm *StateMachine) OnEvent(event Event) (State, string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	oldState := sm.state
	output := "игнор"

	switch sm.state {
	case WaitingForCard:
		switch event {
		case EventCardInserted:
			sm.state = CardInserted
			output = "+"
			// case EventCodeCorrect:
			// 	sm.state = DoorOpen
			// 	output = "-"
		}
	case CardInserted:
		switch event {
		case EventCodeCorrect:
			sm.state = DoorOpen
			output = "+"

		// Измененный обработчик для CardInserted -> CodeCorrect
		// case EventCodeCorrect:
		// 	// ОШИБКА: Должно быть DoorOpen, но установлено WaitingForCard
		// 	sm.state = WaitingForCard
		// 	output = "+" // ОШИБКА: Должен быть "+", но для демонстрации оставим

		case EventCodeIncorrect:
			sm.state = WaitingForCard
			output = "закрыт доступ"
		case EventCardRemoved:
			sm.state = WaitingForCard
			output = "+"
		}
	case DoorOpen:
		if event == EventCardRemoved {
			sm.state = WaitingForCard
			output = "+"
		}
	}

	return oldState, output
}

func main() {
	fsm := NewStateMachine()

	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"state": fsm.GetState().String()})
	})

	// server/main.go
	http.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		fsm = NewStateMachine()
		w.Write([]byte("Состояние сброшено"))
	})

	http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Event string }
		if json.NewDecoder(r.Body).Decode(&req) != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		var event Event
		switch req.Event {
		case "CardInserted":
			event = EventCardInserted
		case "CodeCorrect":
			event = EventCodeCorrect
		case "CodeIncorrect":
			event = EventCodeIncorrect
		case "CardRemoved":
			event = EventCardRemoved
		default:
			http.Error(w, "Unknown event", http.StatusBadRequest)
			return
		}

		oldState, output := fsm.OnEvent(event)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"old_state": oldState.String(),
			"new_state": fsm.GetState().String(),
			"output":    output,
		})
	})

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
