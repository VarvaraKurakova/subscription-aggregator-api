package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/VarvaraKurakova/subscription-aggregator-api/internal/domain"
	"github.com/VarvaraKurakova/subscription-aggregator-api/internal/repository"
	"github.com/VarvaraKurakova/subscription-aggregator-api/internal/service"
)

type SubscriptionHandler struct {
	service *service.SubscriptionService
}

func NewSubscriptionHandler(service *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
	}
}

type CreateSubscriptionRequest struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
}

type UpdateSubscriptionRequest struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
}

type SubscriptionResponse struct {
	ID          int64   `json:"id"`
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type TotalResponse struct {
	Total       int     `json:"total"`
	Currency    string  `json:"currency"`
	PeriodFrom  string  `json:"period_from"`
	PeriodTo    string  `json:"period_to"`
	UserID      *string `json:"user_id,omitempty"`
	ServiceName *string `json:"service_name,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Create godoc
// @Summary Create subscription
// @Description Create a new user subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body CreateSubscriptionRequest true "Subscription data"
// @Success 201 {object} SubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/ [post]
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateSubscriptionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	sub, err := h.service.Create(r.Context(), service.CreateSubscriptionInput{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toSubscriptionResponse(sub))
}

// GetByID godoc
// @Summary Get subscription by ID
// @Description Get one subscription by numeric ID
// @Tags subscriptions
// @Produce json
// @Param id path int true "Subscription ID"
// @Success 200 {object} SubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription id")
		return
	}

	sub, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toSubscriptionResponse(sub))
}

// List godoc
// @Summary List subscriptions
// @Description Get subscriptions with optional filters and pagination
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "User UUID"
// @Param service_name query string false "Subscription service name"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} SubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/ [get]
func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	limit := parseIntQuery(query.Get("limit"), 20)
	offset := parseIntQuery(query.Get("offset"), 0)

	subs, err := h.service.List(r.Context(), service.ListSubscriptionsInput{
		UserID:      optionalString(query.Get("user_id")),
		ServiceName: optionalString(query.Get("service_name")),
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response := make([]SubscriptionResponse, 0, len(subs))
	for _, sub := range subs {
		response = append(response, toSubscriptionResponse(sub))
	}

	writeJSON(w, http.StatusOK, response)
}

// Update godoc
// @Summary Update subscription
// @Description Update an existing subscription by ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int true "Subscription ID"
// @Param request body UpdateSubscriptionRequest true "Subscription data"
// @Success 200 {object} SubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription id")
		return
	}

	var req UpdateSubscriptionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	sub, err := h.service.Update(r.Context(), service.UpdateSubscriptionInput{
		ID:          id,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toSubscriptionResponse(sub))
}

// Delete godoc
// @Summary Delete subscription
// @Description Delete subscription by ID
// @Tags subscriptions
// @Param id path int true "Subscription ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription id")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetTotal godoc
// @Summary Calculate total subscription cost
// @Description Calculate total subscription cost for selected period with optional filters by user_id and service_name
// @Tags subscriptions
// @Produce json
// @Param from query string true "Period start in MM-YYYY format"
// @Param to query string true "Period end in MM-YYYY format"
// @Param user_id query string false "User UUID"
// @Param service_name query string false "Subscription service name"
// @Success 200 {object} TotalResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/total [get]
func (h *SubscriptionHandler) GetTotal(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	from := query.Get("from")
	to := query.Get("to")

	if from == "" || to == "" {
		writeError(w, http.StatusBadRequest, "from and to are required")
		return
	}

	result, err := h.service.GetTotal(r.Context(), service.TotalSubscriptionsInput{
		From:        from,
		To:          to,
		UserID:      optionalString(query.Get("user_id")),
		ServiceName: optionalString(query.Get("service_name")),
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, TotalResponse{
		Total:       result.Total,
		Currency:    "RUB",
		PeriodFrom:  result.PeriodFrom,
		PeriodTo:    result.PeriodTo,
		UserID:      result.UserID,
		ServiceName: result.ServiceName,
	})
}

func parseIDParam(r *http.Request) (int64, error) {
	idParam := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return 0, err
	}

	if id <= 0 {
		return 0, strconv.ErrSyntax
	}

	return id, nil
}

func parseIntQuery(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

func toSubscriptionResponse(sub domain.Subscription) SubscriptionResponse {
	var endDate *string

	if sub.EndDate != nil {
		formattedEndDate := domain.FormatMonthYear(*sub.EndDate)
		endDate = &formattedEndDate
	}

	return SubscriptionResponse{
		ID:          sub.ID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID.String(),
		StartDate:   domain.FormatMonthYear(sub.StartDate),
		EndDate:     endDate,
		CreatedAt:   sub.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   sub.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, repository.ErrSubscriptionNotFound):
		writeError(w, http.StatusNotFound, "subscription not found")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, ErrorResponse{
		Error: message,
	})
}
