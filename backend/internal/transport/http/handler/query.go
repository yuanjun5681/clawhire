package handler

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/account"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/bid"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/milestone"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/progress"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/review"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/settlement"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/submission"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/response"
)

type Query struct {
	tasks       task.Repository
	bids        bid.Repository
	progress    progress.Repository
	milestones  milestone.Repository
	submissions submission.Repository
	reviews     review.Repository
	settlements settlement.Repository
	accounts    account.Repository
}

func NewQuery(
	tasks task.Repository,
	bids bid.Repository,
	progress progress.Repository,
	milestones milestone.Repository,
	submissions submission.Repository,
	reviews review.Repository,
	settlements settlement.Repository,
	accounts account.Repository,
) *Query {
	return &Query{
		tasks:       tasks,
		bids:        bids,
		progress:    progress,
		milestones:  milestones,
		submissions: submissions,
		reviews:     reviews,
		settlements: settlements,
		accounts:    accounts,
	}
}

type taskListItem struct {
	TaskID         string      `json:"taskId"`
	Title          string      `json:"title"`
	Category       string      `json:"category"`
	Status         task.Status `json:"status"`
	Requester      interface{} `json:"requester"`
	Reward         task.Reward `json:"reward"`
	Deadline       *time.Time  `json:"deadline,omitempty"`
	LastActivityAt *time.Time  `json:"lastActivityAt,omitempty"`
}

type executorHistoryItem struct {
	TaskID     string      `json:"taskId"`
	Title      string      `json:"title"`
	Category   string      `json:"category"`
	Status     task.Status `json:"status"`
	Reward     task.Reward `json:"reward"`
	AcceptedAt *time.Time  `json:"acceptedAt,omitempty"`
	SettledAt  *time.Time  `json:"settledAt,omitempty"`
}

func (h *Query) ListTasks(c *gin.Context) {
	filter, page, pageSize, err := parseTaskFilter(c)
	if err != nil {
		response.FailErr(c, err)
		return
	}
	items, total, err := h.tasks.List(c.Request.Context(), filter)
	if err != nil {
		response.FailErr(c, apierr.Wrap(apierr.CodeInternalError, "list tasks", err))
		return
	}
	out := make([]taskListItem, 0, len(items))
	for _, item := range items {
		out = append(out, taskListItem{
			TaskID:         item.TaskID,
			Title:          item.Title,
			Category:       item.Category,
			Status:         item.Status,
			Requester:      item.Requester,
			Reward:         item.Reward,
			Deadline:       item.Deadline,
			LastActivityAt: item.LastActivityAt,
		})
	}
	response.OKMeta(c, out, &response.Meta{Page: page, PageSize: pageSize, Total: total})
}

func (h *Query) GetTask(c *gin.Context) {
	item, err := h.tasks.FindByID(c.Request.Context(), c.Param("taskId"))
	if err != nil {
		response.FailErr(c, repoToHTTPError("get task", err))
		return
	}
	response.OK(c, item)
}

func (h *Query) ListTaskBids(c *gin.Context) {
	page, pageSize, err := parsePage(c)
	if err != nil {
		response.FailErr(c, err)
		return
	}
	items, total, err := h.bids.ListByTask(c.Request.Context(), c.Param("taskId"), page, pageSize)
	if err != nil {
		response.FailErr(c, apierr.Wrap(apierr.CodeInternalError, "list task bids", err))
		return
	}
	response.OKMeta(c, items, &response.Meta{Page: page, PageSize: pageSize, Total: total})
}

func (h *Query) ListTaskProgress(c *gin.Context) {
	page, pageSize, err := parsePage(c)
	if err != nil {
		response.FailErr(c, err)
		return
	}
	items, total, err := h.progress.ListByTask(c.Request.Context(), c.Param("taskId"), page, pageSize)
	if err != nil {
		response.FailErr(c, apierr.Wrap(apierr.CodeInternalError, "list task progress", err))
		return
	}
	response.OKMeta(c, items, &response.Meta{Page: page, PageSize: pageSize, Total: total})
}

func (h *Query) ListTaskMilestones(c *gin.Context) {
	items, err := h.milestones.ListByTask(c.Request.Context(), c.Param("taskId"))
	if err != nil {
		response.FailErr(c, apierr.Wrap(apierr.CodeInternalError, "list task milestones", err))
		return
	}
	response.OK(c, items)
}

func (h *Query) ListTaskSubmissions(c *gin.Context) {
	page, pageSize, err := parsePage(c)
	if err != nil {
		response.FailErr(c, err)
		return
	}
	items, total, err := h.submissions.ListByTask(c.Request.Context(), c.Param("taskId"), page, pageSize)
	if err != nil {
		response.FailErr(c, apierr.Wrap(apierr.CodeInternalError, "list task submissions", err))
		return
	}
	response.OKMeta(c, items, &response.Meta{Page: page, PageSize: pageSize, Total: total})
}

func (h *Query) ListTaskReviews(c *gin.Context) {
	page, pageSize, err := parsePage(c)
	if err != nil {
		response.FailErr(c, err)
		return
	}
	items, total, err := h.reviews.ListByTask(c.Request.Context(), c.Param("taskId"), page, pageSize)
	if err != nil {
		response.FailErr(c, apierr.Wrap(apierr.CodeInternalError, "list task reviews", err))
		return
	}
	response.OKMeta(c, items, &response.Meta{Page: page, PageSize: pageSize, Total: total})
}

func (h *Query) ListTaskSettlements(c *gin.Context) {
	items, err := h.settlements.ListByTask(c.Request.Context(), c.Param("taskId"))
	if err != nil {
		response.FailErr(c, apierr.Wrap(apierr.CodeInternalError, "list task settlements", err))
		return
	}
	response.OK(c, items)
}

func (h *Query) ExecutorHistory(c *gin.Context) {
	page, pageSize, err := parsePage(c)
	if err != nil {
		response.FailErr(c, err)
		return
	}
	statuses, err := parseTaskStatuses(c.Query("status"))
	if err != nil {
		response.FailErr(c, err)
		return
	}
	items, total, err := h.tasks.ListByExecutor(c.Request.Context(), c.Param("executorId"), statuses, page, pageSize)
	if err != nil {
		response.FailErr(c, apierr.Wrap(apierr.CodeInternalError, "list executor history", err))
		return
	}
	out := make([]executorHistoryItem, 0, len(items))
	for _, item := range items {
		acceptedAt := latestAcceptedAt(c, h.reviews, item.TaskID)
		settledAt := latestSettledAt(c, h.settlements, item.TaskID)
		out = append(out, executorHistoryItem{
			TaskID:     item.TaskID,
			Title:      item.Title,
			Category:   item.Category,
			Status:     item.Status,
			Reward:     item.Reward,
			AcceptedAt: acceptedAt,
			SettledAt:  settledAt,
		})
	}
	response.OKMeta(c, out, &response.Meta{Page: page, PageSize: pageSize, Total: total})
}

func (h *Query) ListAccounts(c *gin.Context) {
	page, pageSize, err := parsePage(c)
	if err != nil {
		response.FailErr(c, err)
		return
	}
	filter := account.Filter{
		OwnerAccountID: strings.TrimSpace(c.Query("ownerAccountId")),
		NodeID:         strings.TrimSpace(c.Query("nodeId")),
		Keyword:        strings.TrimSpace(c.Query("keyword")),
		Page:           page,
		PageSize:       pageSize,
	}
	if rawType := strings.TrimSpace(c.Query("type")); rawType != "" {
		t, err := parseAccountType(rawType)
		if err != nil {
			response.FailErr(c, err)
			return
		}
		filter.Type = &t
	}
	if rawStatus := strings.TrimSpace(c.Query("status")); rawStatus != "" {
		s, err := parseAccountStatus(rawStatus)
		if err != nil {
			response.FailErr(c, err)
			return
		}
		filter.Status = &s
	}
	items, total, err := h.accounts.List(c.Request.Context(), filter)
	if err != nil {
		response.FailErr(c, apierr.Wrap(apierr.CodeInternalError, "list accounts", err))
		return
	}
	response.OKMeta(c, items, &response.Meta{Page: page, PageSize: pageSize, Total: total})
}

func parseTaskFilter(c *gin.Context) (task.Filter, int, int, error) {
	page, pageSize, err := parsePage(c)
	if err != nil {
		return task.Filter{}, 0, 0, err
	}
	statuses, err := parseTaskStatuses(c.Query("status"))
	if err != nil {
		return task.Filter{}, 0, 0, err
	}
	return task.Filter{
		Status:      statuses,
		Category:    strings.TrimSpace(c.Query("category")),
		RequesterID: strings.TrimSpace(c.Query("requesterId")),
		ExecutorID:  strings.TrimSpace(c.Query("executorId")),
		Keyword:     strings.TrimSpace(c.Query("keyword")),
		Page:        page,
		PageSize:    pageSize,
	}, page, pageSize, nil
}

func parsePage(c *gin.Context) (int, int, error) {
	page := 1
	pageSize := 20
	var err error
	if raw := strings.TrimSpace(c.Query("page")); raw != "" {
		page, err = strconv.Atoi(raw)
		if err != nil || page < 1 {
			return 0, 0, apierr.New(apierr.CodeInvalidRequest, "invalid page")
		}
	}
	if raw := strings.TrimSpace(c.Query("pageSize")); raw != "" {
		pageSize, err = strconv.Atoi(raw)
		if err != nil || pageSize < 1 || pageSize > 200 {
			return 0, 0, apierr.New(apierr.CodeInvalidRequest, "invalid pageSize")
		}
	}
	return page, pageSize, nil
}

func parseTaskStatuses(raw string) ([]task.Status, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	statuses := make([]task.Status, 0, len(parts))
	for _, part := range parts {
		s := task.Status(strings.ToUpper(strings.TrimSpace(part)))
		if !s.IsValid() {
			return nil, apierr.New(apierr.CodeInvalidRequest, "invalid task status")
		}
		statuses = append(statuses, s)
	}
	return statuses, nil
}

func parseAccountType(raw string) (account.Type, error) {
	t := account.Type(strings.ToLower(strings.TrimSpace(raw)))
	switch t {
	case account.TypeHuman, account.TypeAgent:
		return t, nil
	default:
		return "", apierr.New(apierr.CodeInvalidRequest, "invalid account type")
	}
}

func parseAccountStatus(raw string) (account.Status, error) {
	s := account.Status(strings.ToLower(strings.TrimSpace(raw)))
	switch s {
	case account.StatusActive, account.StatusDisabled, account.StatusPending:
		return s, nil
	default:
		return "", apierr.New(apierr.CodeInvalidRequest, "invalid account status")
	}
}

func repoToHTTPError(op string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, task.ErrTaskNotFound), errors.Is(err, account.ErrAccountNotFound), errors.Is(err, bid.ErrBidNotFound), errors.Is(err, submission.ErrSubmissionNotFound), errors.Is(err, settlement.ErrSettlementNotFound):
		return apierr.Wrap(apierr.CodeNotFound, op, err)
	default:
		return apierr.Wrap(apierr.CodeInternalError, op, err)
	}
}

func latestAcceptedAt(c *gin.Context, repo review.Repository, taskID string) *time.Time {
	items, _, err := repo.ListByTask(c.Request.Context(), taskID, 1, 100)
	if err != nil {
		return nil
	}
	for _, item := range items {
		if item.Decision == review.DecisionAccepted {
			at := item.ReviewedAt
			return &at
		}
	}
	return nil
}

func latestSettledAt(c *gin.Context, repo settlement.Repository, taskID string) *time.Time {
	items, err := repo.ListByTask(c.Request.Context(), taskID)
	if err != nil || len(items) == 0 {
		return nil
	}
	at := items[0].RecordedAt
	return &at
}
