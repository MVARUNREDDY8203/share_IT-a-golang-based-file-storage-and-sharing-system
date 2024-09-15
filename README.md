# [ShareIt: Secure File Sharing Service](youtube.com) 🚀

ShareIt: a robust and secure file sharing service built with Go, designed to provide users with a safe and efficient way to upload, store, and share files.

## Table of Contents
- [Features](#-features)
- [Technologies Used](#️-technologies-used)
- [Project Structure](#-project-structure)
- [Getting Started](#-getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation and Project Setup](#installation-and-project-setup)
    - [Option 1: Local Setup](#option-1-local-setup)
    - [Option 2: Docker Setup](#option-2-docker-setup)
- [API Endpoints](#-api-endpoints)
- [Security Features](#️-security-features)
- [Bonus Tasks Completed](#bonus-tasks-completed)
- [Project Functionality / Documentation](#project-functionality--documentation)
  - [Login/SignUp - JWT Authentication](#loginsignup--jwt-authentication)
  - [Database: SQL Schema](#database-sql-schema)
  - [File Storage and Management](#file-storage-and-management)
  - [Redis Layer](#redis-layer)
  - [Background Worker](#background-worker)

## 🌟 Features

- User registration and authentication
- Secure file upload/storage with (AES) encryption/management
- File search functionality
- Easy file sharing via public URLs
- Caching layer for file metadata
- Automatic file deletion on file expiry using a background worker
- Rate limiting using Redis to prevent abuse

## 🛠️ Technologies Used

- **Backend**: Go (Golang)
- **Database**: MySQL (Database)
- **Caching**: Redis (Caching meta-data and Rate-limiting) [Upstash.com](Upstash.com)
- **Routing**: Gorilla Mux
- **Authentication**: JWT (JSON Web Tokens)
- **Encryption**: AES (Advanced Encryption Standard)
- **Containerization**: Docker

## 📁 Project Structure

```
shareit/
├── main.go                 # Application entry point
├── auth/                   # Authentication module
│   ├── auth.go
│   └── jwt.go
├── db/                     # Database management
│   ├── db.go
│   └── redis.go
├── files/                  # File operations
│   ├── manage.go
│   └── upload.go
├── rate_limiter/           # Rate limiting
│   └── limiter.go
├── background_worker/      # Cleanup operations
│   └── worker.go
├── Dockerfile              # Docker configuration
└── init.sql                # SQL initialization script
```

## 🚀 Getting Started

### Prerequisites

- Go (version 1.16 or later)
- MySQL
- [Redis - Upstash.com](Upstash.com)
- Docker (optional)

### Installation and Project Setup

#### Option 1: Local Setup

1. Clone the repository:
   ```sh
   git clone https://github.com/MVARUNREDDY8203/21brs1507_backend.git
   cd shareit
   ```

2. Set up environment variables:
   Create a `.env` file in the root directory:
   ```
   DB_USER=your_mysql_username
   DB_PASSWORD=your_mysql_password
   DB_NAME=shareit
   DB_HOST=localhost:3306
   CLEANUP_INTERVAL=24h     - 1 day
   EXPIRY_DURATION=168h     - 7 days
   REDIS_URL=your_upstash_redis_url_with_password:port
   ```

3. Set up the MySQL database:
   ```sh
   mysql -u root -p
   CREATE DATABASE shareit;
   USE shareit;
   ```
   Run these queries to initialize the database tables - `users` and `files`:
   ```sql
   CREATE TABLE IF NOT EXISTS users (
        id INT AUTO_INCREMENT PRIMARY KEY,
        email VARCHAR(255) NOT NULL UNIQUE,
        password VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS files (
        id INT AUTO_INCREMENT PRIMARY KEY,
        filename VARCHAR(255),
        file_type VARCHAR(50),
        file_path VARCHAR(255),
        user_id INT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users(id)
    );
   ```

4. Install dependencies:
   ```sh
   go mod tidy
   ```

5. Build and run the application:
   ```sh
   go build
   ./shareit
   ```

6. Run the background worker:
   ```sh
   cd background_worker
   go run worker.go
   ```

#### Option 2: Docker Setup

1. Clone the repository:
   ```sh
   git clone https://github.com/MVARUNREDDY8203/21brs1507_backend.git
   cd shareit
   ```

2. Build the Docker image:
   ```sh
   docker build -t shareit .
   ```

3. Run the Docker container:
   ```sh
   docker run -p 8080:8080 -e REDIS_URL=your_upstash_redis_url_with_password:port shareit
   ```

   Note: Replace `your_upstash_redis_url_with_password:port` with your actual Redis URL from Upstash.

4. The application and background worker will start automatically inside the container.

## 🔗 API Endpoints

| Method | Endpoint | Description | Authentication | Payload | Response |
|--------|----------|-------------|----------------|-------|------|
| POST | `/signup` | User registration | No | {email, password} | successful id creation message |
| POST | `/login` | User login | No | {email, password} | jwt_token |
| POST | `/files/upload` | Upload a file | Yes | file - {chosen file} | {message, public_url} |

| Method | Endpoint | Description | Authentication | Query | Response |
|--------|----------|-------------|----------------|-------|------|
| GET | `/files/search` | Search for files | Yes | nil/ file_id/ file_name/ file_type | {file_id, name, path, user_id, created_at} |
| DELETE | `/files/delete` | Delete a file | Yes | file_id | response message |
| GET | `/files/share` | Get a shareable link | Yes | file_id | public_url |
| GET | `/files/access/{file_id}` | Serves public_url of files | No | file_id | decrypted_file_requested |

## 🛡️ Security Features

- 🔒 Password hashing with bcrypt
- 🔐 AES file encryption - files are encrypted at rest
- 🎫 JWT-based authentication
- 🚦 Request rate limiting - Redis Layer

## BONUS TASKS COMPLETED

- [x] File Encryption
- [x] Rate Limiting
- [ ] Websockets for notifications
- [x] Containerization with Docker
- [ ] Hosting on AWS 

## Project Functionality / Documentation

### Login/SignUp - JWT Authentication

- `auth` package:
    - `func GenerateJWT(email string)` 
        - Generates JWT Token
    - `func ValidateJWT(tokenString string)`
        - Validates JWT token 

### Database: SQL Schema

![image](https://github.com/user-attachments/assets/ebe2e84b-9a11-4bd9-b523-dca30f6c3dcb)

- `db` package:
    - `func ConnectDB()` 
        - Establishes and validates connection to MySQL database

### File Storage and Management

- Local storage is used; AWS S3 requires a credit card, and Google Firebase lacks a good SDK for Go.

- `files` package:
    - `func SaveFile(w http.ResponseWriter, r *http.Request, userID int)` 
        - Handles file uploads
        - Sets file metadata in the database
        - Encrypts the file 
        - Stores the file in local storage
    - `func SearchFile(w http.ResponseWriter, r *http.Request, userID int)` 
        - Searches for files based on query parameters:
        - `nil`: Returns all files created by users
        - `file_id`, `file_name`, `file_type`, `created_at`: Returns files matching the filters
    - `func ShareFile(w http.ResponseWriter, r *http.Request, userID int)` 
        - Returns a publicly accessible URL for the requested file with query parameter `file_id`
    - `func ServeFile(w http.ResponseWriter, r *http.Request)` 
        - Serves the file upon request
        - Decrypts and serves the requested files 
    - `func DeleteFile(w http.ResponseWriter, r *http.Request, userID int)` 
        - Deletes the file's metadata from the database and the file from storage using `file_id`
    - `func EncryptFile(file multipart.File, key []byte)` 
        - Uses AES encryption to encrypt the file 
    - `func DecryptFile(encryptedData []byte, key []byte)`
        - Uses AES decryption to decrypt the encrypted file

### Redis Layer

- `db` package:
    - `func RateLimit(userID int)` 
        - Returns `true` if requests are within the limit number, otherwise `false`
    - `func ConnectRedis()` 
        - Checks and validates connection with Redis server (Upstash)
    - `func CacheFileMetadata(fileID int, metadata string)` 
        - Sets metadata cache for the file with `

fileID` and `metadata`
    - `func GetCachedFileMetadata(fileID int)` 
        - Returns metadata of the file with `fileID` if present in cache

### Background Worker

- Runs `cleanupOldFiles()` every `$env:CLEANUP_INTERVAL` to clean up expired files

- `func cleanupOldFiles(expiryDuration time.Duration)` 
    - Cleans up files that have expired: from the database (metadata) and from local storage
- `func deleteOldFilesFromStorage(cutoffTime time.Time)` 
    - Deletes files from local storage
