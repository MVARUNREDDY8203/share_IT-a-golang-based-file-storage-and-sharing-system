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

    cacheKey := fmt.Sprintf("files:user:%d:file_id=%s:file_type=%s:file_name=%s:file_path=%s",
        userID, fileID, fileType, fileName, filePath)

    // Check if the results are cached
    cachedData, err := db.GetCachedFileMetadata(cacheKey)
    if err == nil {
        // If data is found in the cache, return the cached data
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(cachedData))
        return
    }

    // If cache is missed, query the database
    query := "SELECT id, filename, file_type, file_path, user_id, created_at FROM files WHERE user_id = ?"
    args := []interface{}{userID}

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
        file.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
        files = append(files, file)
    }

    if len(files) == 0 {
        http.Error(w, "No files found", http.StatusNotFound)
        return
    }

    // Cache the retrieved data
    jsonData, _ := json.Marshal(files)
    db.CacheFileMetadata(cacheKey, string(jsonData))

    // Return the file metadata
    w.WriteHeader(http.StatusOK)
    w.Write(jsonData)
}

func ShareFile(w http.ResponseWriter, r *http.Request, userID int) {
    fileID := r.URL.Query().Get("file_id")
    if fileID == "" {
        http.Error(w, "file_id is required", http.StatusBadRequest)
        return
    }

    // Check if the shareable URL is cached in Redis
    cachedURL, err := db.GetCachedFileMetadata(fileID)
    if err == nil {
        // Cache hit: Return the cached URL
        response := map[string]string{
            "file_id": fileID,
            "url":     cachedURL,
        }
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(response)
        return
    }

    // If not found in cache, generate a new shareable URL
    shareableURL := "http://localhost:8080/files/access/" + fileID

    // Cache the generated shareable URL for future use (expires in 10 minutes)
    err = db.CacheFileMetadata(fileID, shareableURL)
    if err != nil {
        log.Println("Error caching file metadata:", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Return the new shareable URL
    response := map[string]string{
        "file_id": fileID,
        "url":     shareableURL,
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}



// ShareFile allows users to get the sharable URL of a file based on file_id
func ServeFile(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    fileID := vars["file_id"]

    // Check if file path is cached
    cachedFilePath, err := db.GetCachedFileMetadata(fileID)
    if err == nil {
        // Cache hit: Use the cached file path
        serveFileFromPath(w, cachedFilePath, fileID)
        return
    }

    // If cache is missed, query the database
    var encryptedFilePath string
    var originalFilename string
    err = db.DB.QueryRow("SELECT file_path, filename FROM files WHERE id = ?", fileID).Scan(&encryptedFilePath, &originalFilename)
    if err == sql.ErrNoRows {
        log.Println("File not found:", err)
        http.Error(w, "File not found", http.StatusNotFound)
        return
    } else if err != nil {
        log.Println("Error retrieving file path:", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Cache the file path for future requests
    db.CacheFileMetadata(fileID, encryptedFilePath)

    // Serve the file from the file path
    serveFileFromPath(w, encryptedFilePath, fileID)
}

// Helper function to serve file from a given path
func serveFileFromPath(w http.ResponseWriter, filePath string, fileID string) {
    encryptedData, err := os.ReadFile(filePath)
    if err != nil {
        log.Println("Error reading encrypted file:", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Decrypt the file content
    encryptionKey := []byte("this_is_a_32_byte_key_for_aes_on")
    decryptedData, err := DecryptFile(encryptedData, encryptionKey)
    if err != nil {
        log.Println("Error decrypting file:", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=file_%s", fileID))
    w.Header().Set("Content-Type", "application/octet-stream")
    w.Header().Set("Content-Length", fmt.Sprintf("%d", len(decryptedData)))

    w.Write(decryptedData)
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
