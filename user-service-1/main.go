package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
)

type transaction struct {
	ID        string            `json:"transaction_id"`
	UserID    int               `json:"user_id"`
	Amount    float64           `json:"amount"`
	Currency  string            `json:"currency"`
	Type      string            `json:"type"`
	Metadata  map[string]string `json:"metadata"`
	Timestamp string            `json:"timestamp"`
}

type TransactionEvent struct {
	EventType   string      `json:"event_type"`
	Transaction transaction `json:"transaction"`
	ServiceID   string      `json:"service_id"`
	Timestamp   string      `json:"timestamp"`
}

var transactions = []transaction{
	{
		ID:        "b35d8def-0025-440f-9d0e-b0aab008a093",
		UserID:    1009,
		Amount:    42.38,
		Currency:  "KES",
		Type:      "credit",
		Metadata:  map[string]string{"merchant": "Shopify"},
		Timestamp: "2025-08-27T15:42:57.289182Z",
	},
	{
		ID:        "9b2c85f0-7641-45f0-bf21-219a93c20f5f",
		UserID:    1012,
		Amount:    120.50,
		Currency:  "USD",
		Type:      "debit",
		Metadata:  map[string]string{"merchant": "Amazon"},
		Timestamp: "2025-08-27T16:10:11.582739Z",
	},
	{
		ID:        "f67291b1-41df-4d32-8b87-7c7f2f4d1d88",
		UserID:    1003,
		Amount:    15.75,
		Currency:  "EUR",
		Type:      "credit",
		Metadata:  map[string]string{"merchant": "Netflix"},
		Timestamp: "2025-08-27T17:22:33.192847Z",
	},
}

// Kafka writer (publisher)
var kafkaWriter *kafka.Writer

func initKafkaWriter() {
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "0.0.0.0:9092"
	}

	// Writer (Producer)
	kafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(kafkaBrokers),
		Topic:        "transactions",
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        true,
	}

	log.Printf("Kafka writer initialized with brokers: %s", kafkaBrokers)
}

func publishTransactionEvent(eventType string, txn transaction) error {
	event := TransactionEvent{
		EventType:   eventType,
		Transaction: txn,
		ServiceID:   "user-service-2",
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(txn.ID),
		Value: eventBytes,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(eventType)},
			{Key: "user-id", Value: []byte(fmt.Sprintf("%d", txn.UserID))},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = kafkaWriter.WriteMessages(ctx, message)
	if err != nil {
		log.Printf("Failed to publish transaction event: %v", err)
		return err
	}

	log.Printf("Published %s event for transaction %s", eventType, txn.ID)
	return nil
}

// controllers
// Health check endpoint
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "user-service-2",
		"timestamp": time.Now().Unix(),
	})
}

func getTransactions(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, transactions)
}

func transactionById(c *gin.Context) {
	id := c.Param("id")
	transaction, err := getTransactionById(id)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Transaction not found."})
		return
	}
	c.IndentedJSON(http.StatusOK, transaction)
}

func getTransactionById(id string) (*transaction, error) {
	for i, b := range transactions {
		if b.ID == id {
			return &transactions[i], nil
		}
	}

	return nil, errors.New("transaction not found")
}

func createTransaction(c *gin.Context) {
	var newTransaction transaction

	if err := c.BindJSON(&newTransaction); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction data"})
		return
	}

	// Set timestamp if not provided
	if newTransaction.Timestamp == "" {
		newTransaction.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	// Add transaction to local storage
	transactions = append(transactions, newTransaction)

	// Publish transaction created event to Kafka
	go func() {
		if err := publishTransactionEvent("transaction_created", newTransaction); err != nil {
			log.Printf("Failed to publish transaction created event: %v", err)
		}
	}()

	c.IndentedJSON(http.StatusCreated, newTransaction)
}

func indexHandler(c *gin.Context) {
	c.String(http.StatusOK, "Welcome to the User Service 1!")
}

func main() {
	// create a router
	router := gin.Default()

	// define routes
	router.GET("/", indexHandler)
	router.GET("/transactions", getTransactions)
	router.GET("/transactions/id", transactionById)
	router.POST("/transactions", createTransaction)

	// run the server
	router.Run("0.0.0.0:3000")
}
