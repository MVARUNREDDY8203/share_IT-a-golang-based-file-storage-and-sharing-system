package files

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sendit/db"
	"time"
)

// File represents a file record
type File struct {
    ID        int       `json:"id"`
    Filename  string    `json:"filename"`
    FileType  string    `json:"file_type"`
    FilePath  string    `json:"file_path"`
    UserID    int       `json:"user_id"`
    CreatedAt time.Time `json:"created_at"`
}

// EncryptFile encrypts the file data using AES encryption
func EncryptFile(file multipart.File, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }

    fileData, err := io.ReadAll(file)
    if err != nil {
        return nil, err
    }

    encrypted := gcm.Seal(nonce, nonce, fileData, nil)
    return encrypted, nil
}

// DecryptFile decrypts the encrypted file data using AES-GCM
func DecryptFile(encryptedData []byte, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonceSize := gcm.NonceSize()
    if len(encryptedData) < nonceSize {
        return nil, fmt.Errorf("encrypted data too short")
    }

    nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

    decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, err
    }

    return decrypted, nil
}


func SaveFile(w http.ResponseWriter, r *http.Request, userID int) {
    file, header, err := r.FormFile("file")
    if err != nil {
        log.Println("Error retrieving file from form:", err)
        http.Error(w, "Invalid file upload", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Define where to store the encrypted file on the server
    encryptedFilePath := fmt.Sprintf("./uploads/%d_%s.enc", userID, header.Filename)
    outFile, err := os.Create(encryptedFilePath)
    if err != nil {
        log.Println("Error creating encrypted file:", err)
        http.Error(w, "Unable to save file", http.StatusInternalServerError)
        return
    }
    defer outFile.Close()

    // Generate a 32-byte encryption key (this should be stored securely)
    encryptionKey := []byte("this_is_a_32_byte_key_for_aes_on") // Replace this with a secure key management strategy

    // Encrypt the file data
    encryptedData, err := EncryptFile(file, encryptionKey)
    if err != nil {
        log.Println("Error encrypting file:", err)
        http.Error(w, "Unable to encrypt file", http.StatusInternalServerError)
        return
    }

    // Write the encrypted data to the file
    _, err = outFile.Write(encryptedData)
    if err != nil {
        log.Println("Error writing encrypted data to file:", err)
        http.Error(w, "Unable to save encrypted file", http.StatusInternalServerError)
        return
    }

    // Save encrypted file metadata to the database and retrieve file ID
    result, err := db.DB.Exec("INSERT INTO files (filename, file_type, file_path, user_id, created_at) VALUES (?, ?, ?, ?, ?)",
        header.Filename, http.DetectContentType([]byte(header.Filename)), encryptedFilePath, userID, time.Now())
    if err != nil {
        log.Println("Error saving encrypted file metadata:", err)
        http.Error(w, "Unable to save file metadata", http.StatusInternalServerError)
        return
    }

    // Get the last inserted file ID
    fileID, err := result.LastInsertId()
    if err != nil {
        log.Println("Error retrieving file ID:", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Generate the absolute path for the file on the server
    absPath, err := filepath.Abs(encryptedFilePath)
    if err != nil {
        log.Println("Error getting absolute file path:", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Generate the publicly accessible URL
    publicURL := fmt.Sprintf("http://localhost:8080/files/access/%d", fileID)

    // Return the success message, file path, and public URL
    response := map[string]string{
        "message": "File uploaded and encrypted successfully",
        "file_path": absPath,
        "public_url": publicURL,
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

