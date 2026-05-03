package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const maxBodySize = 8 << 20

var dataFilePath string

func getenv(key, def string) string {
	v := os.Getenv(key)
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func main() {
	dataFilePath = getenv("DATA_FILE", "/app/data/projects.json")
	if err := os.MkdirAll(filepath.Dir(dataFilePath), 0755); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/projects", handleProjects)
	mux.Handle("/", http.FileServer(http.Dir("public")))

	addr := ":" + getenv("PORT", "8080")
	log.Fatal(http.ListenAndServe(addr, mux))
}

func handleProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGet(w, r)
	case http.MethodPost:
		handlePost(w, r)
	default:
		logErrResp(r, http.StatusMethodNotAllowed, "method not allowed")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func logErrResp(r *http.Request, status int, reason string) {
	log.Printf("%s %s %s %d %s", r.Method, r.URL.Path, r.RemoteAddr, status, reason)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	b, err := os.ReadFile(dataFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			jsonOK(w, []byte("[]"))
			return
		}
		logErrResp(r, http.StatusInternalServerError, "read failed")
		http.Error(w, "read failed", http.StatusInternalServerError)
		return
	}
	if len(bytes.TrimSpace(b)) == 0 {
		jsonOK(w, []byte("[]"))
		return
	}
	var raw any
	if err := json.Unmarshal(b, &raw); err != nil {
		jsonOK(w, []byte("[]"))
		return
	}
	if _, ok := raw.([]any); !ok {
		jsonOK(w, []byte("[]"))
		return
	}
	jsonOK(w, b)
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	b, err := io.ReadAll(r.Body)
	if err != nil {
		logErrResp(r, http.StatusRequestEntityTooLarge, "body too large")
		http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
		return
	}
	trimmed := bytes.TrimSpace(b)
	if len(trimmed) == 0 || trimmed[0] != '[' {
		logErrResp(r, http.StatusBadRequest, "expected json array")
		http.Error(w, "expected json array", http.StatusBadRequest)
		return
	}
	var arr []any
	if err := json.Unmarshal(trimmed, &arr); err != nil {
		logErrResp(r, http.StatusBadRequest, "invalid json")
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	out, err := json.Marshal(arr)
	if err != nil {
		logErrResp(r, http.StatusInternalServerError, "marshal failed")
		http.Error(w, "marshal failed", http.StatusInternalServerError)
		return
	}
	if err := atomicWrite(dataFilePath, out); err != nil {
		logErrResp(r, http.StatusInternalServerError, "write failed")
		log.Printf("write_failed_detail data_file=%s err=%v", dataFilePath, err)
		http.Error(w, "write failed", http.StatusInternalServerError)
		return
	}
	jsonOK(w, []byte(`{"ok":true}`))
}

func jsonOK(w http.ResponseWriter, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(absDir, "projects-*.json")
	if err != nil {
		return err
	}
	tmpName := f.Name()
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(tmpName)
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(tmpName)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	dest, err := filepath.Abs(path)
	if err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, dest); err != nil {
		_ = os.Remove(dest)
		if err2 := os.Rename(tmpName, dest); err2 != nil {
			os.Remove(tmpName)
			return err2
		}
	}
	return nil
}
