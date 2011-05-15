package main

import (
	"fmt"
	"os"
	"io"
	"template"
	"http"
)

const UploadDir = "upload/"
const TemplateDir = "template/"

type ResultPage struct {
	UploadedImage string
}

var uploadTemplate = template.MustParseFile(TemplateDir + "upload.html", nil)
var resultTemplate = template.MustParseFile(TemplateDir + "result.html", nil)
var errorTemplate = template.MustParseFile(TemplateDir + "error.html", nil)

/*
 * Check for error and panic if needed
 */
func checkError(e os.Error) {
	if e != nil {
		panic(e)
	}
}

/*
 * Handle uploading images
 */
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// just show upload form
	if r.Method != "POST" {
		uploadTemplate.Execute(w, nil)
		return
	}
	
	f, fHeader,err := r.FormFile("image")
	checkError(err)
	defer f.Close()
	
	t, err := os.Create(UploadDir + fHeader.Filename)
	checkError(err)
	defer t.Close()

	_, e := io.Copy(t, f)
	checkError(e)

	resultTemplate.Execute(w, &ResultPage{UploadedImage: fHeader.Filename})
}

/*
 * Display image
 */
func imgHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image")
	http.ServeFile(w, r, UploadDir + r.FormValue("file"));
}

/*
 * Handle errors
 */
func errorHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(os.Error); ok {
				w.WriteHeader(500)
				errorTemplate.Execute(w, e)
			}
		}()
		fn(w, r)
	}
}

/*
 * Start server
 */
func main() {
	fmt.Println("Server started.")
	http.HandleFunc("/", errorHandler(uploadHandler))
	http.HandleFunc("/img", errorHandler(imgHandler))
	http.ListenAndServe(":8080", nil)
}
