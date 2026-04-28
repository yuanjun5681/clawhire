package webhook

import (
	"encoding/json"
	"testing"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawhire"
)

func TestValidateSubmissionRejectsURLArtifactWithoutURL(t *testing.T) {
	var payload clawhire.CreateSubmissionPayload
	raw := []byte(`{
		"taskId": "task_001",
		"summary": "Done",
		"artifacts": [
			{"type": "url", "name": "Result"}
		]
	}`)
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if err := validateSubmission(payload); err == nil {
		t.Fatalf("validateSubmission succeeded, want missing artifacts[0].url error")
	}
}

func TestValidateSubmissionAcceptsProtocolArtifactURLField(t *testing.T) {
	payload := clawhire.CreateSubmissionPayload{
		TaskID:  "task_001",
		Summary: "Done",
		Artifacts: []shared.Artifact{
			{Type: shared.ArtifactTypeURL, URL: "https://example.com/result", Name: "Result"},
		},
	}

	if err := validateSubmission(payload); err != nil {
		t.Fatalf("validateSubmission err = %v", err)
	}
}
