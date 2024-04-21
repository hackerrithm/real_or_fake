package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/rs/cors"
)

// S3 bucket and region
var AWSID = os.Getenv("AWSID")
var AWSSecret = os.Getenv("AWSSECRET")
var port = os.Getenv("PORT")

// ImageResponse represents the response returned by the API
type ImageResponse struct {
	Answer string `json:"answer"`
	Img    string `json:"img"`
}

// Folder names
var folders = []string{"fake", "real"}

// S3 bucket and region
var bucketName = "real-fake-images"
var region = "us-east-1"

// Create an AWS session
var sess = session.Must(session.NewSession(&aws.Config{
	Region: aws.String(region),
	Credentials: credentials.NewStaticCredentials(
		AWSID,     //os.Getenv("AWSID"),     //cred.AWSID,
		AWSSecret, //os.Getenv("AWSSECRET"), //cred.AWSSecret,
		""),
}))

// Create an S3 client
var svc = s3.New(sess)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/image", getImageHandler)
	mux.HandleFunc("/compare", compareHandler)

	// Configure CORS
	c := cors.AllowAll() // Allow all origins

	// Create a handler chain with CORS
	handler := c.Handler(mux)

	// Start server with the CORS-enabled handler
	log.Fatal(http.ListenAndServe("localhost:"+port, handler))
}

func getImageHandler(w http.ResponseWriter, r *http.Request) {
	// // Get the parent directory of the current file
	// parentDir := filepath.Dir("main.go")

	// // Get a list of files in the "real" and "fake" folders
	// realFiles, err := os.ReadDir(filepath.Join(parentDir, "../images/real"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fakeFiles, err := os.ReadDir(filepath.Join(parentDir, "../images/fake"))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Choose a folder at random
	// var files []os.DirEntry
	// var folderName string
	// if rand.Intn(2) == 0 {
	// 	files = realFiles
	// 	folderName = "real"
	// } else {
	// 	files = fakeFiles
	// 	folderName = "fake"
	// }

	// // Choose a file at random from the chosen folder
	// randFile := files[rand.Intn(len(files))]

	// // Get the full path of the file
	// filePath := filepath.Join(filepath.Dir(randFile.Name()), randFile.Name())

	// resp := ImageResponse{
	// 	Answer: folderName,
	// 	Img:    filePath,
	// }

	// jsonResp, err := json.Marshal(resp)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// w.Header().Set("Content-Type", "application/json")
	// w.Write(jsonResp)

	// Choose a random folder
	folder := folders[rand.Intn(len(folders))]

	// List objects in the folder
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(fmt.Sprintf("%s/", folder)),
	})
	if err != nil {
		log.Println("Error listing objects:", err)
		http.Error(w, "Failed to list objects", http.StatusInternalServerError)
		return
	}
	if len(resp.Contents) == 0 {
		log.Println("No images found in the folder")
		http.Error(w, "No images found in the folder", http.StatusInternalServerError)
		return
	}

	// Randomly select an image
	imageKey := *resp.Contents[rand.Intn(len(resp.Contents))].Key

	// Create and return the response
	newResp := ImageResponse{
		Answer: folder,
		Img:    imageKey,
	}

	jsonResp, err := json.Marshal(newResp)
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

	// Extract folder name from image key
	parts := strings.Split(request.Img, "/")
	folder := parts[0] // Assume the folder name is the first part of the key

	// Compare response with request
	if request.Answer == folder {
		// Send tick response
		w.Write([]byte("✓"))

		// Make another API call to get new image
		resp, err := http.Get("localhost:" + port + "/image")
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
