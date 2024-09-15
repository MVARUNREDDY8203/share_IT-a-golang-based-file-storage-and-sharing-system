package files

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"shareit/db"
	"time"

	"github.com/gorilla/mux"
)


func SearchFile(w http.ResponseWriter, r *http.Request, userID int) {
    fileID := r.URL.Query().Get("file_id")
    fileType := r.URL.Query().Get("file_type")
    fileName := r.URL.Query().Get("file_name")
    filePath := r.URL.Query().Get("file_path")

    query := "SELECT id, filename, file_type, file_path, user_id, created_at FROM files WHERE user_id = ?"
    args := []interface{}{userID}
    log.Println(fileID, fileType, fileName, filePath)
    // Add filters to the query if they are provided
    if fileID != "" {
        query += " AND id = ?"
        args = append(args, fileID)
    }
    if fileType != "" {
        query += " AND file_type = ?"
        args = append(args, fileType)
    }
    if fileName != "" {
        query += " AND filename LIKE ?"
        args = append(args, "%"+fileName+"%")
    }
    if filePath != "" {
        query += " AND file_path LIKE ?"
        args = append(args, "%"+filePath+"%")
    }

    rows, err := db.DB.Query(query, args...)
    if err != nil {
        log.Println("Error retrieving files:", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var files []File
    for rows.Next() {
        var file File
        var createdAtStr string
        if err := rows.Scan(&file.ID, &file.Filename, &file.FileType, &file.FilePath, &file.UserID, &createdAtStr); err != nil {
            log.Println("Error scanning file:", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }
        // Parse the createdAtStr into a time.Time
        file.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
        if err != nil {
            log.Println("Error parsing created_at timestamp:", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }
        files = append(files, file)
    }

    if len(files) == 0 {
        http.Error(w, "No files found", http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(files)
}

// ShareFile allows users to get the sharable URL of a file based on file_id
func ShareFile(w http.ResponseWriter, r *http.Request, userID int) {
    fileID := r.URL.Query().Get("file_id")
    if fileID == "" {
        http.Error(w, "file_id is required", http.StatusBadRequest)
        return
    }

    var filePath string
    err := db.DB.QueryRow("SELECT file_path FROM files WHERE id = ? AND user_id = ?", fileID, userID).Scan(&filePath)
    if err == sql.ErrNoRows {
        log.Println("File not found:", err)
        http.Error(w, "File not found", http.StatusNotFound)
        return
    } else if err != nil {
        log.Println("Error retrieving file path:", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Generate a public URL to access the file
    shareableURL := "http://localhost:8080/files/access/" + fileID

    // Return the shareable URL
    response := map[string]string{
        "file_id": fileID,
        "url":     shareableURL,
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

// ServeFile allows users to access a file publicly via its file_id
func ServeFile(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    fileID := vars["file_id"]

    // Get the encrypted file path from the database
    var encryptedFilePath string
    var originalFilename string
    err := db.DB.QueryRow("SELECT file_path, filename FROM files WHERE id = ?", fileID).Scan(&encryptedFilePath, &originalFilename)
    if err == sql.ErrNoRows {
        log.Println("File not found:", err)
        http.Error(w, "File not found", http.StatusNotFound)
        return
    } else if err != nil {
        log.Println("Error retrieving file path:", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Read the encrypted file content
    encryptedData, err := os.ReadFile(encryptedFilePath)
    if err != nil {
        log.Println("Error reading encrypted file:", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Decrypt the file content
    encryptionKey := []byte("this_is_a_32_byte_key_for_aes_on") // Ensure you use the correct key
    decryptedData, err := DecryptFile(encryptedData, encryptionKey)
    if err != nil {
        log.Println("Error decrypting file:", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Serve the decrypted content with the original filename
    w.Header().Set("Content-Disposition", "attachment; filename="+originalFilename)
    w.Header().Set("Content-Type", "application/octet-stream")
    w.Header().Set("Content-Length", fmt.Sprintf("%d", len(decryptedData)))

    // Write the decrypted content to the response
    _, err = w.Write(decryptedData)
    if err != nil {
        log.Println("Error serving decrypted file:", err)
    }
}

// DeleteFile allows users to delete their files
func DeleteFile(w http.ResponseWriter, r *http.Request, userID int) {
    fileID := r.URL.Query().Get("file_id")

    var filePath string
    err := db.DB.QueryRow("SELECT file_path FROM files WHERE id = ? AND user_id = ?", fileID, userID).Scan(&filePath)
    if err == sql.ErrNoRows {
		log.Println(err)

        http.Error(w, "File not found", http.StatusNotFound)
        return
    } else if err != nil {
        log.Println("Error retrieving file path:", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Delete the file from local storage
    err = os.Remove(filePath)
    if err != nil {
		log.Println(err)

        http.Error(w, "Error deleting file", http.StatusInternalServerError)
        return
    }

    // Delete the metadata from the database
    _, err = db.DB.Exec("DELETE FROM files WHERE id = ? AND user_id = ?", fileID, userID)
    if err != nil {
		log.Println(err)

        http.Error(w, "Error deleting file metadata", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "File deleted successfully"})
}
