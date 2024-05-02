package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/otiai10/gosseract/v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	http.HandleFunc("/heartbeat", heartbeatHandler)
	http.HandleFunc("/enabled", enabledHandler)
	http.HandleFunc("/ocr_text", ocrHandler)
	host := "127.0.0.1"
	port := "10070"
	if envHost, exists := os.LookupEnv("APP_HOST"); exists {
		host = envHost
	}
	if envPort, exists := os.LookupEnv("APP_PORT"); exists {
		port = envPort
	}
	log.Fatal(http.ListenAndServe(host+":"+port, nil))
}

type Payload map[string]interface{}

func getNcURL() string {
	url := os.Getenv("NEXTCLOUD_URL")
	url = strings.TrimSuffix(url, "/index.php")
	url = strings.TrimSuffix(url, "/")
	return url
}

func signCheck(r *http.Request) (string, error) {
	appID := r.Header.Get("EX-APP-ID")
	if appID != os.Getenv("APP_ID") {
		return "", fmt.Errorf("invalid APP ID: %v", appID)
	}
	appVersion := r.Header.Get("EX-APP-VERSION")
	if appVersion != os.Getenv("APP_VERSION") {
		return "", fmt.Errorf("invalid APP VERSION: %v", appVersion)
	}
	decodedBytes, err := base64.StdEncoding.DecodeString(r.Header.Get("AUTHORIZATION-APP-API"))
	if err != nil {
		return "", fmt.Errorf("failed to decode string: %v", err)
	}
	decodedString := string(decodedBytes)
	parts := strings.SplitN(decodedString, ":", 2)
	userName := parts[0]
	appSecret := parts[1]
	if appSecret != os.Getenv("APP_SECRET") {
		return "", fmt.Errorf("invalid APP SECRET: %v", appSecret)
	}
	return userName, nil
}

func ocsCall(method, url string, username string, data Payload) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error encoding payload to JSON: %v", err)
	}
	fullUrl := getNcURL() + url
	req, err := http.NewRequest(method, fullUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("EX-APP-ID", os.Getenv("APP_ID"))
	req.Header.Set("EX-APP-VERSION", os.Getenv("APP_VERSION"))
	req.Header.Set("OCS-APIRequest", "true")
	req.Header.Set("AUTHORIZATION-APP-API", base64.StdEncoding.EncodeToString([]byte(username+":"+os.Getenv("APP_SECRET"))))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	return string(body), nil
}

func davCall(method, url string, username string, data []byte) ([]byte, error) {
	fullUrl := getNcURL() + "/remote.php/dav" + url
	req, err := http.NewRequest(method, fullUrl, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}
	req.Header.Set("EX-APP-ID", os.Getenv("APP_ID"))
	req.Header.Set("EX-APP-VERSION", os.Getenv("APP_VERSION"))
	req.Header.Set("OCS-APIRequest", "true")
	req.Header.Set("AUTHORIZATION-APP-API", base64.StdEncoding.EncodeToString([]byte(username+":"+os.Getenv("APP_SECRET"))))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	return body, nil
}

func heartbeatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		fmt.Println("heartbeatHandler called.")
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"status": "ok"}
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func enabledHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		fmt.Println("enabledHandler called.")
		_, err := signCheck(r)
		if err != nil {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
		value := r.URL.Query().Get("enabled")
		isEnabled, err := strconv.ParseBool(value)
		if err != nil {
			http.Error(w, "Invalid boolean value", http.StatusBadRequest)
			return
		}
		fmt.Printf("enabledHandler: %v\n", isEnabled)
		r := ""
		if isEnabled {
			_, err := ocsCall("POST", "/ocs/v1.php/apps/app_api/api/v1/ui/files-actions-menu", "", Payload{
				"name":          "ocr_text",
				"displayName":   "Optical Text",
				"mime":          "image/png, image/jpeg",
				"permissions":   31,
				"actionHandler": "/ocr_text",
			})
			if err != nil {
				r = err.Error()
			}
		} else {
			_, err := ocsCall("DELETE", "/ocs/v1.php/apps/app_api/api/v1/ui/files-actions-menu", "", Payload{
				"name": "ocr_text",
			})
			if err != nil {
				r = err.Error()
			}
		}
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"error": r}
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

type UiFileActionHandlerInfo struct {
	FileId     int    `json:"fileId"`
	Name       string `json:"name"`
	Directory  string `json:"directory"`
	Etag       string `json:"etag"`
	Mime       string `json:"mime"`
	FileType   string `json:"fileType"`
	Mtime      int    `json:"mtime"`
	Size       int    `json:"size"`
	UserId     string `json:"userId"`
	InstanceId string `json:"instanceId"`
}

func ocrHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		fmt.Println("ocrHandler called.")
		userId, err := signCheck(r)
		if err != nil {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		var data UiFileActionHandlerInfo
		err = json.Unmarshal(body, &data)
		if err != nil {
			http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		go func() {
			davFileInPath := ""
			if data.Directory == "/" {
				davFileInPath = "/files/" + userId + "/" + data.Name
			} else {
				davFileInPath = "/files/" + userId + data.Directory + "/" + data.Name
			}
			oldExt := filepath.Ext(davFileInPath)
			davFileOutPath := davFileInPath[:len(davFileInPath)-len(oldExt)] + ".txt"
			fmt.Println(davFileInPath)
			fmt.Println(davFileOutPath)
			tmpfile, err := ioutil.TempFile("", "*.temp"+oldExt)
			if err != nil {
				log.Fatal(err)
				http.Error(w, "Failed create temp file", http.StatusBadRequest)
				return
			}
			defer os.Remove(tmpfile.Name())
			defer tmpfile.Close()
			fileData, err := davCall("GET", davFileInPath, userId, nil)
			if err != nil {
				log.Fatal(err)
				http.Error(w, "Failed to get file", http.StatusBadRequest)
				return
			}
			if _, err := tmpfile.Write(fileData); err != nil {
				log.Fatalf("Failed to write to temporary file: %v", err)
				http.Error(w, "Failed to write tmp file", http.StatusBadRequest)
				return
			}

			//here use tesseract lib
			client := gosseract.NewClient()
			defer client.Close()
			client.SetImage(tmpfile.Name())
			text, _ := client.Text()

			_, err = davCall("PUT", davFileOutPath, userId, []byte(text))
			if err != nil {
				log.Fatal(err)
				http.Error(w, "Failed to upload result", http.StatusBadRequest)
				return
			}
		}()
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
