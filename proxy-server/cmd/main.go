package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"proxy-server/pkg/mongoclient"
	"proxy-server/pkg/repository"
	"proxy-server/proxy-server/internal/delivery"
	pkgCert "proxy-server/proxy-server/pkg/cert"
)

var (
	hostname, _ = os.Hostname()

	dir      = path.Join(os.Getenv("HOME"), ".mitm")
	keyFile  = path.Join(dir, "ca-key.pem")
	certFile = path.Join(dir, "ca-cert.pem")
)

const URI = "mongodb://root:root@localhost:27017"

func main() {
	log.SetPrefix("[PROXY] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	client, closeConn, err := mongoclient.NewMongoClient(URI)
	if err != nil {
		log.Fatal(err)
	}
	defer closeConn()
	repo, err := repository.NewRepository(client)
	if err != nil {
		log.Fatal(err)
	}

	middleware := delivery.NewMiddleware(&repo)

	ca, err := loadCA()
	if err != nil {
		log.Fatal(err)
	}

	proxyHandler := &delivery.Proxy{
		CA: &ca,
		TLSServerConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			//CipherSuites: []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA},
		},
		Wrap: middleware.Save,
	}

	log.Println("proxy :8080")
	log.Fatal(http.ListenAndServe(":8080", proxyHandler))
}

func loadCA() (cert tls.Certificate, err error) {
	// TODO(kr): check file permissions
	cert, err = tls.LoadX509KeyPair(certFile, keyFile)
	if os.IsNotExist(err) {
		cert, err = genCA()
	}
	if err == nil {
		cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	}
	return
}

func genCA() (cert tls.Certificate, err error) {
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return
	}
	// сертификат (в нем откр. ключ) и приватный ключ
	certPEM, keyPEM, err := pkgCert.GenCA(hostname)
	if err != nil {
		return
	}
	cert, _ = tls.X509KeyPair(certPEM, keyPEM)
	err = ioutil.WriteFile(certFile, certPEM, 0400)
	if err == nil {
		err = ioutil.WriteFile(keyFile, keyPEM, 0400)
	}
	return cert, err
}
