package user

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"start/internal/auth"
	"start/internal/models"
	"start/internal/response"

	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	store  *Store
	secret string
}

func NewHandler(st *Store, s string) *Handler {
	return &Handler{store: st, secret: s}
}

type SafeUser struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type LoginUser struct {
	Email        string `json:"email"`
	PasswordHash string `json:"passwordHash"`
}

// InsertUser creates a new user
// @Summary Register a new user
// @Description Create a new user account with the provided details
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.User true "User registration data"
// @Success 201 {object} SafeUser "Created user"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload or validation error"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /admin/users [post]
func (h *Handler) InsertUser(w http.ResponseWriter, r *http.Request) {
	var baseUser models.User
	err := json.NewDecoder(r.Body).Decode(&baseUser)
	if err != nil {
		slog.Error("Bad request on payload", "error", err, "path", r.URL.Path)
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	errValidate := baseUser.Validate()
	if errValidate != nil {
		slog.Error("Bad request on payload", "error", errValidate, "path", r.URL.Path)
		response.Error(w, http.StatusBadRequest, errValidate.Error())
		return
	}

	hashPassword, errHash := bcrypt.GenerateFromPassword([]byte(baseUser.PasswordHash), 10)
	if errHash != nil {
		slog.Error("Error while hashing the password", "error", errHash)
		response.Error(w, http.StatusInternalServerError, "An unexpected error ocurred")
		return
	}

	user := models.User{
		ID:           baseUser.ID,
		Name:         baseUser.Name,
		Email:        baseUser.Email,
		PasswordHash: string(hashPassword),
		IsActive:     baseUser.IsActive,
		Role:         baseUser.Role,
	}

	createdUser, err := h.store.CreateUser(user)
	if err != nil {
		slog.Error("Error creating the user", "error", err)
		response.Error(w, http.StatusInternalServerError, "An unexpected error ocurred")
		return
	}

	safeUser := SafeUser{
		ID:    createdUser.ID,
		Name:  createdUser.Name,
		Email: createdUser.Email,
		Role:  createdUser.Role,
	}

	response.Respond(w, http.StatusCreated, safeUser)
}

// LoginUser authenticates a user and returns a JWT token
// @Summary Authenticate user
// @Description Login with email and password to receive an authentication token
// @Tags users
// @Accept json
// @Produce json
// @Param request body LoginUser true "Login credentials"
// @Success 200 {object} response.Response "Logged in successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Invalid email or password"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /public/login [post]
func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var loginUser LoginUser
	err := json.NewDecoder(r.Body).Decode(&loginUser)
	if err != nil {
		slog.Error("Bad request on payload", "error", err, "path", r.URL.Path)
		response.Error(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, err := h.store.GetUserByEmail(loginUser.Email)
	if err != nil {
		slog.Error("Get user by email failed", "error", err, "email", loginUser.Email)
		response.Error(w, http.StatusUnauthorized, "Failed to login")
		return
	}

	errCompare := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginUser.PasswordHash))
	if errCompare != nil {
		slog.Error("Failed to compare hashed password", "error", err, "user", loginUser.Email)
		response.Error(w, http.StatusUnauthorized, "Failed to login. Check the email or the password")
		return
	}

	tokenString, tokenExp, errToken := auth.GenerateToken(user.ID, user.Role, h.secret)
	if errToken != nil {
		slog.Error("Error generating JWT token", "error", errToken)
		response.Error(w, http.StatusInternalServerError, "An unexpected error ocurred")
		return
	}

	httpCookie := http.Cookie{
		Name:     "cinema_auth_token",
		Value:    tokenString,
		Expires:  time.Now().Add(tokenExp),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}

	http.SetCookie(w, &httpCookie)

	slog.Info("User logged in", "email", user.Email)
	response.Respond(w, http.StatusOK, "Logged In")
}
