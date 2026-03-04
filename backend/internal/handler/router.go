package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/pathplanner-lab/backend/internal/middleware"
	"github.com/pathplanner-lab/backend/internal/repository"
	"github.com/pathplanner-lab/backend/internal/service"
)

type Deps struct {
	AuthSvc        *service.AuthService
	MapSvc         *service.MapService
	ExpSvc         *service.ExperimentService
	RunSvc         *service.RunService
	CompareSvc     *service.CompareService
	LLMSvc         *service.LLMService
	IdempRepo      repository.IdempotencyRepository
	JWTSecret    string
	Auth0KeyFunc *middleware.Auth0KeyFunc
	UpsertUser   middleware.UpsertUserFunc
}

func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "Idempotency-Key"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	authHandler := NewAuthHandler(deps.AuthSvc)
	mapHandler := NewMapHandler(deps.MapSvc)
	expHandler := NewExperimentHandler(deps.ExpSvc, deps.IdempRepo)
	runHandler := NewRunHandler(deps.RunSvc)
	compareHandler := NewCompareHandler(deps.CompareSvc, deps.LLMSvc)

	authMW := middleware.NewAuthMiddleware(deps.JWTSecret, deps.Auth0KeyFunc, deps.UpsertUser)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/v1", func(r chi.Router) {
		// Auth
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
		r.With(authMW).Get("/me", authHandler.Me)

		// Maps
		r.With(authMW).Route("/maps", func(r chi.Router) {
			r.Post("/", mapHandler.Create)
			r.Get("/", mapHandler.List)
			r.Get("/{mapId}", mapHandler.Get)
			r.Put("/{mapId}", mapHandler.Update)
			r.Delete("/{mapId}", mapHandler.Delete)
		})

		// Experiments
		r.With(authMW).Route("/experiments", func(r chi.Router) {
			r.Post("/", expHandler.Create)
			r.Get("/", expHandler.List)
			r.Get("/{experimentId}", expHandler.Get)
			r.Get("/{experimentId}/runs", expHandler.GetRuns)
			r.Post("/{experimentId}/cancel", expHandler.Cancel)
		})

		// Runs
		r.With(authMW).Route("/runs", func(r chi.Router) {
			r.Get("/{runId}", runHandler.Get)
			r.Get("/{runId}/metrics", runHandler.GetMetrics)
			r.Get("/{runId}/artifacts", runHandler.GetArtifacts)
			if deps.LLMSvc != nil {
				r.Post("/{runId}/llm-review", func(w http.ResponseWriter, req *http.Request) {
					userID := middleware.GetUserID(req.Context())
					runID := chi.URLParam(req, "runId")
					var body struct {
						Mode          string `json:"mode"`
						PromptVersion string `json:"promptVersion"`
					}
					body.Mode = "teacher"
					body.PromptVersion = "v1"
					_ = decodeJSON(req, &body)
					report, err := deps.LLMSvc.ReviewRun(req.Context(), runID, userID, body.Mode, body.PromptVersion)
					if err != nil {
						handleServiceError(w, err)
						return
					}
					writeJSON(w, http.StatusOK, map[string]interface{}{
						"reportId":        report.ID,
						"summary":         report.Summary,
						"recommendations": report.Recommendations,
					})
				})
			}
		})

		// Compare
		r.Post("/compare", compareHandler.Compare)
		r.With(authMW).Post("/compare/llm-review", compareHandler.LLMReview)
	})

	return r
}
