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
	Type ArtifactType `bson:"type"           json:"type"`
	URL  string       `bson:"url"            json:"url"`
	Name string       `bson:"name,omitempty" json:"name,omitempty"`
}
