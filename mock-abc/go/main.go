package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
)

func getClientIp() (string, error) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return "", err
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}

		}
	}

	return "", fmt.Errorf("Can not find the client ip address!")
}

func doReq(url string, requestId string) (content string) {

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("x-request-id", requestId)

	fmt.Printf("x-request-id :%s\n", requestId)

	resp, err := (&http.Client{}).Do(req)

	fmt.Printf("do req upstream url: %s\n", url)

	if err != nil {

		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {

		fmt.Println(err)
		return
	}

	return string(body)
}

func main() {

	// Create a mux for routing incoming requests
	m := http.NewServeMux()

	version := os.Getenv("version")
	app := os.Getenv("app")
	url := os.Getenv("upstream_url")
	bind := os.Getenv("bind")

	if bind == "" {
		bind = ":8000"
	}

	ip, err := getClientIp()

	if err != nil {
		log.Println(err)
	}

	// All URLs will be handled by this function
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		requestId := r.Header.Get("x-request-id")

		fmt.Printf("receive request: x-request-id: %s\n", requestId)

		response := fmt.Sprintf("-> %s(version: %s, ip: %s)", app, version, ip)

		if url != "" {

			content := doReq(url, requestId)
			response = response + content

		}
		w.Write([]byte(response))
	})

	// Create a server listening on port 8000
	s := &http.Server{
		Addr:    bind,
		Handler: m,
	}

	// Continue to process new requests until an error occurs
	log.Fatal(s.ListenAndServe())
}
