package command

import (
	"fmt"
	"strings"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawhire"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

func validateActor(actor shared.Actor, field string) error {
	if strings.TrimSpace(actor.ID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, fmt.Sprintf("%s.id is required", field))
	}
	if !actor.Kind.Valid() {
		return apierr.New(apierr.CodeInvalidMessagePayload, fmt.Sprintf("%s.kind is invalid", field))
	}
	return nil
}

func validatePostTask(p clawhire.PostTaskPayload) error {
	if strings.TrimSpace(p.TaskID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId is required")
	}
	if strings.TrimSpace(p.Title) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "title is required")
	}
	if strings.TrimSpace(p.Category) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "category is required")
	}
	if err := validateActor(p.Requester, "requester"); err != nil {
		return err
	}
	if p.Reviewer != nil {
		if err := validateActor(*p.Reviewer, "reviewer"); err != nil {
			return err
		}
	}
	if strings.TrimSpace(p.Reward.Mode) == "" || strings.TrimSpace(p.Reward.Currency) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "reward.mode and reward.currency are required")
	}
	return nil
}

func validatePlaceBid(p clawhire.PlaceBidPayload) error {
	if strings.TrimSpace(p.TaskID) == "" || strings.TrimSpace(p.BidID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId and bidId are required")
	}
	if err := validateActor(p.Executor, "executor"); err != nil {
		return err
	}
	if strings.TrimSpace(p.Currency) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "currency is required")
	}
	return nil
}

func validateAwardTask(p clawhire.AwardTaskPayload) error {
	if strings.TrimSpace(p.TaskID) == "" || strings.TrimSpace(p.ContractID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId and contractId are required")
	}
	if err := validateActor(p.Executor, "executor"); err != nil {
		return err
	}
	if strings.TrimSpace(p.AgreedReward.Currency) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "agreedReward.currency is required")
	}
	return nil
}

func validateSubmission(p clawhire.CreateSubmissionPayload) error {
	if strings.TrimSpace(p.TaskID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId is required")
	}
	if strings.TrimSpace(p.Summary) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "summary is required")
	}
	return validateArtifacts(p.Artifacts)
}

func validateArtifacts(items []shared.Artifact) error {
	for i, item := range items {
		if strings.TrimSpace(string(item.Type)) == "" {
			return apierr.New(apierr.CodeInvalidMessagePayload, fmt.Sprintf("artifacts[%d].type is required", i))
		}
		if item.Type == shared.ArtifactTypeURL && strings.TrimSpace(item.URL) == "" {
			return apierr.New(apierr.CodeInvalidMessagePayload, fmt.Sprintf("artifacts[%d].url is required", i))
		}
	}
	return nil
}

func validateAcceptSubmission(p clawhire.AcceptSubmissionPayload) error {
	if strings.TrimSpace(p.TaskID) == "" || strings.TrimSpace(p.SubmissionID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId and submissionId are required")
	}
	return validateActor(p.AcceptedBy, "acceptedBy")
}

func validateRejectSubmission(p clawhire.RejectSubmissionPayload) error {
	if strings.TrimSpace(p.TaskID) == "" || strings.TrimSpace(p.SubmissionID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId and submissionId are required")
	}
	if strings.TrimSpace(p.Reason) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "reason is required")
	}
	return validateActor(p.RejectedBy, "rejectedBy")
}

func validateSettlement(p clawhire.RecordSettlementPayload) error {
	if strings.TrimSpace(p.TaskID) == "" || strings.TrimSpace(p.SettlementID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId and settlementId are required")
	}
	if p.Amount <= 0 {
		return apierr.New(apierr.CodeInvalidMessagePayload, "amount must be greater than 0")
	}
	if strings.TrimSpace(p.Currency) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "currency is required")
	}
	return validateActor(p.Payee, "payee")
}
