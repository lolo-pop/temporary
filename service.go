package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type image struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func main() {
	http.HandleFunc("/processImages", processImagesHandler)
	http.ListenAndServe(":8081", nil)
}

func processImagesHandler(w http.ResponseWriter, r *http.Request) {
	var imageBatch []image
	err := json.NewDecoder(r.Body).Decode(&imageBatch)
	//imageBatch, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading image batch data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// images := bytes.Split(imageBatch, []byte{'\n'})
	processedImages := processImages(imageBatch)
	fmt.Printf("prcessed %v", processedImages)
	//w.Header().Set("Content-Type", "application/octet-stream")
	//returnData, err := json.Marshal(processedImages)
	//if err != nil {
	//	// 处理错误
	//}
	//w.Write(returnData)
	/*
		for _, img := range processedImages {
			w.Write(img)
			w.Write([]byte{'\n'})
		}
	*/
}

func processImages(images []image) []image {
	// Process images as needed
	processedImages := []image{}
	fmt.Printf("batch len is %d\n", len(images))
	fmt.Printf("batch %v\n", images)
	for _, img := range images {
		processedImages = append(processedImages, img)
	}
	return processedImages
}

func processImage(Image image) image {
	// Process a single image as needed
	return Image // Replace with actual processing
}
