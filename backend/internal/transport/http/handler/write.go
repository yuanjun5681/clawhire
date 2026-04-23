package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	appcmd "github.com/yuanjun5681/clawhire/backend/internal/application/command"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/account"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawhire"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/response"
	"github.com/yuanjun5681/clawhire/backend/internal/transport/http/middleware"
)

const headerAccountID = "X-Account-ID"

type Write struct {
	commands *appcmd.Service
	accounts account.Repository
}

func NewWrite(commands *appcmd.Service, accounts account.Repository) *Write {
	return &Write{commands: commands, accounts: accounts}
}

type createTaskRequest struct {
	TaskID          string                    `json:"taskId"`
	ReviewerID      string                    `json:"reviewerId,omitempty"`
	Title           string                    `json:"title"`
	Description     string                    `json:"description,omitempty"`
	Category        string                    `json:"category"`
	Reward          clawhire.Reward           `json:"reward"`
	AcceptanceSpec  *clawhire.AcceptanceSpec  `json:"acceptanceSpec,omitempty"`
	SettlementTerms *clawhire.SettlementTerms `json:"settlementTerms,omitempty"`
	Deadline        *time.Time                `json:"deadline,omitempty"`
}

type createBidRequest struct {
	BidID    string  `json:"bidId"`
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
	Proposal string  `json:"proposal,omitempty"`
}

func (h *Write) CreateTask(c *gin.Context) {
	var req createTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apierr.CodeInvalidRequest, "invalid JSON body")
		return
	}
	actor, err := h.currentHumanActor(c)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	var reviewer *shared.Actor
	if rid := strings.TrimSpace(req.ReviewerID); rid != "" {
		reviewer, err = h.loadReviewer(c, rid)
		if err != nil {
			response.FailErr(c, err)
			return
		}
	}

	var deadline *time.Time
	if req.Deadline != nil {
		t := req.Deadline.UTC()
		deadline = &t
	}

	res, err := h.commands.PostTask(c.Request.Context(), appcmd.PostTaskCommand{
		Payload: clawhire.PostTaskPayload{
			TaskID:          strings.TrimSpace(req.TaskID),
			Requester:       actor,
			Reviewer:        reviewer,
			Title:           req.Title,
			Description:     req.Description,
			Category:        req.Category,
			Reward:          req.Reward,
			AcceptanceSpec:  req.AcceptanceSpec,
			SettlementTerms: req.SettlementTerms,
			Deadline:        deadline,
		},
		Event: h.httpEvent(c, "clawhire.task.posted", req.TaskID),
	})
	if err != nil {
		response.FailErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Success{Success: true, Data: res})
}

func (h *Write) CreateBid(c *gin.Context) {
	var req createBidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apierr.CodeInvalidRequest, "invalid JSON body")
		return
	}
	actor, err := h.currentHumanActor(c)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	taskID := strings.TrimSpace(c.Param("taskId"))
	res, err := h.commands.PlaceBid(c.Request.Context(), appcmd.PlaceBidCommand{
		Payload: clawhire.PlaceBidPayload{
			TaskID:   taskID,
			BidID:    strings.TrimSpace(req.BidID),
			Executor: actor,
			Price:    req.Price,
			Currency: req.Currency,
			Proposal: req.Proposal,
		},
		Event: h.httpEvent(c, "clawhire.bid.placed", fmt.Sprintf("%s:%s", taskID, req.BidID)),
	})
	if err != nil {
		response.FailErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Success{Success: true, Data: res})
}

func (h *Write) currentHumanActor(c *gin.Context) (shared.Actor, error) {
	accountID := strings.TrimSpace(c.GetHeader(headerAccountID))
	if accountID == "" {
		return shared.Actor{}, apierr.New(apierr.CodeInvalidRequest, "missing X-Account-ID header")
	}
	acc, err := h.accounts.FindByID(c.Request.Context(), accountID)
	if err != nil {
		if err == account.ErrAccountNotFound {
			return shared.Actor{}, apierr.Wrap(apierr.CodeNotFound, "find current account", err)
		}
		return shared.Actor{}, apierr.Wrap(apierr.CodeInternalError, "find current account", err)
	}
	if acc.Type != account.TypeHuman {
		return shared.Actor{}, apierr.New(apierr.CodeForbidden, "current account is not a human account")
	}
	if acc.Status != account.StatusActive {
		return shared.Actor{}, apierr.New(apierr.CodeForbidden, "current account is not active")
	}
	return shared.Actor{
		ID:   acc.AccountID,
		Kind: shared.ActorKindUser,
		Name: acc.DisplayName,
	}, nil
}

func (h *Write) loadReviewer(c *gin.Context, reviewerID string) (*shared.Actor, error) {
	acc, err := h.accounts.FindByID(c.Request.Context(), reviewerID)
	if err != nil {
		if err == account.ErrAccountNotFound {
			return nil, apierr.Wrap(apierr.CodeNotFound, "find reviewer account", err)
		}
		return nil, apierr.Wrap(apierr.CodeInternalError, "find reviewer account", err)
	}
	actor := shared.Actor{
		ID:   acc.AccountID,
		Name: acc.DisplayName,
	}
	switch acc.Type {
	case account.TypeHuman:
		actor.Kind = shared.ActorKindUser
	case account.TypeAgent:
		actor.Kind = shared.ActorKindAgent
	default:
		return nil, apierr.New(apierr.CodeInvalidRequest, "reviewer account type is invalid")
	}
	return &actor, nil
}

func (h *Write) httpEvent(c *gin.Context, eventType, suffix string) *appcmd.EventMeta {
	rid, _ := c.Get(middleware.CtxKeyRequestID)
	requestID, _ := rid.(string)
	requestID = strings.TrimSpace(requestID)
	suffix = strings.TrimSpace(suffix)
	if requestID == "" || suffix == "" {
		return nil
	}
	return &appcmd.EventMeta{
		ID:   fmt.Sprintf("http:%s:%s:%s", eventType, requestID, suffix),
		Type: eventType,
	}
}
