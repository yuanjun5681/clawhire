package webhook

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/account"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/bid"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/contract"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/milestone"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/settlement"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/submission"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawhire"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawsynapse"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

func decodeMessage(env *clawsynapse.Envelope, out interface{}) error {
	if err := env.DecodeMessage(out); err != nil {
		return apierr.Wrap(apierr.CodeInvalidMessagePayload, "decode message payload", err)
	}
	return nil
}

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

func normalizeClaim(claim *clawhire.Claim) *milestone.Claim {
	if claim == nil {
		return nil
	}
	return &milestone.Claim{
		Type:     strings.TrimSpace(claim.Type),
		Amount:   claim.Amount,
		Currency: strings.TrimSpace(claim.Currency),
	}
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

func chooseTime(ts *time.Time, now Now) time.Time {
	if ts != nil && !ts.IsZero() {
		return ts.UTC()
	}
	return now().UTC()
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func payloadMap(payload interface{}) (map[string]interface{}, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func toAPIError(op string, err error) error {
	switch err {
	case nil:
		return nil
	case task.ErrTaskNotFound, submission.ErrSubmissionNotFound, bid.ErrBidNotFound, contract.ErrContractNotFound, settlement.ErrSettlementNotFound, account.ErrAccountNotFound:
		return apierr.Wrap(apierr.CodeNotFound, op, err)
	case task.ErrStatusConflict:
		return apierr.Wrap(apierr.CodeInvalidState, op, err)
	default:
		return apierr.Wrap(apierr.CodeInternalError, op, err)
	}
}
