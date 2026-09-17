package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Métricas expostas no padrão Prometheus.
var (
	// requestsTotal mede o VOLUME de requisições, quebrado por método, rota e status HTTP.
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests received by http-server-projeto-korp.",
		},
		[]string{"method", "path", "status"},
	)

	// requestDuration ajuda a analisar latência/comportamento do serviço (bônus).
	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

// respostaProjetoKorp representa o corpo JSON retornado pelo endpoint /projeto-korp.
type respostaProjetoKorp struct {
	Nome    string `json:"nome"`
	Horario string `json:"horario"`
}

// statusRecorder embrulha o ResponseWriter para capturar o status code retornado.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// metricsMiddleware instrumenta cada requisição: conta o volume e mede a duração.
func metricsMiddleware(path string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next(rec, r)

		duracao := time.Since(inicio).Seconds()
		requestDuration.WithLabelValues(r.Method, path).Observe(duracao)
		requestsTotal.WithLabelValues(r.Method, path, strconv.Itoa(rec.status)).Inc()
	}
}

// projetoKorpHandler responde o JSON com o nome do projeto e o horário atual em UTC,
// resolvido dinamicamente a cada requisição.
func projetoKorpHandler(w http.ResponseWriter, r *http.Request) {
	resposta := respostaProjetoKorp{
		Nome:    "Projeto Korp",
		Horario: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(resposta); err != nil {
		http.Error(w, "erro ao gerar resposta", http.StatusInternalServerError)
		return
	}
}

// healthHandler expõe a disponibilidade do serviço (endpoint dedicado de health check).
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// metricsHandler expõe as métricas sempre no formato texto clássico do Prometheus
// (text/plain; version=0.0.4). Forçamos esse formato para evitar a negociação
// OpenMetrics, garantindo que todas as famílias de métricas sejam coletadas.
func metricsHandler() http.Handler {
	h := promhttp.Handler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Accept", "text/plain; version=0.0.4; charset=utf-8")
		h.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /projeto-korp", metricsMiddleware("/projeto-korp", projetoKorpHandler))
	mux.HandleFunc("GET /health", metricsMiddleware("/health", healthHandler))

	// Endpoint padrão do Prometheus para coleta das métricas.
	mux.Handle("GET /metrics", metricsHandler())

	servidor := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("http-server-projeto-korp escutando na porta 8080")
	if err := servidor.ListenAndServe(); err != nil {
		log.Fatalf("erro ao iniciar o servidor: %v", err)
	}
}
