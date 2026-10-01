package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /upload", uploadFile)
	mux.HandleFunc("POST /upload-multiple", uploadMultipleFiles)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("server is running...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func isAllowedContentType(file multipart.File) error {
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		fmt.Println(err)
		return err
	}

	fmt.Println("read: ", n)

	contentType := http.DetectContentType(buffer[:n])
	// we just read file upto 512 bytes
	// so the cursor now at 512 , we have to reset cursor to position 0
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to reset file: %v", err)
	}

	switch contentType {
	case "image/jpeg",
		"image/png",
		"image/webp", "application/pdf":
		return nil
	default:
		return fmt.Errorf("content type not allowd: %s", contentType)
	}
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 5 << 20

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	err := r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "failed to parse form: file too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Println(file)
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}

	defer file.Close()

	fmt.Println(header.Filename)
	fmt.Println(header.Header)
	fmt.Println(header.Size)
	fmt.Println(header.Header.Get("content-type"))

	if err := isAllowedContentType(file); err != nil {
		fmt.Println(file)
		http.Error(w, "unsupported file", http.StatusBadRequest)
		return
	}

	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	dstPath := filepath.Join("uploads", filename)
	fmt.Println(dstPath)

	dst, err := os.Create(dstPath)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "failed to create file", http.StatusInternalServerError)
		return
	}

	defer dst.Close()

	n, err := io.Copy(dst, file)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "uploaded successfully: %s, written: %d, title: %s, description: %s", header.Filename, n, r.FormValue("title"), r.FormValue("description"))
}

func uploadMultipleFiles(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 20 << 20

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		log.Println(err)
		http.Error(
			w,
			"invalid multipart form", http.StatusBadRequest,
		)
		return
	}

	files := r.MultipartForm.File["images"]

	if len(files) == 0 {
		http.Error(
			w,
			"images are reqquired",
			http.StatusBadRequest,
		)

		return
	}

	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			http.Error(
				w,
				"failed to open file",
				http.StatusInternalServerError,
			)
			return
		}

		/*
			IMPORTANT:
			Do not defer inside a potentially huge loop
			Close each file on done processing
		*/

		if err := isAllowedContentType(file); err != nil {
			file.Close()
			http.Error(
				w,
				"unsupported file",
				http.StatusBadRequest,
			)
			return
		}

		ext := filepath.Ext(header.Filename)
		filename := fmt.Sprintf(
			"%d%s",
			time.Now().UnixNano(),
			ext,
		)
		dstPath := filepath.Join("uploads", filename)

		dst, err := os.Create(dstPath)

		if err != nil {
			file.Close()
			http.Error(w, "failed to create file", http.StatusInternalServerError)
			return
		}

		_, err = io.Copy(dst, file)

		dst.Close()
		file.Close()

		if err != nil {
			http.Error(w, "failed to save file", http.StatusInternalServerError)
			return
		}

	}

	fmt.Fprintf(w, "%d Files uploaded successfully", len(files))

}
