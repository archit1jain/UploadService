package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form. 32MB max memory, rest goes to temp files.
		// We actually stream the parts below to avoid loading everything into memory.
		reader, err := r.MultipartReader()
		if err != nil {
			http.Error(w, "Could not parse multipart form", http.StatusBadRequest)
			return
		}

		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				http.Error(w, "Error reading multipart form", http.StatusInternalServerError)
				return
			}

			// We only care about parts with a filename
			filename := part.FileName()
			if filename == "" {
				continue
			}

			// Basic path traversal prevention: get just the base name
			filename = filepath.Base(filename)

			// If filename is invalid after Base (e.g., "." or "/"), generate a safe one or skip
			if filename == "." || filename == "/" {
				filename = "upload.dat"
			}

			destPath, err := getUniqueFilepath(uploadDir, filename)
			if err != nil {
				log.Printf("Error generating unique filename: %v", err)
				http.Error(w, "Error preparing file", http.StatusInternalServerError)
				return
			}

			dst, err := os.Create(destPath)
			if err != nil {
				log.Printf("Error creating file %s: %v", destPath, err)
				http.Error(w, "Error saving file", http.StatusInternalServerError)
				return
			}

			// Stream directly from the part to disk
			_, err = io.Copy(dst, part)
			dst.Close()

			if err != nil {
				log.Printf("Error saving file %s: %v", destPath, err)
				os.Remove(destPath) // Try to clean up partial file
				http.Error(w, "Error saving file", http.StatusInternalServerError)
				return
			}
			log.Printf("Saved: %s", destPath)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Upload complete"))
	})

	addr := ":" + port
	log.Printf("Server listening on %s", addr)
	log.Printf("Saving uploads to %s", uploadDir)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// getUniqueFilepath checks if a file exists and appends (1), (2), etc. if it does.
func getUniqueFilepath(dir, filename string) (string, error) {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)

	for i := 0; i < 10000; i++ { // Prevent infinite loop
		name := filename
		if i > 0 {
			name = fmt.Sprintf("%s (%d)%s", base, i, ext)
		}

		path := filepath.Join(dir, name)

		// Prevent traversal attacks just in case
		cleanPath := filepath.Clean(path)
		if !strings.HasPrefix(cleanPath, filepath.Clean(dir)+string(filepath.Separator)) && filepath.Clean(dir) != cleanPath {
			return "", fmt.Errorf("invalid path: %s", path)
		}

		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			return path, nil
		} else if err != nil {
			return "", err // Other error (e.g., permission denied)
		}
		// File exists, try next iteration
	}
	return "", fmt.Errorf("could not find a unique filename for %s", filename)
}
