package proxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"os/exec"
	"proxy/internal/models"
	"strconv"
	"strings"
)

type ProxyImpl struct {
}

func NewProxyServer() *ProxyImpl {
	return &ProxyImpl{}
}

func (ps *ProxyImpl) ListenAndServe() error {
	server := http.Server{
		Addr: "0.0.0.0:8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "CONNECT" {
				ps.proxyHTTPS(w, r)
			} else {
				ps.proxyHTTP(w, r)
			}
		}),
	}
	return server.ListenAndServe()
}

func (ps *ProxyImpl) proxyHTTPS(w http.ResponseWriter, r *http.Request) {
	localConn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		http.Error(w, "no upstream", http.StatusServiceUnavailable)
		return
	}
	defer localConn.Close()

	_, err = localConn.Write([]byte(models.ConnectionEstablished))
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	host := dnsName(r.Host)

	tlsConfig, err := GenTLSConf(host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tlsLocalConn := tls.Server(localConn, &tlsConfig)
	err = tlsLocalConn.Handshake()
	if err != nil {
		tlsLocalConn.Close()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tlsLocalConn.Close()

	remoteConn, err := tls.Dial("tcp", r.URL.Host, &tlsConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer remoteConn.Close()

	reader := bufio.NewReader(tlsLocalConn)
	request, err := http.ReadRequest(reader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	requestByte, err := httputil.DumpRequest(request, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = remoteConn.Write(requestByte)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	serverReader := bufio.NewReader(remoteConn)
	response, err := http.ReadResponse(serverReader, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rawResponse, err := httputil.DumpResponse(response, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tlsLocalConn.Write(rawResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//request.URL.Scheme = "https"
	//hostAndPort := strings.Split(r.URL.Host, ":")
	//request.URL.Host = hostAndPort[0]

	//reqId, err := p.repo.SaveRequest(request)
	//if err != nil {
	//	log.Printf("Error save:  %v", err)
	//}
	//
	//_, err = p.repo.SaveResponse(reqId, response)
	//if err != nil {
	//	log.Printf("Error save:  %v", err)
	//}
}

func (ps *ProxyImpl) proxyHTTP(w http.ResponseWriter, r *http.Request) {
	r.RequestURI = strings.Replace(r.RequestURI, r.URL.Scheme+"://"+r.URL.Host, "", 1)
	r.Header.Del("Proxy-Connection")

	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	// Cookies parses and returns the cookies set in the Set-Cookie headers
	resp.Cookies()
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// dnsName returns the DNS name in addr, if any.
func dnsName(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	}
	return host
}

func GenTLSConf(host string) (tls.Config, error) {
	cmd := exec.Command("/certs/gen_cert.sh", host, strconv.Itoa(rand.Int()))

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		log.Fatal(fmt.Sprint(err) + ": " + stderr.String())
	}

	tlsCert, err := tls.LoadX509KeyPair("/certs/nck.crt", "/certs/cert.key")
	if err != nil {
		return tls.Config{}, err
	}

	return tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}, nil
}
