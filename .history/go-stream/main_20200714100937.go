package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func main() {
	http.Handle("/", handlers())
	http.ListenAndServe(":8000", nil)
}

func handlers() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", indexPage).Methods("GET")
	router.HandleFunc("/media/{folder}/stream/", streamHandler).Methods("GET")
	router.HandleFunc("/media/{folder}/stream/{segName:index[0-2]+.ts}", streamHandler).Methods("GET")
	return router
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func streamHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	mId, err := strconv.Atoi(vars["mId"])
	if err != nil {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	folder, err := vars["folder"]
	if err != nil {
		response.WriteHeader(http.StatusNotFound)
		return
	}

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

func getMediaBase(folder string) string {
	mediaRoot := "m3u8s"
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
