package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"

	"smartwallet/internal/service"
)

type UserHandler struct {
	userService    *service.UserService
	sessionService *service.SessionService
}

func NewUserHandler(userService *service.UserService,
	sessionService *service.SessionService,
) *UserHandler {
	return &UserHandler{
		userService:    userService,
		sessionService: sessionService,
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req *registerRequest

	if !decodeJSON(w, r, &req) {
		return
	}

	if req == nil {
		http.Error(w, "request must be a JSON object", http.StatusBadRequest)
		return
	}

	userID, err := h.userService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			http.Error(w, "invalid email or password", http.StatusBadRequest)

		case errors.Is(err, service.ErrEmailTaken):
			http.Error(w, "email is already registered", http.StatusConflict)

		default:
			log.Printf("registration failed: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"user_id": userID.String(),
	})
}

func writeJSONDecodeError(w http.ResponseWriter, err error) {
	var maxBytesErr *http.MaxBytesError

	if errors.As(err, &maxBytesErr) {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	http.Error(w, "invalid JSON", http.StatusBadRequest)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return false
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		writeJSONDecodeError(w, err)
		return false
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			writeJSONDecodeError(w, err)
		} else {
			http.Error(w, "request must contain exactly one JSON object", http.StatusBadRequest)
		}
		return false
	}

	return true
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req *registerRequest

	if !decodeJSON(w, r, &req) {
		return
	}

	if req == nil {
		http.Error(w, "request must be a JSON object", http.StatusBadRequest)
		return
	}

	userID, err := h.userService.Authenticate(
		r.Context(),
		req.Email,
		req.Password,
	)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(w, "invalid email or password", http.StatusUnauthorized)
			return
		}

		log.Printf("authentication failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	token, err := h.sessionService.Create(r.Context(), userID)
	if err != nil {
		log.Printf("create session: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // поменять на true на проде для https
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600, // 1 час
	})

	w.WriteHeader(http.StatusOK)

}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	email, err := h.userService.GetEmailById(
		r.Context(),
		userID,
	)
	if err != nil {
		if errors.Is(err, service.ErrIdNotExists) {
			http.Error(w, "invalid userid", http.StatusUnauthorized)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"user_id": userID.String(),
		"email":   email,
	})
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err == nil {
		if err := h.sessionService.RevokeToken(r.Context(), cookie.Value); err != nil {
			log.Printf("revoke session: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	} else if !errors.Is(err, http.ErrNoCookie) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusNoContent)
}
