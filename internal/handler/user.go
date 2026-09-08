package handler

import (
	"encoding/json"
	"net/http"
	"slices"

	errs "github.com/newaccg/todo-api/internal/errors"
)

func (h *handler) Register(w http.ResponseWriter, r *http.Request) error {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errs.ErrInvalidJSON
	}

	if err := validateInputStrings(input.Name, input.Email, input.Password); err != nil {
		return err
	}

	tokens, err := h.service.RegisterUser(r.Context(), input.Name, input.Email, input.Password)
	if err != nil {
		return err
	}

	writeJSON(w, tokens)

	return nil
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) error {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errs.ErrInvalidJSON
	}

	if err := validateInputStrings(input.Email, input.Password); err != nil {
		return err
	}

	tokens, err := h.service.LoginUser(r.Context(), input.Email, input.Password)
	if err != nil {
		return err
	}

	writeJSON(w, tokens)

	return nil
}

func (h *handler) Refresh(w http.ResponseWriter, r *http.Request) error {
	var input struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errs.ErrInvalidJSON
	}

	if input.RefreshToken == "" {
		return errs.ErrEmptyToken
	}

	ctx := r.Context()

	tokens, err := h.service.UpdateRefreshToken(ctx, input.RefreshToken)
	if err != nil {
		return err
	}

	writeJSON(w, tokens)

	return nil
}

// structural validation
func validateInputStrings(input ...string) error {
	if slices.Contains(input, "") {
		return errs.ErrEmptyField
	}

	return nil
}
