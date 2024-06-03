package metrics

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/janithht/GoStreamBalancer/database"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartMetricsServer() {
	database.InitDB()
	mux := http.NewServeMux()

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	handlerWithCORS := handlers.CORS(originsOk, headersOk, methodsOk)(mux)

	mux.Handle("/metrics", promhttp.HandlerFor(CustomRegistry, promhttp.HandlerOpts{}))
	mux.Handle("/connections", http.HandlerFunc(getConnections))

	if err := http.ListenAndServe(":8000", handlerWithCORS); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		return
	}
}

func getConnections(w http.ResponseWriter, r *http.Request) {
	clientIP := r.URL.Query().Get("client_ip")
	serverURL := r.URL.Query().Get("server_url")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	query := "SELECT client_ip, server_url, timestamp FROM connections WHERE 1=1"
	args := []interface{}{}

	if clientIP != "" {
		query += " AND client_ip = ?"
		args = append(args, clientIP)
	}
	if serverURL != "" {
		query += " AND server_url = ?"
		args = append(args, serverURL)
	}
	if startDate != "" {
		query += " AND timestamp >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND timestamp <= ?"
		args = append(args, endDate)
	}

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to execute query", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var connections []map[string]interface{}
	for rows.Next() {
		var clientIP, serverURL, timestamp string
		if err := rows.Scan(&clientIP, &serverURL, &timestamp); err != nil {
			http.Error(w, "Failed to scan row", http.StatusInternalServerError)
			return
		}
		connections = append(connections, map[string]interface{}{
			"client_ip":  clientIP,
			"server_url": serverURL,
			"timestamp":  timestamp,
		})
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to iterate over rows", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(connections); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
