package main

import (
	"flag"
	"net/http"

	"github.com/golang/glog"

	"github.com/jw-s/updatey/pkg/client/docker"

	"github.com/jw-s/updatey/pkg/admission"
	"github.com/jw-s/updatey/pkg/k8s"
	"github.com/jw-s/updatey/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	cert = flag.String("cert", "/certs/tls.crt", "path location to TLS certificate")
	key  = flag.String("key", "/certs/tls.key", "path location to TLS private key")
)

func init() {
	flag.Set("logtostderr", "true")
}

func main() {
	flag.Parse()

	cfg, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}

	wrapper := k8s.New(k8s.NewSecretRetriever(kubeClient.CoreV1()), version.NewSemVersionResolver(), &docker.Client{})

	server := &http.Server{
		Handler: admission.AdmitHandler(wrapper),
		Addr:    ":8080",
	}

	glog.Fatal(server.ListenAndServeTLS(*cert, *key))
}
