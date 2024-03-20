package shorturl

import (
	_ "embed"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/idna"

	"github.com/wzshiming/shorturl/storage"
)

type Handler struct {
	storage storage.Storage
	baseURL string
}

func WithBaseURL(baseURL string) func(*Handler) {
	return func(h *Handler) {
		h.baseURL = baseURL
	}
}

func NewHandler(storage storage.Storage, opts ...func(*Handler)) *Handler {
	h := &Handler{
		storage: storage,
	}
	for _, opt := range opts {
		opt(h)
	}

	return h
}

func errorResponse(w http.ResponseWriter, status int, msg ...string) {
	m := ""
	if len(msg) > 0 {
		m = strings.Join(msg, " ")
	} else {
		m = http.StatusText(status)
	}
	http.Error(w, m, status)
}

//go:embed page.html
var page string

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		h.get(w, r)
	case http.MethodPost:
		h.post(w, r)
	default:
		errorResponse(w, http.StatusMethodNotAllowed)
	}
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[1:]
	if code == "" {
		if r.Method == http.MethodGet {
			d, err := h.basePrefix(r)
			if err != nil {
				errorResponse(w, http.StatusBadRequest, "invalid domain")
				return
			}
			w.Write([]byte(fmt.Sprintf(page, d, d)))
		}
		return
	}
	if !Validate(code) {
		errorResponse(w, http.StatusBadRequest, "invalid code")
		return
	}
	url, err := h.storage.Decode(r.Context(), code)
	if err != nil {
		errorResponse(w, http.StatusNotFound)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) post(w http.ResponseWriter, r *http.Request) {
	u := r.FormValue("url")
	if len(u) == 0 {
		errorResponse(w, http.StatusBadRequest, "empty url")
		return
	}
	if len(u) > 2048 {
		errorResponse(w, http.StatusRequestURITooLong, "url too long")
		return
	}
	url, err := url.Parse(u)
	if err != nil || url.Scheme == "" || url.Host == "" {
		errorResponse(w, http.StatusBadRequest, "invalid url")
		return
	}
	host, err := idna.ToASCII(url.Host)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid host")
		return
	}
	url.Host = host
	url.ForceQuery = false
	if url.Path == "" {
		url.Path = "/"
	}
	code, err := h.storage.Encode(r.Context(), url.String())
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "storage error")
		return
	}
	d, err := h.basePrefix(r)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError)
		return
	}
	w.Write([]byte(d + code))
}

func (h *Handler) basePrefix(r *http.Request) (string, error) {
	if h.baseURL != "" {
		return h.baseURL, nil
	}
	u, err := r.URL.Parse("/")
	if err != nil {
		return "", err
	}
	u.Host = r.Host
	u.Scheme = "http"
	if r.TLS != nil {
		u.Scheme = "https"
	}
	return u.String(), nil
}
