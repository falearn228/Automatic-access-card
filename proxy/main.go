package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
)

type EventPair struct {
	Output string
	Event  string
	State  string // Состояние системы до события
}

type CircularBuffer struct {
	sync.Mutex
	Buffer   []EventPair
	Capacity int
	Index    int
}

var buffer = &CircularBuffer{
	Buffer:   make([]EventPair, 0, 4),
	Capacity: 4,
}

type Sequinces struct {
	State        string
	FirstEvent   string
	SecondEvent  string
	FirstOutput  string
	SecondOutput string
	NextState    string
}

type Transition struct {
	CurrentState string
	Event        string
	NextState    string
	Output       string
}

var (
	totalRequests    uint64
	detectedErrors   uint64
	coveredTrans     = make(map[string]struct{})
	coveredTransLock sync.Mutex
)

var (
	uniqueSequences     = make(map[string]struct{}) // Хранит уникальные последовательности
	uniqueSequencesLock sync.Mutex                  // Для потокобезопасности
)

var installationSeq = []Sequinces{
	{"WaitingForCard", "CodeCorrect", "CardInserted", "игнор", "+", "CardInserted"},
	{"WaitingForCard", "CodeCorrect", "CardRemoved", "игнор", "игнор", "WaitingForCard"},
	{"WaitingForCard", "CardInserted", "CodeCorrect", "+", "+", "DoorOpen"},
	{"WaitingForCard", "CardInserted", "CodeIncorrect", "+", "закрыт доступ", "CardInserted"},
	{"WaitingForCard", "CardInserted", "CardRemoved", "+", "+", "CardInserted"},
	{"WaitingForCard", "CodeIncorrect", "CardInserted", "игнор", "+", "CardInserted"},
	{"WaitingForCard", "CodeIncorrect", "CardRemoved", "игнор", "игнор", "WaitingForCard"},
	{"WaitingForCard", "CardRemoved", "CodeCorrect", "игнор", "игнор", "WaitingForCard"},
	{"WaitingForCard", "CardRemoved", "CodeIncorrect", "игнор", "игнор", "WaitingForCard"},
	{"WaitingForCard", "CardRemoved", "CardInserted", "игнор", "+", "CardInserted"},
	{"WaitingForCard", "CardRemoved", "CardRemoved", "игнор", "игнор", "WaitingForCard"},
	{"CardInserted", "CodeCorrect", "CardInserted", "+", "игнор", "DoorOpen"},
	{"CardInserted", "CodeCorrect", "CardRemoved", "+", "+", "WaitingForCard"},
	{"CardInserted", "CardInserted", "CodeCorrect", "игнор", "+", "DoorOpen"},
	{"CardInserted", "CardInserted", "CodeIncorrect", "игнор", "закрыт доступ", "WaitingForCard"},
	{"CardInserted", "CardInserted", "CardRemoved", "игнор", "+", "WaitingForCard"},
	{"CardInserted", "CodeIncorrect", "CardInserted", "закрыт доступ", "+", "CardInserted"},
	{"CardInserted", "CodeIncorrect", "CardRemoved", "закрыт доступ", "игнор", "WaitingForCard"},
	{"CardInserted", "CardRemoved", "CodeCorrect", "+", "+", "WaitingForCard"},
	{"CardInserted", "CardRemoved", "CodeIncorrect", "+", "игнор", "WaitingForCard"},
	{"CardInserted", "CardRemoved", "CardInserted", "+", "+", "CardInserted"},
	{"CardInserted", "CardRemoved", "CardRemoved", "+", "игнор", "WaitingForCard"},

	{"DoorOpen", "CodeCorrect", "CardInserted", "игнор", "игнор", "DoorOpen"},
	{"DoorOpen", "CodeCorrect", "CardRemoved", "игнор", "+", "WaitingForCard"},
	{"DoorOpen", "CardInserted", "CodeCorrect", "игнор", "игнор", "DoorOpen"},
	{"DoorOpen", "CardInserted", "CodeIncorrect", "игнор", "игнор", "DoorOpen"},
	{"DoorOpen", "CardInserted", "CardRemoved", "игнор", "+", "WaitingForCard"},
	{"DoorOpen", "CodeIncorrect", "CardInserted", "игнор", "игнор", "DoorOpen"},
	{"DoorOpen", "CodeIncorrect", "CardRemoved", "игнор", "+", "WaitingForCard"},
	{"DoorOpen", "CardRemoved", "CodeCorrect", "+", "игнор", "WaitingForCard"},
	{"DoorOpen", "CardRemoved", "CodeIncorrect", "+", "игнор", "WaitingForCard"},
	{"DoorOpen", "CardRemoved", "CardInserted", "+", "+", "CardInserted"},
	{"DoorOpen", "CardRemoved", "CardRemoved", "+", "игнор", "WaitingForCard"},
}

var transitions = []Transition{
	{"WaitingForCard", "CardRemoved", "WaitingForCard", "игнор"},
	{"WaitingForCard", "CodeCorrect", "WaitingForCard", "игнор"},
	{"WaitingForCard", "CodeIncorrect", "WaitingForCard", "игнор"},
	{"WaitingForCard", "CardInserted", "CardInserted", "+"},

	{"CardInserted", "CardInserted", "CardInserted", "игнор"},
	{"CardInserted", "CodeCorrect", "DoorOpen", "+"},
	{"CardInserted", "CodeIncorrect", "WaitingForCard", "закрыт доступ"},
	{"CardInserted", "CardRemoved", "WaitingForCard", "+"},

	{"DoorOpen", "CodeCorrect", "DoorOpen", "игнор"},
	{"DoorOpen", "CodeIncorrect", "DoorOpen", "игнор"},
	{"DoorOpen", "CardInserted", "DoorOpen", "игнор"},
	{"DoorOpen", "CardRemoved", "WaitingForCard", "+"},
}

func (cb *CircularBuffer) Add(output, event, state string) {
	cb.Lock()
	defer cb.Unlock()

	pair := EventPair{
		Output: output,
		Event:  event,
		State:  state,
	}

	if len(cb.Buffer) < cb.Capacity {
		cb.Buffer = append(cb.Buffer, pair)
	} else {
		cb.Buffer[cb.Index] = pair
		cb.Index = (cb.Index + 1) % cb.Capacity
	}
}

func getCurrentState() string {
	if len(buffer.Buffer) == 0 {
		return "WaitingForCard" // Начальное состояние
	}
	return buffer.Buffer[len(buffer.Buffer)-1].State
}

func (cb *CircularBuffer) GetLastPairs(n int) []EventPair {
	cb.Lock()
	defer cb.Unlock()

	if len(cb.Buffer) < n {
		return cb.Buffer
	}
	return append(cb.Buffer[cb.Index:], cb.Buffer[:cb.Index]...)[:n]
}

func main() {
	client := &http.Client{}
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&totalRequests, 1)

		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(body))

		// Получаем текущее состояние
		stateResp, err := http.Get("http://localhost:8080/state")
		if err != nil {
			log.Printf("🚨 Ошибка получения состояния: %v", err)
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}

		var state struct{ State string }
		json.NewDecoder(stateResp.Body).Decode(&state)
		stateResp.Body.Close()

		// Парсим событие
		var req struct{ Event string }
		if err := json.Unmarshal(body, &req); err != nil {
			log.Printf("⚠️ Ошибка парсинга: %v", err)
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		checkInstallationSequences()
		// Поиск разрешенных переходов
		var allowed []Transition
		for _, t := range transitions {
			if t.CurrentState == state.State && t.Event == req.Event {
				allowed = append(allowed, t)
				buffer.Add(t.Output, t.Event, t.CurrentState)
				checkInstallationSequences()
			}
		}

		if len(allowed) == 0 {
			log.Printf("🚫 Блокировка: %s → %s", state.State, req.Event)
			buffer.Add("ERROR", req.Event, state.State)
			http.Error(w, "Transition not allowed", http.StatusForbidden)

			// atomic.AddUint64(&detectedErrors, 1)
			// buffer.Add("ERROR", req.Event)
			// http.Error(w, "Transition not allowed", http.StatusForbidden)
			return
		}

		// Проксируем запрос
		resp, err := client.Post("http://localhost:8080/event", "application/json", bytes.NewReader(body))
		if err != nil {
			log.Printf("🔥 Ошибка: %v", err)
			http.Error(w, "Gateway error", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Анализ результата
		var result struct {
			OldState string `json:"old_state"`
			NewState string `json:"new_state"`
			Output   string `json:"output"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Printf("🚨 Ошибка декодирования ответа: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Проверка переходов
		transitionKey := fmt.Sprintf("%s→%s→%s", result.OldState, req.Event, result.NewState)
		coveredTransLock.Lock()
		coveredTrans[transitionKey] = struct{}{}
		coveredTransLock.Unlock()

		// Поиск несоответствий
		for _, t := range allowed {
			if t.NextState == result.NewState {
				return
			}
		}

		atomic.AddUint64(&detectedErrors, 1)
		log.Printf("❗️ Несоответствие! Ожидалось: %v, Получено: %s", allowed, result.NewState)
		http.Error(w, "Command blocked", http.StatusForbidden)
	})

	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		coveredTransLock.Lock()
		defer coveredTransLock.Unlock()

		stats := map[string]interface{}{
			"total_requests":    atomic.LoadUint64(&totalRequests),
			"detected_errors":   atomic.LoadUint64(&detectedErrors),
			"covered_trans":     len(coveredTrans),
			"total_trans":       len(transitions),
			"coverage_percent":  fmt.Sprintf("%.1f%%", 100*float64(len(coveredTrans))/float64(len(transitions))),
			"unique_sequences":  len(uniqueSequences),  // Количество уникальных последовательностей
			"undetected_states": getUndetectedStates(), // Необнаруженные состояния
		}

		json.NewEncoder(w).Encode(stats)
	})

	log.Println("🚀 Прокси запущен на :8081")
	http.ListenAndServe(":8081", nil)
}

func checkInstallationSequences() {
	pairs := buffer.GetLastPairs(2) // Получаем последние 2 события
	if len(pairs) < 2 {
		return
	}

	for _, seq := range installationSeq {
		if pairs[0].Event == seq.FirstEvent &&
			pairs[0].Output == seq.FirstOutput &&
			pairs[1].Event == seq.SecondEvent &&
			pairs[1].Output == seq.SecondOutput &&
			pairs[0].State == seq.State {

			// Формируем ключ для уникальной последовательности
			sequenceKey := fmt.Sprintf(
				"%s:%s:%s:%s:%s",
				seq.FirstEvent, seq.FirstOutput,
				seq.SecondEvent, seq.SecondOutput,
				seq.NextState,
			)

			// Проверяем, была ли уже такая последовательность
			uniqueSequencesLock.Lock()
			if _, exists := uniqueSequences[sequenceKey]; !exists {
				// Если последовательность уникальна, добавляем её в множество
				uniqueSequences[sequenceKey] = struct{}{}
			}
			uniqueSequencesLock.Unlock()

			log.Printf(
				"Обнаружена последовательность: %s -> %s (Исходное состояние: %s, Новое состояние: %s)",
				seq.FirstEvent, seq.SecondEvent, pairs[0].State, seq.NextState,
			)
		}
	}
}

// func getUndetectedStates() []string {

// 	var undetected []string
// 	for state := range installationSeq {
// 		if _, found := installationSeq[state]; !found {
// 			undetected = append(undetected, state)
// 		}
// 	}
// 	return undetected
// }
