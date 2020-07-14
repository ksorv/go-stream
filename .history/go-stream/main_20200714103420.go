package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

const uploadPath = "in"
mediaRoot := "m3u8s"

func main() {
	http.Handle("/", handlers())
	http.ListenAndServe(":8000", nil)
}

func handlers() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", indexPage).Methods("GET")
	router.HandleFunc("/media/{folder}/stream/index.m3u8", streamHandler).Methods("GET")
	router.HandleFunc("/media/{folder}/stream/{segName:index[0-2]+.ts}", streamHandler).Methods("GET")
	router.HandleFunc("/upload", uploadHandler)
	return router
}

func uploadHandler(response http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
			template, _ := template.ParseFiles("upload.gtpl")
			template.Execute(w, nil)
			return
		}

	}

}

func streamHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	folder := vars["folder"]

	segName, ok := vars["segName"]
	if !ok {
		mediaBase := getMediaBase(folder)
		m3u8Name := "index.m3u8"
		serveHlsM3u8(response, request, mediaBase, m3u8Name)
	} else {
		mediaBase := getMediaBase(folder)
		serveHlsTs(response, request, mediaBase, segName)
	}
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	template, _ := template.ParseFiles("index.gtpl")
			template.Execute(w, nil)
			return
}

func getMediaBase(folder string) string {
	return fmt.Sprintf("%s/%s", mediaRoot, folder)
}

func serveHlsM3u8(w http.ResponseWriter, r *http.Request, mediaBase, m3u8Name string) {
	mediaFile := fmt.Sprintf("%s/%s", mediaBase, m3u8Name)
	http.ServeFile(w, r, mediaFile)
	w.Header().Set("Content-Type", "application/x-mpegURL")
}

func serveHlsTs(w http.ResponseWriter, r *http.Request, mediaBase, segName string) {
	mediaFile := fmt.Sprintf("%s/%s", mediaBase, segName)
	http.ServeFile(w, r, mediaFile)
	w.Header().Set("Content-Type", "video/MP2T")
}
