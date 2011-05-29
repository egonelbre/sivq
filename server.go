package main

import (
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"template"
	"http"
	"json"
	"strings"
	"websocket"
	"image"
	"image/png"
	"log"
	"strconv"
	"gob"
)

const (
	UploadDir = "img/upload/"
	ResultDir = "img/result/"
	VectorDir = "img/vec/"
	TemplateDir = "template/"
	StaticDir = "static/"
)

type UploadResult struct {
	Error bool
	Message string
	Image string
}

type UploadPage struct {
	VectorFiles string
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
	// load vector files
	vectorFiles, _ := ioutil.ReadDir(VectorDir);
	vectorFilesString := "<option value=\"\"></option>"
	for _, file := range vectorFiles {
		if (file.Size > 5) {
			vectorFilesString = vectorFilesString + "<option value=\""+ file.Name +"\">"+ file.Name +"</option>"
		}
	}
	page := &UploadPage{VectorFiles: vectorFilesString}
	
	uploadTemplate.Execute(w, page);
}

/*
 * Handle uploading images
 */
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
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
	http.ServeFile(w, r, "img/" + fileName);
}

/*
 * Handle static files
 */
func staticHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Path[8:]
	http.ServeFile(w, r, StaticDir + fileName)
}

/*
 * Get messages from client
 */
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
        	ws.Write([]byte("Connection closed."+ err.String()));
			break
		}

		workChan <- Work{ws, &input}
    }
}

/*
 * Hub for processing work
 */
func hub() {
    for {
        select {
		case work := <-workChan:
			work.conn.Write([]byte("0"))
			process(work.input)

			response, _ := json.MarshalForHTML(&UploadResult{Image: work.input.Image, Error: false, Message: "Processed."})
			work.conn.Write(response)
			work.conn.Close()
        }
    }
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
	http.HandleFunc("/saveVector", uploadErrorHandler(saveVectorHandler))
	http.Handle("/process", websocket.Handler(clientHandler))
	http.ListenAndServe(":8080", nil)
}



func process(input *ProcessInput) {
    // open input file
    inputFile, err := os.OpenFile(UploadDir + input.Image, os.O_RDONLY, 0666)
    if err != nil {
        log.Fatalln("No input defined");
    }
    defer inputFile.Close()
    
    // create output file
    outputFile, err := os.OpenFile(ResultDir + input.Image, os.O_CREATE | os.O_WRONLY, 0666)
    if err != nil {
		log.Fatalln(err)
    }
    defer outputFile.Close()

    // decode png image
    inputImage, _, err := image.Decode(inputFile)
    if err != nil {
		log.Fatalln(err)
    }
    rgbaInput := rgba(inputImage)
    
    vectorParams := RingVectorParameters{ 
        Radius : input.Radius,
        Count : input.VectorRings,
        RadiusInc : input.RingSizeInc}
    
    sivqParams := SIVQParameters {
        GammaAdjustment : float(input.GammaAdjust),
        RotationStride : float(input.RotationStride),
        MatchingStride : input.MatchStride,
        MatchingOffset : input.MatchingOffset,
        Threshold : float(input.Threshold)}

    ringVector := NewRingVector(vectorParams)
    ringVector.LoadData(rgbaInput, input.VecX, input.VecY)

    outputImage := SIVQ(sivqParams, rgbaInput, ringVector)
    
    if err = png.Encode(outputFile, outputImage); err != nil {
        log.Fatalln(err) 
    } 
}

/*
 * Create vector and save to file
 */
func saveVectorHandler(w http.ResponseWriter, r *http.Request) {
	vectorName := r.FormValue("vectorName");
	imageName := r.FormValue("image");
	radius, err := strconv.Atoi(r.FormValue("radius"))
	checkError(err)
	vectorRings, err := strconv.Atoi(r.FormValue("vectorRings"))
	checkError(err)
	ringSizeInc, err := strconv.Atoi(r.FormValue("ringSizeInc"))
	checkError(err)
	vecX, err := strconv.Atoi(r.FormValue("vecX"))
	checkError(err)
	vecY, err := strconv.Atoi(r.FormValue("vecY"))
	checkError(err)
	
	// open input file
    inputFile, err := os.OpenFile(UploadDir + imageName, os.O_RDONLY, 0666)
    checkError(err)
    defer inputFile.Close()
    
    // create output file
    outputFile, err := os.OpenFile(VectorDir + vectorName, os.O_CREATE | os.O_WRONLY, 0666)
	checkError(err)
    defer outputFile.Close()

	// decode png image
    inputImage, _, err := image.Decode(inputFile)
    checkError(err)
    rgbaInput := rgba(inputImage)

    // create vector
	vectorParams := RingVectorParameters{ 
        Radius : radius,
        Count : vectorRings,
        RadiusInc : ringSizeInc}
	ringVector := NewRingVector(vectorParams)
	ringVector.LoadData(rgbaInput, vecX, vecY)

    // save into file
	encoder := gob.NewEncoder(outputFile)
	e := encoder.Encode(ringVector)
	checkError(e);

	jsonResponse, _ := json.MarshalForHTML(&UploadResult{Image: vectorName, Error: false, Message: "Saved."});
	fmt.Fprint(w, string(jsonResponse))
}