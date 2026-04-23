package command

import (
	"strings"

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

func toAPIError(op string, err error) error {
	switch err {
	case nil:
		return nil
	case task.ErrTaskNotFound:
		return apierr.Wrap(apierr.CodeNotFound, op, err)
	case task.ErrStatusConflict:
		return apierr.Wrap(apierr.CodeInvalidState, op, err)
	default:
		return apierr.Wrap(apierr.CodeInternalError, op, err)
	}
}
