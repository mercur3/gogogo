package handle

import (
	"errors"
	"fmt"
	"goweb/internal/api"
	"goweb/internal/common"
	"goweb/internal/middleware"
	"goweb/internal/otel"
	"goweb/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
	"go.opentelemetry.io/otel/attribute"
)

func MakeServer(author service.Author) *http.Server {
	v1 := http.NewServeMux()
	v1.HandleFunc("GET /test", getAllHandler)
	v1.HandleFunc("POST /test", func(w http.ResponseWriter, r *http.Request) {
		writeBody(w, http.StatusAccepted, "POST /v1/test")
	})
	v1.HandleFunc("GET /test/{id}", func(w http.ResponseWriter, r *http.Request) {
		val, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			writeBody(w, http.StatusOK, fmt.Sprintf("GET /v1/test/{%d}", val))
		}
	})

	v2 := http.NewServeMux()
	v2.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		a, err := author.GetAll(r.Context())
		if err != nil {
			setError(w, err)
		} else {
			writeBody(w, http.StatusOK, a)
		}
	})
	v2.HandleFunc("POST /test", func(w http.ResponseWriter, r *http.Request) {
		writeBody(w, http.StatusAccepted, "POST /v2/test")
	})
	v2.HandleFunc("GET /test/{id}", func(w http.ResponseWriter, r *http.Request) {
		tracer := otel.Tracer()
		ctx, span := tracer.Start(r.Context(), "GET /test/{id}")
		defer span.End()

		id, err := strconv.Atoi(r.PathValue("id"))
		if err == nil {
			a, err := author.Get(ctx, int64(id))
			if err != nil {
				setError(w, err)
			} else {
				writeBody(w, http.StatusOK, a)
			}
		} else {
			span.SetAttributes(attribute.String("error", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	mux := http.NewServeMux()
	mux.Handle("/v1/", middleware.DeprecatedEndpoint(http.StripPrefix("/v1", v1)))
	mux.Handle("/v2/", http.StripPrefix("/v2", v2))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeBody(w, http.StatusOK, map[string]string{"status": "alive"})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		writeBody(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	return &http.Server{
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      middleware.TraceRequest(mux),
	}
}

func getAllHandler(w http.ResponseWriter, r *http.Request) {
	writeBody(w, http.StatusOK, "GET /v1/test")
}

func MakeServerFromOpenAPI(a service.Author, b service.Book) *http.Server {
	server := api.NewServer(a, b)
	middlewares := []api.StrictMiddlewareFunc{
		// middleware.TraceRequestMiddleware,
		// middleware.MaxRequestBodyMiddleware,
	}
	i := api.NewStrictHandlerWithOptions(server, middlewares, api.StrictHTTPServerOptions{
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			requestID := r.Context().Value(middleware.RequestID).(uuid.UUID)

			if tErr, ok := errors.AsType[*common.TypedErr](err); ok {
				errMsg := api.ErrorMsg{
					Msg:       tErr.Msg,
					RequestId: requestID,
				}
				switch tErr.Kind {
				case common.ErrNotFound:
					writeBody(w, http.StatusNotFound, errMsg)
				case common.ErrAlreadyExists:
					writeBody(w, http.StatusBadRequest, errMsg)
				default:
					writeBody(w, http.StatusInternalServerError, errMsg)
				}
			} else {
				writeBody(w, http.StatusInternalServerError, api.ErrorMsg{
					Msg: err.Error(),
				})
			}
		},
	})

	r := http.NewServeMux()
	h := api.HandlerFromMux(i, r)
	// h := api.HandlerFromMuxWithBaseURL(i, r, "/api") // TODO add /api base point

	swagger, err := api.GetSpec()
	if err != nil {
		panic(err)
	}
	swagger.Servers = nil // clear servers so it doesn't validate the host
	h = nethttpmiddleware.OapiRequestValidator(swagger)(h)
	// these need to run over the std net/http or otherwise there is no instrumentation on the
	// request that fail due to the oapi validator middleware
	h = middleware.TraceRequest(h)
	h = http.MaxBytesHandler(h, 1<<20)

	return &http.Server{
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      h,
	}
}
