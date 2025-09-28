package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
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

// controllers
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
		return
	}

	transactions = append(transactions, newTransaction)
	c.IndentedJSON(http.StatusCreated, newTransaction)
}

func indexHandler(c *gin.Context) {
	c.String(http.StatusOK, "Welcome to the User Service 2!")
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
	router.Run("localhost:3001")
}
