package sessionimport

type ValidationResult struct {
	SourceTool     string   `json:"sourceTool"`
	SourceType     string   `json:"sourceType"`
	SourcePath     string   `json:"sourcePath"`
	Format         string   `json:"format"`
	LastModifiedAt string   `json:"lastModifiedAt,omitempty"`
	EstimatedSize  int64    `json:"estimatedSize"`
	Valid          bool     `json:"valid"`
	DetectedModels []string `json:"detectedModels"`
	Errors         []string `json:"errors,omitempty"`
}

type Manifest struct {
	GeneratedAt string             `json:"generatedAt"`
	Count       int                `json:"count"`
	Candidates  []ValidationResult `json:"candidates"`
}
