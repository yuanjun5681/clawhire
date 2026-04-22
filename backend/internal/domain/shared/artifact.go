package shared

type ArtifactType string

const (
	ArtifactTypeURL   ArtifactType = "url"
	ArtifactTypeFile  ArtifactType = "file"
	ArtifactTypeJSON  ArtifactType = "json"
	ArtifactTypeText  ArtifactType = "text"
	ArtifactTypeImage ArtifactType = "image"
	ArtifactTypeRepo  ArtifactType = "repo"
)

// Artifact 描述进度、里程碑、交付等附件。
type Artifact struct {
	Type  ArtifactType `bson:"type"            json:"type"`
	Value string       `bson:"value"           json:"value"`
	Label string       `bson:"label,omitempty" json:"label,omitempty"`
}
