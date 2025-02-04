// package main

// import (
// 	"testing"
// )

// // Вспомогательная функция для выполнения последовательности событий
// // Теперь возвращает срез выходных реакций
// func executeSequence(fsm *StateMachine, events []Event) (State, []string) {
// 	outputs := make([]string, len(events))
// 	for i, event := range events {
// 		outputs[i] = fsm.OnEvent(event)
// 	}
// 	return fsm.state, outputs
// }

// // Первая группа тестов
// func TestFirstGroup(t *testing.T) {
// 	sequences := []struct {
// 		name            string
// 		events          []Event
// 		expectedState   State
// 		expectedOutputs []string
// 	}{
// 		{
// 			"код+",
// 			[]Event{EventCodeCorrect},
// 			WaitingForCard,
// 			[]string{"игнор"},
// 		},
// 		{
// 			"код-",
// 			[]Event{EventCodeIncorrect},
// 			WaitingForCard,
// 			[]string{"игнор"},
// 		},
// 		{
// 			"карта_убрана",
// 			[]Event{EventCardRemoved},
// 			WaitingForCard,
// 			[]string{"игнор"},
// 		},
// 		{
// 			"карта_вставлена",
// 			[]Event{EventCardInserted},
// 			CardInserted,
// 			[]string{"+"},
// 		},
// 		{
// 			"карта_вставлена, код-",
// 			[]Event{EventCardInserted, EventCodeIncorrect},
// 			WaitingForCard,
// 			[]string{"+", "закрыт доступ"},
// 		},
// 		{
// 			"карта_вставлена, карта_убрана",
// 			[]Event{EventCardInserted, EventCardRemoved},
// 			WaitingForCard,
// 			[]string{"+", "+"},
// 		},
// 		{
// 			"карта_вставлена, карта_вставлена",
// 			[]Event{EventCardInserted, EventCardInserted},
// 			CardInserted,
// 			[]string{"+", "игнор"},
// 		},
// 		{
// 			"карта_вставлена, код+",
// 			[]Event{EventCardInserted, EventCodeCorrect},
// 			DoorOpen,
// 			[]string{"+", "+"},
// 		},
// 		{
// 			"карта_вставлена, код+, код+",
// 			[]Event{EventCardInserted, EventCodeCorrect, EventCodeCorrect},
// 			DoorOpen,
// 			[]string{"+", "+", "игнор"},
// 		},
// 		{
// 			"карта_вставлена, код+, код-",
// 			[]Event{EventCardInserted, EventCodeCorrect, EventCodeIncorrect},
// 			DoorOpen,
// 			[]string{"+", "+", "игнор"},
// 		},
// 		{
// 			"карта_вставлена, код+, карта_вставлена",
// 			[]Event{EventCardInserted, EventCodeCorrect, EventCardInserted},
// 			DoorOpen,
// 			[]string{"+", "+", "игнор"},
// 		},
// 		{
// 			"карта_вставлена, код+, карта_убрана",
// 			[]Event{EventCardInserted, EventCodeCorrect, EventCardRemoved},
// 			WaitingForCard,
// 			[]string{"+", "+", "+"},
// 		},
// 	}

// 	for _, seq := range sequences {
// 		t.Run(seq.name, func(t *testing.T) {
// 			fsm := NewStateMachine()
// 			finalState, outputs := executeSequence(fsm, seq.events)

// 			// Проверка конечного состояния
// 			if finalState != seq.expectedState {
// 				t.Errorf("Неправильное конечное состояние для последовательности %s: получили %v, ожидали %v",
// 					seq.name, finalState, seq.expectedState)
// 			}

// 			// Проверка выходных реакций
// 			if len(outputs) != len(seq.expectedOutputs) {
// 				t.Errorf("Неправильное количество выходных реакций для последовательности %s: получили %v, ожидали %v",
// 					seq.name, len(outputs), len(seq.expectedOutputs))
// 				return
// 			}

// 			for i, output := range outputs {
// 				if output != seq.expectedOutputs[i] {
// 					t.Errorf("Неправильная выходная реакция для последовательности %s на шаге %d: получили %v, ожидали %v",
// 						seq.name, i, output, seq.expectedOutputs[i])
// 				}
// 			}
// 		})
// 	}
// }

// // Вторая группа тестов
// func TestSecondGroup(t *testing.T) {
// 	sequences := []struct {
// 		name            string
// 		events          []Event
// 		expectedState   State
// 		expectedOutputs []string
// 	}{
// 		{
// 			"код+, карта_убрана",
// 			[]Event{EventCodeCorrect, EventCardRemoved},
// 			WaitingForCard,
// 			[]string{"игнор", "игнор"},
// 		},
// 		{
// 			"код+, код+, карта_убрана",
// 			[]Event{EventCodeCorrect, EventCodeCorrect, EventCardRemoved},
// 			WaitingForCard,
// 			[]string{"игнор", "игнор", "игнор"},
// 		},
// 		{
// 			"код-, код+, карта_убрана",
// 			[]Event{EventCodeIncorrect, EventCodeCorrect, EventCardRemoved},
// 			WaitingForCard,
// 			[]string{"игнор", "игнор", "игнор"},
// 		},
// 		{
// 			"карта_убрана, код+, карта_убрана",
// 			[]Event{EventCardRemoved, EventCodeCorrect, EventCardRemoved},
// 			WaitingForCard,
// 			[]string{"игнор", "игнор", "игнор"},
// 		},
// 		{
// 			"карта_вставлена, код+, карта_убрана",
// 			[]Event{EventCardInserted, EventCodeCorrect, EventCardRemoved},
// 			WaitingForCard,
// 			[]string{"+", "+", "+"},
// 		},
// 		{
// 			"карта_вставлена, код-, код+, карта_убрана",
// 			[]Event{EventCardInserted, EventCodeIncorrect, EventCodeCorrect, EventCardRemoved},
// 			WaitingForCard,
// 			[]string{"+", "закрыт доступ", "игнор", "игнор"},
// 		},
// 		{
// 			"карта_вставлена, карта_убрана, код+, карта_убрана",
// 			[]Event{EventCardInserted, EventCardRemoved, EventCodeCorrect, EventCardRemoved},
// 			WaitingForCard,
// 			[]string{"+", "+", "игнор", "игнор"},
// 		},
// 		{
// 			"карта_вставлена, карта_вставлена, код+, карта_убрана",
// 			[]Event{EventCardInserted, EventCardInserted, EventCodeCorrect, EventCardRemoved},
// 			WaitingForCard,
// 			[]string{"+", "игнор", "+", "+"},
// 		},
// 		{
// 			"карта_вставлена, код+, код+, карта_убрана",
// 			[]Event{EventCardInserted, EventCodeCorrect, EventCodeCorrect, EventCardRemoved},
// 			WaitingForCard,
// 			[]string{"+", "+", "игнор", "+"},
// 		},
// 		{
// 			"карта_вставлена, код+, код+, код+, карта_убрана",
// 			[]Event{EventCardInserted, EventCodeCorrect, EventCodeCorrect, EventCodeCorrect, EventCardRemoved},
// 			WaitingForCard,
// 			[]string{"+", "+", "игнор", "игнор", "+"},
// 		},
// 		{
// 			"карта_вставлена, код+, код-, код+, карта_убрана",
// 			[]Event{EventCardInserted, EventCodeCorrect, EventCodeIncorrect, EventCodeCorrect, EventCardRemoved},
// 			WaitingForCard,
// 			[]string{"+", "+", "игнор", "игнор", "+"},
// 		},
// 		{
// 			"карта_вставлена, код+, карта_вставлена, код+, карта_убрана",
// 			[]Event{EventCardInserted, EventCodeCorrect, EventCardInserted, EventCodeCorrect, EventCardRemoved},
// 			WaitingForCard,
// 			[]string{"+", "+", "игнор", "игнор", "+"},
// 		},
// 		{
// 			"карта_вставлена, код+, карта_убрана, код+, карта_убрана",
// 			[]Event{EventCardInserted, EventCodeCorrect, EventCardRemoved, EventCodeCorrect, EventCardRemoved},
// 			WaitingForCard,
// 			[]string{"+", "+", "+", "игнор", "игнор"},
// 		},
// 	}

// 	for _, seq := range sequences {
// 		t.Run(seq.name, func(t *testing.T) {
// 			fsm := NewStateMachine()
// 			finalState, outputs := executeSequence(fsm, seq.events)

// 			// Проверка конечного состояния
// 			if finalState != seq.expectedState {
// 				t.Errorf("Неправильное конечное состояние для последовательности %s: получили %v, ожидали %v",
// 					seq.name, finalState, seq.expectedState)
// 			}

// 			// Проверка выходных реакций
// 			if len(outputs) != len(seq.expectedOutputs) {
// 				t.Errorf("Неправильное количество выходных реакций для последовательности %s: получили %v, ожидали %v",
// 					seq.name, len(outputs), len(seq.expectedOutputs))
// 				return
// 			}

// 			for i, output := range outputs {
// 				if output != seq.expectedOutputs[i] {
// 					t.Errorf("Неправильная выходная реакция для последовательности %s на шаге %d: получили %v, ожидали %v",
// 						seq.name, i, output, seq.expectedOutputs[i])
// 				}
// 			}
// 		})
// 	}
// }
