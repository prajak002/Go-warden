package handler

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/go-sql-driver/mysql"
)

type AuditRequest struct {
	Content  string `json:"content"`
	Url      string `json:"url"`
	Action   string `json:"action"`
	TimeZone string `json:"timezone"`
	UserID   string `json:"user_id"`
}

type URLInfo struct {
	Domain         string
	PathWithParams string
}

func Audit(w http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var auditReq AuditRequest
	err = json.Unmarshal(body, &auditReq)
	if err != nil {
		http.Error(w, "Failed to parse request JSON", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(auditReq.Action) != "C" && strings.TrimSpace(auditReq.Action) != "P" && strings.TrimSpace(auditReq.Action) == "U" {
		http.Error(w, "Invalid Action", http.StatusInternalServerError)
		return
	}

	if ((strings.TrimSpace(auditReq.Action) == "C" || strings.TrimSpace(auditReq.Action) == "P") && strings.TrimSpace(auditReq.Content) == "") || strings.TrimSpace(auditReq.Action) == "" || strings.TrimSpace(auditReq.TimeZone) == "" || strings.TrimSpace(auditReq.Url) == "" || strings.TrimSpace(auditReq.UserID) == "" {
		http.Error(w, "Fields cannot be empty", http.StatusInternalServerError)
		return
	}

	urlInfo, err := parseURL(auditReq.Url)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusInternalServerError)
		return
	}

	clientIP := getIPAddress(request)

	DB_URL := os.Getenv("DB_URL")
	DB_PORT := os.Getenv("DB_PORT")
	DB_NAME := os.Getenv("DB_NAME")
	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	if DB_URL == "" || DB_PORT == "" || DB_NAME == "" || DB_USER == "" || DB_PASSWORD == "" {
		http.Error(w, "Failed to read environment variables for database", http.StatusBadRequest)
	}

	mysql.RegisterTLSConfig("tidb", &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: DB_URL,
	})

	db, err := sql.Open("mysql", DB_USER+":"+DB_PASSWORD+"@tcp("+DB_URL+":"+DB_PORT+")/"+DB_NAME+"?tls=tidb")
	if err != nil {
		log.Fatal("failed to connect database", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Fatal("Error starting transaction:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Create a new Node with a Node number of 1
	node, err := snowflake.NewNode(1)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fullContent := ""
	shortContent := ""
	if strings.TrimSpace(auditReq.Action) == "C" || strings.TrimSpace(auditReq.Action) == "P" {
		fullContent = strings.TrimSpace(auditReq.Content)
		shortContent = createShortContent(fullContent)
	}

	//generate Twitter Snowflake ID
	id := node.Generate()
	currentTime := time.Now()
	utcTime := currentTime.UTC()

	insertQuery := "INSERT INTO AUDIT_REQUEST_MASTER(ARM_ROW_ID, USER_ID, USER_IP_ADDR,USER_ACTION,CONTENT_SHORT,URL_DOMAIN,URL_PATH,CREATED_DATETIME, TIME_ZONE) VALUES (?, ?,?,?,?,?,?,?,?)"
	_, err = tx.Exec(insertQuery, id.String(), strings.TrimSpace(auditReq.UserID), clientIP, strings.TrimSpace(auditReq.Action), shortContent, urlInfo.Domain, urlInfo.PathWithParams, utcTime, strings.TrimSpace(auditReq.TimeZone))
	if err != nil {
		tx.Rollback()
		log.Fatal("Error inserting data:", err)
		http.Error(w, "Error inserting data:", http.StatusInternalServerError)
		return
	}
	if len(fullContent) > 30 {
		insertQuery = "INSERT INTO AUDIT_REQUEST_CONTENT(ARM_ROW_ID, CONTENT_FULL) VALUES (?, ?)"
		_, err = tx.Exec(insertQuery, id.String(), fullContent)
		if err != nil {
			tx.Rollback()
			log.Fatal("Error inserting data:", err)
			http.Error(w, "Error inserting data:", http.StatusInternalServerError)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		log.Fatal("Error committing transaction:", err)
		http.Error(w, "Error committing transaction:", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"status\":\"ok\"}"))
}

func getIPAddress(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}

func parseURL(urlString string) (*URLInfo, error) {
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}
	domain := strings.TrimPrefix(parsedURL.Hostname(), "www.")
	pathWithParams := parsedURL.Path + parsedURL.RawQuery
	info := &URLInfo{
		Domain:         domain,
		PathWithParams: pathWithParams,
	}
	return info, nil
}

func createShortContent(content string) string {
	if len(content) > 30 {
		return content[:30]
	} else {
		return content
	}
}
