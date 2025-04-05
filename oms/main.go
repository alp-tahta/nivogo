package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const (
	dbHost     = "postgres-oms"
	dbPort     = 5432
	dbUser     = "postgres"
	dbPassword = "example"
	dbName     = "oms"
)

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type OrderItem struct {
	Product  Product `json:"product"`
	Quantity int     `json:"quantity"`
}

type Order struct {
	ID        string      `json:"id"`
	Items     []OrderItem `json:"items"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

type OrderSaga struct {
	OrderID   string
	Status    string
	Step      int
	CreatedAt time.Time
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Database connection
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create tables if they don't exist
	createTables(db)

	// HTTP server setup
	mux := http.NewServeMux()

	// Order endpoints
	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreateOrder(w, r, db)
		case http.MethodGet:
			handleGetOrders(w, r, db)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Printf("Starting OMS service on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func createTables(db *sql.DB) {
	ordersTable := `
		CREATE TABLE IF NOT EXISTS orders (
			id VARCHAR(36) PRIMARY KEY,
			status VARCHAR(20) NOT NULL,
			created_at TIMESTAMP NOT NULL
		);
	`

	orderItemsTable := `
		CREATE TABLE IF NOT EXISTS order_items (
			id SERIAL PRIMARY KEY,
			order_id VARCHAR(36) NOT NULL,
			product_id VARCHAR(36) NOT NULL,
			product_name VARCHAR(255) NOT NULL,
			product_description TEXT,
			quantity INTEGER NOT NULL,
			FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
		);
	`

	sagaTable := `
		CREATE TABLE IF NOT EXISTS order_sagas (
			order_id VARCHAR(36) PRIMARY KEY,
			status VARCHAR(20) NOT NULL,
			step INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL
		);
	`

	_, err := db.Exec(ordersTable)
	if err != nil {
		log.Fatalf("Failed to create orders table: %v", err)
	}

	_, err = db.Exec(orderItemsTable)
	if err != nil {
		log.Fatalf("Failed to create order_items table: %v", err)
	}

	_, err = db.Exec(sagaTable)
	if err != nil {
		log.Fatalf("Failed to create saga table: %v", err)
	}
}

func handleCreateOrder(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		sendError(w, http.StatusBadRequest, "INVALID_REQUEST", "Geçersiz istek formatı", "İstek gövdesi JSON formatında olmalıdır")
		return
	}

	// Validate order
	if order.ID == "" {
		sendError(w, http.StatusBadRequest, "MISSING_ORDER_ID", "Sipariş ID'si eksik", "Sipariş ID'si boş olamaz")
		return
	}

	if len(order.Items) == 0 {
		sendError(w, http.StatusBadRequest, "EMPTY_ORDER", "Sipariş boş", "Sipariş en az bir ürün içermelidir")
		return
	}

	// Start saga
	saga := OrderSaga{
		OrderID:   order.ID,
		Status:    "STARTED",
		Step:      0,
		CreatedAt: time.Now(),
	}

	// Save saga state
	_, err := db.Exec(
		"INSERT INTO order_sagas (order_id, status, step, created_at) VALUES ($1, $2, $3, $4)",
		saga.OrderID, saga.Status, saga.Step, saga.CreatedAt,
	)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "SAGA_CREATION_FAILED", "Saga oluşturulamadı", "Sipariş işlemi başlatılamadı")
		return
	}

	// Step 1: Reserve inventory for all items
	for _, item := range order.Items {
		if item.Quantity <= 0 {
			sendError(w, http.StatusBadRequest, "INVALID_QUANTITY", "Geçersiz miktar",
				fmt.Sprintf("Ürün ID: %s için miktar 0'dan büyük olmalıdır", item.Product.ID))
			return
		}

		if err := reserveInventory(item.Product.ID, item.Quantity); err != nil {
			// Compensating transaction: Release all previously reserved inventory
			for _, releasedItem := range order.Items {
				if releasedItem.Product.ID == item.Product.ID {
					break // Skip the current item as it wasn't reserved
				}
				releaseInventory(releasedItem.Product.ID, releasedItem.Quantity)
			}
			updateSagaStatus(db, saga.OrderID, "FAILED")
			sendError(w, http.StatusInternalServerError, "INVENTORY_RESERVATION_FAILED",
				"Stok rezervasyonu başarısız",
				fmt.Sprintf("Ürün ID: %s için %d adet stok rezerve edilemedi", item.Product.ID, item.Quantity))
			return
		}
	}

	// Update saga step
	updateSagaStatus(db, saga.OrderID, "INVENTORY_RESERVED")

	// Step 2: Create order and order items
	tx, err := db.Begin()
	if err != nil {
		// Compensating transaction: Release all inventory
		for _, item := range order.Items {
			releaseInventory(item.Product.ID, item.Quantity)
		}
		updateSagaStatus(db, saga.OrderID, "FAILED")
		sendError(w, http.StatusInternalServerError, "TRANSACTION_FAILED", "İşlem başlatılamadı",
			"Veritabanı işlemi başlatılamadı")
		return
	}

	// Insert order
	_, err = tx.Exec(
		"INSERT INTO orders (id, status, created_at) VALUES ($1, $2, $3)",
		order.ID, "CREATED", time.Now(),
	)
	if err != nil {
		tx.Rollback()
		// Compensating transaction: Release all inventory
		for _, item := range order.Items {
			releaseInventory(item.Product.ID, item.Quantity)
		}
		updateSagaStatus(db, saga.OrderID, "FAILED")
		sendError(w, http.StatusInternalServerError, "ORDER_CREATION_FAILED", "Sipariş oluşturulamadı",
			"Sipariş veritabanına kaydedilemedi")
		return
	}

	// Insert order items
	for _, item := range order.Items {
		_, err = tx.Exec(
			"INSERT INTO order_items (order_id, product_id, product_name, product_description, quantity) VALUES ($1, $2, $3, $4, $5)",
			order.ID, item.Product.ID, item.Product.Name, item.Product.Description, item.Quantity,
		)
		if err != nil {
			tx.Rollback()
			// Compensating transaction: Release all inventory
			for _, releaseItem := range order.Items {
				releaseInventory(releaseItem.Product.ID, releaseItem.Quantity)
			}
			updateSagaStatus(db, saga.OrderID, "FAILED")
			sendError(w, http.StatusInternalServerError, "ORDER_ITEM_CREATION_FAILED", "Sipariş kalemi oluşturulamadı",
				fmt.Sprintf("Ürün ID: %s için sipariş kalemi kaydedilemedi", item.Product.ID))
			return
		}
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		// Compensating transaction: Release all inventory
		for _, item := range order.Items {
			releaseInventory(item.Product.ID, item.Quantity)
		}
		updateSagaStatus(db, saga.OrderID, "FAILED")
		sendError(w, http.StatusInternalServerError, "TRANSACTION_COMMIT_FAILED", "İşlem tamamlanamadı",
			"Veritabanı işlemi tamamlanamadı")
		return
	}

	// Update saga status to completed
	updateSagaStatus(db, saga.OrderID, "COMPLETED")

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func handleGetOrders(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// First, get all orders
	orderRows, err := db.Query("SELECT id, status, created_at FROM orders")
	if err != nil {
		sendError(w, http.StatusInternalServerError, "ORDER_FETCH_FAILED", "Siparişler getirilemedi",
			"Veritabanından siparişler alınamadı")
		return
	}
	defer orderRows.Close()

	var orders []Order
	orderMap := make(map[string]*Order)

	for orderRows.Next() {
		var order Order
		err := orderRows.Scan(&order.ID, &order.Status, &order.CreatedAt)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "ORDER_SCAN_FAILED", "Sipariş verisi okunamadı",
				"Sipariş verisi veritabanından okunamadı")
			return
		}
		orders = append(orders, order)
		orderMap[order.ID] = &orders[len(orders)-1]
	}

	// Then, get all order items
	itemRows, err := db.Query("SELECT order_id, product_id, product_name, product_description, quantity FROM order_items")
	if err != nil {
		sendError(w, http.StatusInternalServerError, "ORDER_ITEMS_FETCH_FAILED", "Sipariş kalemleri getirilemedi",
			"Veritabanından sipariş kalemleri alınamadı")
		return
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var orderID string
		var item OrderItem
		err := itemRows.Scan(&orderID, &item.Product.ID, &item.Product.Name, &item.Product.Description, &item.Quantity)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "ORDER_ITEM_SCAN_FAILED", "Sipariş kalemi verisi okunamadı",
				"Sipariş kalemi verisi veritabanından okunamadı")
			return
		}
		if order, exists := orderMap[orderID]; exists {
			order.Items = append(order.Items, item)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func updateSagaStatus(db *sql.DB, orderID string, status string) {
	_, err := db.Exec("UPDATE order_sagas SET status = $1 WHERE order_id = $2", status, orderID)
	if err != nil {
		log.Printf("Failed to update saga status: %v", err)
	}
}

func reserveInventory(productID string, quantity int) error {
	// Create request body with quantity
	requestBody := struct {
		Quantity int `json:"quantity"`
	}{
		Quantity: quantity,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("stok rezervasyon isteği oluşturulamadı: %v", err)
	}

	// Call inventory service to reserve items
	resp, err := http.Post(
		fmt.Sprintf("http://inventory:8080/reserve/%s", productID),
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return fmt.Errorf("stok rezervasyon isteği gönderilemedi: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return fmt.Errorf("stok rezervasyonu başarısız: %s - %s", errorResp.Error, errorResp.Details)
		}
		return fmt.Errorf("stok rezervasyonu başarısız: %s", resp.Status)
	}
	return nil
}

func releaseInventory(productID string, quantity int) error {
	// Create request body with quantity
	requestBody := struct {
		Quantity int `json:"quantity"`
	}{
		Quantity: quantity,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("stok serbest bırakma isteği oluşturulamadı: %v", err)
	}

	// Call inventory service to release items
	resp, err := http.Post(
		fmt.Sprintf("http://inventory:8080/release/%s", productID),
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return fmt.Errorf("stok serbest bırakma isteği gönderilemedi: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return fmt.Errorf("stok serbest bırakma başarısız: %s - %s", errorResp.Error, errorResp.Details)
		}
		return fmt.Errorf("stok serbest bırakma başarısız: %s", resp.Status)
	}
	return nil
}

func sendError(w http.ResponseWriter, status int, code string, message string, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   message,
		Code:    code,
		Details: details,
	})
}
