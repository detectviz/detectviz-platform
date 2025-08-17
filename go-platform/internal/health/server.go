package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/detectviz/detectviz-platform/go-platform/internal/contracts"
	"go.uber.org/zap"
)

// Server 提供 liveness/readiness 健康檢查
type Server struct {
	httpSrv   *http.Server
	readyFlag atomic.Bool
}

func NewServer(addr string) *Server {
	mux := http.NewServeMux()
	s := &Server{httpSrv: &http.Server{Addr: addr, Handler: mux}}

	// liveness: 只要進程活著即 200
	mux.HandleFunc("/livez", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// readiness: 依據 readyFlag
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if s.readyFlag.Load() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ready"))
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("not ready"))
	})

	// contract info endpoint for debugging and validation
	mux.HandleFunc("/contracts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		contractInfo := contracts.GetContractInfo()

		if contractInfo["status"] == "error" {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		if err := json.NewEncoder(w).Encode(contractInfo); err != nil {
			zap.L().Error("Failed to encode contract info", zap.Error(err))
		}
	})

	return s
}

func (s *Server) SetReady(v bool) { s.readyFlag.Store(v) }

func (s *Server) Start() {
	go func() {
		zap.L().Info("Health server listening", zap.String("addr", s.httpSrv.Addr))
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Error("Health server error", zap.Error(err))
		}
	}()
}

func (s *Server) Stop(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := s.httpSrv.Shutdown(ctx); err != nil {
		zap.L().Warn("Health server shutdown error", zap.Error(err))
	}
}
