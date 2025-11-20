package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Общий секретный ключ (должен быть у клиента и сервера)
const sharedSecret = os.Environ.Get("hmac-secret")

// Генерация HMAC подписи
func generateHMAC(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// СЕРВЕР: Middleware проверки HMAC
func hmacAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем подпись и timestamp из заголовков
		receivedSignature := r.Header.Get("X-Signature")
		timestamp := r.Header.Get("X-Timestamp")

		if receivedSignature == "" || timestamp == "" {
			http.Error(w, "Missing signature or timestamp", http.StatusUnauthorized)
			return
		}

		// Проверяем, что запрос не слишком старый (защита от replay-атак)
		requestTime, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			http.Error(w, "Invalid timestamp", http.StatusUnauthorized)
			return
		}

		if time.Now().Unix()-requestTime > 300 { // 5 минут
			http.Error(w, "Request too old", http.StatusUnauthorized)
			return
		}

		// Читаем тело запроса
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Cannot read body", http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Восстанавливаем body

		// Создаем сообщение для проверки (body + timestamp)
		message := string(bodyBytes) + timestamp
		expectedSignature := generateHMAC(message, sharedSecret)

		// Сравниваем подписи (защита от timing-атак)
		if !hmac.Equal([]byte(receivedSignature), []byte(expectedSignature)) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		fmt.Printf("✅ Valid HMAC signature from request\n")
		// Подпись валидна - продолжаем обработку
		next(w, r)
	}
}

// Защищенный endpoint
func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	var order struct {
		UserID int    `json:"user_id"`
		Action string `json:"action"`
	}
	json.NewDecoder(r.Body).Decode(&order)

	fmt.Printf("📦 Processing order: UserID=%d, Action=%s\n", order.UserID, order.Action)

	response := map[string]interface{}{
		"status":  "success",
		"message": "Order created successfully",
		"order":   order,
	}
	json.NewEncoder(w).Encode(response)
}

// КЛИЕНТ: Отправка запроса с HMAC
func sendRequestWithHMAC() {
	body := `{"user_id": 123, "action": "create_order"}`
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	// Создаем сообщение для подписи (body + timestamp)
	message := body + timestamp
	signature := generateHMAC(message, sharedSecret)

	fmt.Printf("\n🔐 Client sending request:\n")
	fmt.Printf("   Body: %s\n", body)
	fmt.Printf("   Timestamp: %s\n", timestamp)
	fmt.Printf("   Signature: %s\n\n", signature)

	// Отправляем запрос с заголовками
	req, _ := http.NewRequest("POST", "http://localhost:8080/orders",
		strings.NewReader(body))
	req.Header.Set("X-Signature", signature)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("📬 Server response (%d):\n%s\n", resp.StatusCode, string(responseBody))
}

// Попытка отправить запрос с неверной подписью
func sendInvalidRequest() {
	body := `{"user_id": 456, "action": "hack_system"}`
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	// Неверная подпись
	signature := "invalid-signature-12345"

	fmt.Printf("\n🔴 Client sending INVALID request:\n")
	fmt.Printf("   Body: %s\n", body)
	fmt.Printf("   Invalid Signature: %s\n\n", signature)

	req, _ := http.NewRequest("POST", "http://localhost:8080/orders",
		strings.NewReader(body))
	req.Header.Set("X-Signature", signature)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("📬 Server response (%d): %s\n", resp.StatusCode, string(responseBody))
}

func main() {
	// Запускаем сервер в горутине
	go func() {
		http.HandleFunc("/orders", hmacAuthMiddleware(createOrderHandler))
		fmt.Println("🚀 HMAC Auth Server started on :8080")
		http.ListenAndServe(":8080", nil)
	}()

	// Ждем, пока сервер запустится
	time.Sleep(500 * time.Millisecond)

	// Отправляем валидный запрос
	sendRequestWithHMAC()

	time.Sleep(500 * time.Millisecond)

	// Отправляем невалидный запрос
	sendInvalidRequest()

	// Держим программу запущенной
	fmt.Println("\n✨ Demo completed. Press Ctrl+C to exit.")
	select {}
}
