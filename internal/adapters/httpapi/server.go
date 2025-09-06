package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/oziev02/subscriptions-service/internal/pkg/config"
	"github.com/oziev02/subscriptions-service/internal/usecase"
)

type Server struct {
	cfg *config.Config
	log *zap.Logger
	uc  *usecase.Service
}

func NewServer(cfg *config.Config, log *zap.Logger, repo usecase.SubscriptionRepo) *Server {
	return &Server{cfg: cfg, log: log, uc: usecase.NewService(repo)}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer, middleware.Timeout(60e9))
	r.Get("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/v1/subscriptions", func(r chi.Router) {
		r.Get("/", s.list)
		r.Post("/", s.create)
		r.Get("/summary", s.summary)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", s.get)
			r.Put("/", s.update)
			r.Delete("/", s.delete)
		})
	})
	// serve swagger spec
	r.Get("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		http.ServeFile(w, r, "./docs/openapi.yaml")
	})
	return r
}

type createReq struct {
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date,omitempty"`
}

func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	var req createReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	out, err := s.uc.Create(r.Context(), usecase.CreateInput{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, toDTO(out))
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	res, err := s.uc.Get(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, toDTO(res))
}

type updateReq struct {
	ServiceName *string `json:"service_name,omitempty"`
	Price       *int    `json:"price,omitempty"`
	StartDate   *string `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date"` // may be null or ""
}

func (s *Server) update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	var req updateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	res, err := s.uc.Update(r.Context(), id, usecase.UpdateInput{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		EndDateSet:  true,
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, toDTO(res))
}

func (s *Server) delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if err := s.uc.Delete(r.Context(), id); err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	var f usecase.ListFilter
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		id, err := uuid.Parse(uid)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		f.UserID = &id
	}
	if sn := r.URL.Query().Get("service_name"); sn != "" {
		f.ServiceName = &sn
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			f.Limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil {
			f.Offset = n
		}
	}
	res, err := s.uc.List(r.Context(), f)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	items := make([]any, 0, len(res))
	for _, s := range res {
		items = append(items, toDTO(s))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) summary(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		writeErr(w, http.StatusBadRequest, errors.New("from/to are required"))
		return
	}
	var uid *uuid.UUID
	if q := r.URL.Query().Get("user_id"); q != "" {
		id, err := uuid.Parse(q)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		uid = &id
	}
	var sn *string
	if q := r.URL.Query().Get("service_name"); q != "" {
		sn = &q
	}
	sum, err := s.uc.Summary(r.Context(), from, to, uid, sn)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"total": sum})
}
