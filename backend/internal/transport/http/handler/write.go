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

type awardTaskRequest struct {
	ContractID   string          `json:"contractId"`
	ExecutorID   string          `json:"executorId"`
	AgreedReward clawhire.Reward `json:"agreedReward"`
}

type createSubmissionRequest struct {
	SubmissionID string             `json:"submissionId"`
	ContractID   string             `json:"contractId,omitempty"`
	Artifacts    []shared.Artifact  `json:"artifacts"`
	Summary      string             `json:"summary"`
	Evidence     *clawhire.Evidence `json:"evidence,omitempty"`
}

type acceptSubmissionRequest struct {
	SubmissionID string     `json:"submissionId"`
	AcceptedAt   *time.Time `json:"acceptedAt,omitempty"`
}

type rejectSubmissionRequest struct {
	SubmissionID string     `json:"submissionId"`
	Reason       string     `json:"reason"`
	RejectedAt   *time.Time `json:"rejectedAt,omitempty"`
}

type awardTaskResponse struct {
	TaskID     string `json:"taskId"`
	ContractID string `json:"contractId"`
	EventID    string `json:"eventId,omitempty"`
}

type submissionResponse struct {
	TaskID       string `json:"taskId"`
	SubmissionID string `json:"submissionId"`
	EventID      string `json:"eventId,omitempty"`
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
		reviewer, err = h.loadActor(c, rid, "reviewer")
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

func (h *Write) AwardTask(c *gin.Context) {
	var req awardTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apierr.CodeInvalidRequest, "invalid JSON body")
		return
	}
	if _, err := h.currentHumanActor(c); err != nil {
		response.FailErr(c, err)
		return
	}

	executor, err := h.loadActor(c, strings.TrimSpace(req.ExecutorID), "executor")
	if err != nil {
		response.FailErr(c, err)
		return
	}

	taskID := strings.TrimSpace(c.Param("taskId"))
	event := h.httpEvent(c, "clawhire.task.awarded", fmt.Sprintf("%s:%s", taskID, req.ContractID))
	if err := h.commands.AwardTask(c.Request.Context(), appcmd.AwardTaskCommand{
		Payload: clawhire.AwardTaskPayload{
			TaskID:       taskID,
			ContractID:   strings.TrimSpace(req.ContractID),
			Executor:     *executor,
			AgreedReward: req.AgreedReward,
		},
		Event: event,
	}); err != nil {
		response.FailErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Success{Success: true, Data: awardTaskResponse{
		TaskID:     taskID,
		ContractID: strings.TrimSpace(req.ContractID),
		EventID:    eventID(event),
	}})
}

func (h *Write) CreateSubmission(c *gin.Context) {
	var req createSubmissionRequest
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
	event := h.httpEvent(c, "clawhire.submission.created", fmt.Sprintf("%s:%s", taskID, req.SubmissionID))
	if err := h.commands.CreateSubmission(c.Request.Context(), appcmd.CreateSubmissionCommand{
		Payload: clawhire.CreateSubmissionPayload{
			TaskID:       taskID,
			SubmissionID: strings.TrimSpace(req.SubmissionID),
			ContractID:   strings.TrimSpace(req.ContractID),
			Executor:     actor,
			Artifacts:    req.Artifacts,
			Summary:      req.Summary,
			Evidence:     req.Evidence,
		},
		Event: event,
	}); err != nil {
		response.FailErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Success{Success: true, Data: submissionResponse{
		TaskID:       taskID,
		SubmissionID: strings.TrimSpace(req.SubmissionID),
		EventID:      eventID(event),
	}})
}

func (h *Write) AcceptSubmission(c *gin.Context) {
	var req acceptSubmissionRequest
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
	event := h.httpEvent(c, "clawhire.submission.accepted", fmt.Sprintf("%s:%s", taskID, req.SubmissionID))
	if err := h.commands.AcceptSubmission(c.Request.Context(), appcmd.AcceptSubmissionCommand{
		Payload: clawhire.AcceptSubmissionPayload{
			TaskID:       taskID,
			SubmissionID: strings.TrimSpace(req.SubmissionID),
			AcceptedBy:   actor,
			AcceptedAt:   req.AcceptedAt,
		},
		Event: event,
	}); err != nil {
		response.FailErr(c, err)
		return
	}
	response.OK(c, submissionResponse{
		TaskID:       taskID,
		SubmissionID: strings.TrimSpace(req.SubmissionID),
		EventID:      eventID(event),
	})
}

func (h *Write) RejectSubmission(c *gin.Context) {
	var req rejectSubmissionRequest
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
	event := h.httpEvent(c, "clawhire.submission.rejected", fmt.Sprintf("%s:%s", taskID, req.SubmissionID))
	if err := h.commands.RejectSubmission(c.Request.Context(), appcmd.RejectSubmissionCommand{
		Payload: clawhire.RejectSubmissionPayload{
			TaskID:       taskID,
			SubmissionID: strings.TrimSpace(req.SubmissionID),
			RejectedBy:   actor,
			Reason:       req.Reason,
			RejectedAt:   req.RejectedAt,
		},
		Event: event,
	}); err != nil {
		response.FailErr(c, err)
		return
	}
	response.OK(c, submissionResponse{
		TaskID:       taskID,
		SubmissionID: strings.TrimSpace(req.SubmissionID),
		EventID:      eventID(event),
	})
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

func (h *Write) loadActor(c *gin.Context, accountID string, field string) (*shared.Actor, error) {
	acc, err := h.accounts.FindByID(c.Request.Context(), accountID)
	if err != nil {
		if err == account.ErrAccountNotFound {
			return nil, apierr.Wrap(apierr.CodeNotFound, fmt.Sprintf("find %s account", field), err)
		}
		return nil, apierr.Wrap(apierr.CodeInternalError, fmt.Sprintf("find %s account", field), err)
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
		return nil, apierr.New(apierr.CodeInvalidRequest, fmt.Sprintf("%s account type is invalid", field))
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

func eventID(meta *appcmd.EventMeta) string {
	if meta == nil {
		return ""
	}
	return meta.ID
}
