package command

import (
	"fmt"
	"strings"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/bid"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/contract"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/review"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/submission"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawhire"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

func normalizeAcceptanceSpec(spec *clawhire.AcceptanceSpec) task.AcceptanceSpec {
	if spec == nil {
		return task.AcceptanceSpec{Mode: task.AcceptanceModeManual}
	}
	return task.AcceptanceSpec{
		Mode:  task.AcceptanceMode(clawhire.NormalizeAcceptanceMode(spec.Mode)),
		Rules: spec.Rules,
	}
}

func normalizeSettlementTerms(terms *clawhire.SettlementTerms) *task.SettlementTerms {
	if terms == nil {
		return &task.SettlementTerms{Trigger: "on_acceptance"}
	}
	trigger := strings.TrimSpace(terms.Trigger)
	if trigger == "" {
		trigger = "on_acceptance"
	}
	return &task.SettlementTerms{Trigger: trigger}
}

func normalizeEvidence(ev *clawhire.Evidence) *submission.Evidence {
	if ev == nil {
		return nil
	}
	return &submission.Evidence{
		Type:  strings.TrimSpace(ev.Type),
		Items: ev.Items,
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func reviewID(meta *EventMeta, taskID, submissionID string, decision review.Decision) string {
	if meta != nil && strings.TrimSpace(meta.ID) != "" {
		return "review:" + strings.TrimSpace(meta.ID)
	}
	return fmt.Sprintf("review:%s:%s:%s", strings.TrimSpace(taskID), strings.TrimSpace(submissionID), decision)
}

func toAPIError(op string, err error) error {
	switch err {
	case nil:
		return nil
	case task.ErrTaskNotFound, submission.ErrSubmissionNotFound, bid.ErrBidNotFound, contract.ErrContractNotFound:
		return apierr.Wrap(apierr.CodeNotFound, op, err)
	case task.ErrStatusConflict:
		return apierr.Wrap(apierr.CodeInvalidState, op, err)
	default:
		return apierr.Wrap(apierr.CodeInternalError, op, err)
	}
}
