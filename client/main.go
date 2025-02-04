package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type proxyStats struct {
	Unique     int      `json:"unique_sequences"`
	Total      int      `json:"total_trans"`
	Undedected []string `json:"undetected_states"`
}

var (
	networkErrors int
	blockedErrors int
	serverErrors  int
)

type Stats struct {
	TotalRequests int
	ErrorCount    int
	Coverage      float64
	CoveredTrans  int
	TotalTrans    int
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Выберите режим работы:")
	fmt.Println("1 - Ручной ввод")
	fmt.Println("2 - Автоматическая генерация")
	fmt.Print("> ")

	mode, _ := reader.ReadString('\n')
	switch strings.TrimSpace(mode) {
	case "2":
		autoGenerateTraffic()
	default:
		manualMode(*reader)
	}
}

func manualMode(reader bufio.Reader) {
	for {
		fmt.Println("\n=== Smart Lock Client ===")
		printState()
		fmt.Println("\nДоступные команды:")
		fmt.Println("1. CardInserted")
		fmt.Println("2. CodeCorrect")
		fmt.Println("3. CodeIncorrect")
		fmt.Println("4. CardRemoved")
		fmt.Println("5. Exit")
		fmt.Print("Выберите команду: ")

		input, _ := reader.ReadString('\n')
		cmd := strings.TrimSpace(input)

		switch cmd {
		case "1", "CardInserted":
			sendEvent("CardInserted")
		case "2", "CodeCorrect":
			sendEvent("CodeCorrect")
		case "3", "CodeIncorrect":
			sendEvent("CodeIncorrect")
		case "4", "CardRemoved":
			sendEvent("CardRemoved")
		case "5", "Exit":
			fmt.Println("Выход...")
			return
		default:
			fmt.Println("Неизвестная команда!")
		}
	}
}

func printState() {
	resp, err := http.Get("http://localhost:8080/state")
	if err != nil {
		fmt.Println("Ошибка соединения с сервером:", err)
		return
	}
	defer resp.Body.Close()

	var state struct{ State string }
	json.NewDecoder(resp.Body).Decode(&state)
	fmt.Printf("Текущее состояние: [%s]\n", state.State)
}

func sendEvent(event string) {
	reqBody, _ := json.Marshal(map[string]string{"event": event})
	resp, err := http.Post("http://localhost:8081/event", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	defer resp.Body.Close()

	var result struct {
		OldState, NewState, Output string
	}
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Printf("\nРезультат: %s\nПереход: %s → %s\n",
		result.Output, result.OldState, result.NewState)
}

func autoGenerateTraffic() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	events := []string{"CardInserted", "CodeCorrect", "CodeIncorrect", "CardRemoved"}

	var stats Stats
	start := time.Now()

	for i := 0; i < 435; i++ {
		event := events[rand.Intn(len(events))]
		fmt.Printf("Запрос %d: %s\n", i+1, event)

		resp, err := http.Post("http://localhost:8081/event",
			"application/json",
			bytes.NewReader([]byte(`{"event":"`+event+`"}`)))

		stats.TotalRequests++

		if err != nil {
			fmt.Printf("Сетевая ошибка: %v\n", err)
			networkErrors++
		} else if resp.StatusCode == http.StatusForbidden {
			blockedErrors++
			fmt.Printf("Запрещённая команда: %s\n", event)
		} else if resp.StatusCode != http.StatusOK {
			serverErrors++
			fmt.Printf("Ошибка сервера: %d\n", resp.StatusCode)
		}

		// Получаем статистику покрытия от прокси

		resp.Body.Close()
	}

	resp, _ := http.Get("http://localhost:8081/stats")
	var proxyStats struct {
		Unique     int      `json:"unique_sequences"`
		Total      int      `json:"total_trans"`
		Undedected []string `json:"undetected_states"`
	}
	json.NewDecoder(resp.Body).Decode(&proxyStats)
	resp.Body.Close()

	stats.CoveredTrans = proxyStats.Unique
	// stats.CoveredTrans = proxyStats.Covered
	// stats.TotalTrans = proxyStats.Total
	stats.Coverage = float64(proxyStats.Unique) / float64(stats.TotalRequests) * 100

	printTestResults(stats, time.Since(start), proxyStats)

}

func printTestResults(stats Stats, duration time.Duration, proxyStats proxyStats) {
	fmt.Printf("\n=== Результаты теста ===\n")
	fmt.Printf("Всего запросов: %d\n", stats.TotalRequests)
	fmt.Printf("Обнаружено ошибок: сетевых %d, запр. команд %d, ошибок сервера %d\n", networkErrors, blockedErrors, serverErrors)
	fmt.Printf("Покрыто переходов: %d/%d (%.1f%%)\n",
		stats.CoveredTrans, stats.TotalRequests, stats.Coverage)

	fmt.Printf("Не обнаружены : %d\n", proxyStats.Undedected)
}
