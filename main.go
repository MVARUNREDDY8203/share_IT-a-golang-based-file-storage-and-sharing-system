package main

import (
	"log"
	"net/http"
	"sendit/auth"
	"sendit/db"
	"sendit/files"

	"github.com/gorilla/mux"
)


func main() {
    db.ConnectDB()
    db.ConnectRedis()

    router := mux.NewRouter()

    router.HandleFunc("/signup", auth.SignupHandler).Methods("POST")
    router.HandleFunc("/login", auth.LoginHandler).Methods("POST")

    router.HandleFunc("/files/upload", auth.RequireAuth(files.SaveFile)).Methods("POST")
    router.HandleFunc("/files/search", auth.RequireAuth(files.SearchFile)).Methods("GET")
    router.HandleFunc("/files/delete", auth.RequireAuth(files.DeleteFile)).Methods("DELETE")
    router.HandleFunc("/files/share", auth.RequireAuth(files.ShareFile)).Methods("GET")
	router.HandleFunc("/files/access/{file_id}", files.ServeFile).Methods("GET")

    log.Fatal(http.ListenAndServe(":8080", router))
}
