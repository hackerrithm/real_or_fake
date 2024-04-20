package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// ImageResponse represents the response returned by the API
type ImageResponse struct {
	Answer string `json:"answer"`
	Img    string `json:"img"`
}

// Folder names
var folders = []string{"fake", "real"}

// S3 bucket and region
var bucketName = "your-bucket-name"
var region = "your-region"

// Create an AWS session
var sess = session.Must(session.NewSession(&aws.Config{
	Region: aws.String(region),
}))

// Create an S3 client
var svc = s3.New(sess)

func main() {
	http.HandleFunc("/image", getImageHandler)
	http.HandleFunc("/compare", compareHandler)

	http.ListenAndServe(":8080", &corsHandler{})
}

type corsHandler struct{}

func (*corsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Accept-Language, Accept-Encoding")
	http.DefaultServeMux.ServeHTTP(w, r)
}

func getImageHandler(w http.ResponseWriter, r *http.Request) {
	// Get the parent directory of the current file
	parentDir := filepath.Dir("main.go")

	// Get a list of files in the "real" and "fake" folders
	realFiles, err := os.ReadDir(filepath.Join(parentDir, "../images/real"))
	if err != nil {
		log.Fatal(err)
	}
	fakeFiles, err := os.ReadDir(filepath.Join(parentDir, "../images/fake"))
	if err != nil {
		log.Fatal(err)
	}

	// Choose a folder at random
	var files []os.DirEntry
	var folderName string
	if rand.Intn(2) == 0 {
		files = realFiles
		folderName = "real"
	} else {
		files = fakeFiles
		folderName = "fake"
	}

	// Choose a file at random from the chosen folder
	randFile := files[rand.Intn(len(files))]

	// Get the full path of the file
	filePath := filepath.Join(filepath.Dir(randFile.Name()), randFile.Name())

	resp := ImageResponse{
		Answer: folderName,
		Img:    filePath,
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

func compareHandler(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var request ImageResponse
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Extract folder name from image path
	parts := strings.Split(request.Img, "/")
	imgFolder := parts[len(parts)-2]

	// Compare response with request
	if request.Answer == imgFolder {
		// Send tick response
		w.Write([]byte("✓"))

		// Make another API call to get new image
		resp, err := http.Get("http://localhost:8080/image")
		if err != nil {
			fmt.Println("Error fetching new image:", err)
			return
		}
		defer resp.Body.Close()
		var newImage ImageResponse
		err = json.NewDecoder(resp.Body).Decode(&newImage)
		if err != nil {
			fmt.Println("Error decoding new image response:", err)
			return
		}
		// Here you can send newImage back to the frontend to display the new image
	} else {
		// Send cross response
		w.Write([]byte("❌"))
	}
}
