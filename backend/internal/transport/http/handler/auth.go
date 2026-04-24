package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	appauth "github.com/yuanjun5681/clawhire/backend/internal/application/auth"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/account"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/response"
)

type Auth struct {
	svc *appauth.Service
}

func NewAuth(svc *appauth.Service) *Auth {
	return &Auth{svc: svc}
}

type registerRequest struct {
	AccountID   string `json:"accountId"`
	DisplayName string `json:"displayName"`
	Password    string `json:"password"`
}

type loginRequest struct {
	AccountID string `json:"accountId"`
	Password  string `json:"password"`
}

type authResponse struct {
	Token     string          `json:"token"`
	ExpiresAt time.Time       `json:"expiresAt"`
	Account   accountResponse `json:"account"`
}

type accountResponse struct {
	AccountID   string         `json:"accountId"`
	Type        account.Type   `json:"type"`
	DisplayName string         `json:"displayName"`
	Status      account.Status `json:"status"`
}

func (h *Auth) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apierr.CodeInvalidRequest, "invalid JSON body")
		return
	}
	res, err := h.svc.Register(c.Request.Context(), appauth.RegisterHumanInput{
		AccountID:   req.AccountID,
		DisplayName: req.DisplayName,
		Password:    req.Password,
	})
	if err != nil {
		response.FailErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Success{Success: true, Data: toAuthResponse(res)})
}

func (h *Auth) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apierr.CodeInvalidRequest, "invalid JSON body")
		return
	}
	res, err := h.svc.Login(c.Request.Context(), req.AccountID, req.Password)
	if err != nil {
		response.FailErr(c, err)
		return
	}
	response.OK(c, toAuthResponse(res))
}

func toAuthResponse(res *appauth.AuthResult) authResponse {
	return authResponse{
		Token:     res.Token,
		ExpiresAt: res.ExpiresAt,
		Account: accountResponse{
			AccountID:   res.Account.AccountID,
			Type:        res.Account.Type,
			DisplayName: res.Account.DisplayName,
			Status:      res.Account.Status,
		},
	}
}
