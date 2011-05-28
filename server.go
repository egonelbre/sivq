package main

import (
	"fmt"
	"os"
	"io"
	"template"
	"http"
	"json"
	"strings"
	"websocket"
	"time"
)

const (
	UploadDir = "upload/"
	TemplateDir = "template/"
	StaticDir = "static/"
)

type UploadResult struct {
	Error bool
	Message string
	Image string
}

type ProcessInput struct {
	Image string
	VecX int
	VecY int
	Radius int
	VectorRings int
	RingSizeInc int
	Threshold float64
	RotationStride float64
	MatchStride int
	MatchingOffset int
	GammaAdjust float64
}

type Work struct {
    conn		*websocket.Conn
    input		*ProcessInput
}

var (
	uploadTemplate = template.MustParseFile(TemplateDir + "upload.html", nil)
	errorTemplate = template.MustParseFile(TemplateDir + "error.html", nil)
	workChan = make(chan Work)
)

/*
 * Check for error and panic if needed
 */
func checkError(e os.Error) {
	if e != nil {
		panic(e)
	}
}

/*
 * Index page handler
 */
func indexHandler(w http.ResponseWriter, r *http.Request) {
	uploadTemplate.Execute(w, nil);
}

/*
 * Handle uploading images
 */
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// just show upload form
	if r.Method != "POST" {
		//uploadTemplate.Execute(w, nil)
		return
	}

	f, fHeader,err := r.FormFile("image")
	checkError(err)
	defer f.Close()

	// check file extension
	fileName := strings.TrimSpace(fHeader.Filename);
	fileType := strings.ToLower(fileName[len(fileName) - 3:])
	if (fileType != "png" && fileType != "jpg") {
		panic(os.NewError("Invalid file type."))
	}

	t, err := os.Create(UploadDir + fileName)
	checkError(err)
	defer t.Close()

	_, e := io.Copy(t, f)
	checkError(e)

	// JSON response
	jsonResponse, _ := json.MarshalForHTML(&UploadResult{Image: fileName, Error: false, Message: "Image uploaded."});
	fmt.Fprint(w, string(jsonResponse))
}

/*
 * Display image
 */
func imgHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Path[5:]
	w.Header().Set("Content-Type", "image")
	http.ServeFile(w, r, UploadDir + fileName);
}

/*
 * Handle static files
 */
func staticHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Path[8:]
	http.ServeFile(w, r, StaticDir + fileName)
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
 * Handle uploading errors
 */
func uploadErrorHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(os.Error); ok {
				jsonResponse, _ := json.MarshalForHTML(&UploadResult{Image: "", Error: true, Message: e.String()});
				fmt.Fprint(w, string(jsonResponse))
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
	
	go hub()

	http.HandleFunc("/", errorHandler(indexHandler))
	http.HandleFunc("/upload/", uploadErrorHandler(uploadHandler))
	http.HandleFunc("/static/", errorHandler(staticHandler))
	http.HandleFunc("/img/", errorHandler(imgHandler))
	http.Handle("/process", websocket.Handler(clientHandler))
	http.ListenAndServe(":8080", nil)
}



func hub() {
    for {
        select {
		case work := <-workChan:
			// DO WHAT IS NEEDED
			
			work.conn.Write([]byte("25"));
			time.Sleep(1000000000);
			work.conn.Write([]byte("50"));
			time.Sleep(1000000000);
			work.conn.Write([]byte("75"));
			time.Sleep(1000000000);

			response, _ := json.MarshalForHTML(&UploadResult{Image: "letters.png", Error: false, Message: "Processed."})
			work.conn.Write(response);
			work.conn.Close();
        }
    }
}


func clientHandler(ws *websocket.Conn) {
    defer func() {
        ws.Close()
    }()

    buf := make([]byte, 256)
    var input ProcessInput
    for {
        n, err := ws.Read(buf)
        if err != nil {
            break
        }

        // get data
        err = json.Unmarshal(buf[0:n], &input);
        if err != nil {
        	// close connection
        	ws.Write([]byte("Connection closed."+ err.String()));
			break
		}
		
		ws.Write([]byte("0"));
		workChan <- Work{ws, &input}
    }
}