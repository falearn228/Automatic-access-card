package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestStateTransitions(t *testing.T) {
	testCases := []struct {
		sequence      []string
		expectedState string
	}{
		{
			[]string{"CardInserted", "CodeCorrect", "CardRemoved"},
			"WaitingForCard",
		},
		{
			[]string{"CardInserted", "CodeIncorrect"},
			"WaitingForCard",
		},
		{
			[]string{"CardInserted", "CodeCorrect"},
			"DoorOpen",
		},
	}

	for _, tc := range testCases {
		resetServer() // Отправка специального запроса на сброс состояния

		var finalState string
		for _, event := range tc.sequence {
			resp, _ := http.Post("http://localhost:8081/event")
			var result struct{ NewState string }
			json.NewDecoder(resp.Body).Decode(&result)
			finalState = result.NewState
		}

		if finalState != tc.expectedState {
			t.Errorf("Для последовательности %v ожидалось %s, получено %s",
				tc.sequence, tc.expectedState, finalState)
		}
	}
}
