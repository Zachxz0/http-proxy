package main

import (
	"crypto/tls"
	"fmt"
	"github.com/go-ini/ini"
	"io"
	"log"
	"net/http"
	"net/url"
)

func handler(w http.ResponseWriter, r *http.Request) {
	transfer := r.Header.Get("transfer_url")
	transferUrl, e := url.Parse(transfer)
	if e != nil {
		w.WriteHeader(500)
		return
	}
	if transferUrl.Host == r.Host {
		w.WriteHeader(500)
		return
	}

	requestUrl := transfer + r.RequestURI
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: false},
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: tr,
	}
	req, e := http.NewRequest(r.Method, requestUrl, r.Body)
	if e != nil {
		log.Println("http new request error:", e)
		return
	}
	for k, v := range r.Header {
		for _, h := range v {
			req.Header.Add(k, h)
		}
	}
	req.Header.Set("Cookie", r.Header.Get("Cookie"))

	resp, e := client.Do(req)
	if e != nil {
		log.Println("http do error:", e)
		return
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		log.Println("read error:", e)
		return
	}
	for k, v := range resp.Header {
		for _, h := range v {
			w.Header().Add(k, h)
		}
	}
	w.Header().Set("Cookie", resp.Header.Get("Cookie"))
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func main() {
	cfg, e := ini.Load("conf/config.ini")
	if e != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", e))
	}
	config := cfg.Section("").KeysHash()

	http.HandleFunc("/", handler)
	err := http.ListenAndServeTLS(config["Listen"], config["CertFile"], config["KeyFile"], nil)
	if err != nil {
		log.Fatalln(err.Error())
	}
}
