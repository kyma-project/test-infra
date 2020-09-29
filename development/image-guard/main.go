package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	stdlog "log"
	"net/http"
	"os"
	"strings"
	"time"

	admissioncontrol "github.com/elithrar/admission-control"
	log "github.com/go-kit/kit/log"
)

type conf struct {
	TLSCertPath string
	TLSKeyPath  string
	HTTPOnly    bool
	Port        string
	Host        string
}

func main() {
	ctx := context.Background()

	// Get config
	conf := &conf{}
	flag.StringVar(&conf.TLSCertPath, "cert-path", "./cert.crt", "The path to the PEM-encoded TLS certificate")
	flag.StringVar(&conf.TLSKeyPath, "key-path", "./key.pem", "The path to the unencrypted TLS key")
	flag.BoolVar(&conf.HTTPOnly, "http-only", false, "Only listen on unencrypted HTTP (e.g. for proxied environments)")
	flag.StringVar(&conf.Port, "port", "8443", "The port to listen on (HTTPS).")
	flag.StringVar(&conf.Host, "host", "admissiond.questionable.services", "The hostname for the service")
	flag.Parse()

	// Set up logging
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	stdlog.SetOutput(log.NewStdlibAdapter(logger))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "loc", log.DefaultCaller)

	// TLS configuration
	// Only load the TLS keypair if the -http-only flag is not set.
	var tlsConf *tls.Config
	if !conf.HTTPOnly {
		keyPair, err := tls.LoadX509KeyPair(conf.TLSCertPath, conf.TLSKeyPath)
		if err != nil {
			fatal(logger, err)
		}
		tlsConf = &tls.Config{
			Certificates: []tls.Certificate{keyPair},
			ServerName:   conf.Host,
		}
	}

	// Set up the routes & logging middleware.
	r := mux.NewRouter().StrictSlash(true)
	// Show all available routes
	msg := "Image Guard - Admission Control server"
	r.Handle("/", printAvailableRoutes(r, logger, msg)).Methods(http.MethodGet)
	// Default health-check endpoint
	r.HandleFunc("/healthz", healthCheckHandler).Methods(http.MethodGet)

	// Admission control endpoints
	admissions := r.PathPrefix("/admission-control").Subrouter()
	//admissions.Handle("/enforce-image-registry", &admissioncontrol.AdmissionHandler{
	//	AdmitFunc: enforceImageRegistries("gcr.io/kyma-project", "eu.gcr.io/kyma-project"),
	//	Logger:    logger,
	//})
	admissions.Handle("/collect-used-images", &admissioncontrol.AdmissionHandler{
		AdmitFunc: collectUsedImages(),
		Logger:    logger,
	})

	// HTTP server
	timeout := time.Second * 15
	srv := &http.Server{
		Handler:           admissioncontrol.LoggingMiddleware(logger)(r),
		TLSConfig:         tlsConf,
		Addr:              ":" + conf.Port,
		IdleTimeout:       timeout,
		ReadTimeout:       timeout,
		ReadHeaderTimeout: timeout,
		WriteTimeout:      timeout,
	}

	admissionServer, err := admissioncontrol.NewServer(
		srv,
		log.With(logger, "component", "server"),
	)
	if err != nil {
		fatal(logger, err)
		return
	}

	if err := admissionServer.Run(ctx); err != nil {
		fatal(logger, err)
		return
	}
}

func fatal(logger log.Logger, err error) {
	logger.Log(
		"status", "fatal",
		"err", err,
	)

	os.Exit(1)
	return
}

// healthCheckHandler returns a HTTP 200, everytime.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// printAvailableRoutes prints all routes attached to the provided Router, and
// prepends a message to the response.
func printAvailableRoutes(router *mux.Router, logger log.Logger, msg string) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		var routes []string
		err := router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			path, err := route.GetPathTemplate()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logger.Log("msg", "walkFunc failed", err, err.Error())
				return err
			}

			routes = append(routes, path)
			return nil
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Log("msg", "walkFunc failed", err, err.Error())
			return
		}

		fmt.Fprintln(w, msg)
		fmt.Fprintln(w, "Available routes:")
		for _, path := range routes {
			fmt.Fprintln(w, path)
		}
	}

	return http.HandlerFunc(fn)
}

func enforceImageRegistries(registries ...string) admissioncontrol.AdmitFunc {
	return func(reviewRequest *v1beta1.AdmissionReview) (*v1beta1.AdmissionResponse, error) {
		kind := reviewRequest.Request.Kind.Kind
		resp := &v1beta1.AdmissionResponse{Allowed: false, Result: &metav1.Status{}}
		if kind != "Pod" {
			resp.Allowed = true
			resp.Result.Message = fmt.Sprintf("Got non-Pod type (%s)", kind)
			return resp, nil
		}
		pod := v1.Pod{}
		deserializer := serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
		if _, _, err := deserializer.Decode(reviewRequest.Request.Object.Raw, nil, &pod); err != nil {
			return nil, err
		}
		for _, c := range pod.Spec.Containers {
			for _, r := range registries {
				if strings.HasPrefix(c.Image, r) {
					resp.Allowed = true
					resp.Result.Message = fmt.Sprintf("trusted registry validated for image (%s)", c.Image)
					return resp, nil
				}
			}
		}
		resp.Result.Message = fmt.Sprintf("non-trusted registry was defined for one of the containers. available registries: %s", registries)
		return resp, nil
	}
}

func collectUsedImages() admissioncontrol.AdmitFunc {
	return func(reviewRequest *v1beta1.AdmissionReview) (*v1beta1.AdmissionResponse, error) {
		kind := reviewRequest.Request.Kind.Kind
		resp := &v1beta1.AdmissionResponse{Allowed: true, Result: &metav1.Status{}}
		if kind == "Pod" {
			pod := v1.Pod{}
			deserializer := serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
			if _, _, err := deserializer.Decode(reviewRequest.Request.Object.Raw, nil, &pod); err != nil {
				return nil, err
			}
			for _, c := range pod.Spec.Containers {
				fmt.Printf("image: %s\n", c.Image)
			}
		}
		return resp, nil
	}
}
