package pi

import "encoding/json"

const (
	DefaultMaxLines = 2000
	DefaultMaxBytes = 50 * 1024
)

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ImageContent struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	MimeType string `json:"mimeType"`
}

type ContentBlock interface{}

type ReadToolInput struct {
	Path   string `json:"path"`
	Offset int    `json:"offset,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

type WriteToolInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type EditReplacement struct {
	OldText string `json:"oldText"`
	NewText string `json:"newText"`
}

type EditToolInput struct {
	Path  string            `json:"path"`
	Edits []EditReplacement `json:"edits"`
}

type BashToolInput struct {
	Command string  `json:"command"`
	Timeout float64 `json:"timeout,omitempty"`
}

type TruncationDetails struct {
	Truncated          bool   `json:"truncated"`
	TruncatedBy        string `json:"truncatedBy,omitempty"`
	TotalLines         int    `json:"totalLines,omitempty"`
	OutputLines        int    `json:"outputLines,omitempty"`
	OutputBytes        int    `json:"outputBytes,omitempty"`
	MaxLines           int    `json:"maxLines,omitempty"`
	MaxBytes           int    `json:"maxBytes,omitempty"`
	FirstLineExceeds   bool   `json:"firstLineExceedsLimit,omitempty"`
	LastLinePartial    bool   `json:"lastLinePartial,omitempty"`
	ContinuationOffset int    `json:"continuationOffset,omitempty"`
	FullOutputPath     string `json:"fullOutputPath,omitempty"`
}

type ReadToolDetails struct {
	Truncation *TruncationDetails `json:"truncation,omitempty"`
}

type EditToolDetails struct {
	Diff             string `json:"diff,omitempty"`
	FirstChangedLine int    `json:"firstChangedLine,omitempty"`
}

type BashToolDetails struct {
	Truncation     *TruncationDetails `json:"truncation,omitempty"`
	FullOutputPath string             `json:"fullOutputPath,omitempty"`
}

type GrepToolInput struct {
	Pattern    string `json:"pattern"`
	Path       string `json:"path,omitempty"`
	Glob       string `json:"glob,omitempty"`
	IgnoreCase bool   `json:"ignoreCase,omitempty"`
	Literal    bool   `json:"literal,omitempty"`
	Context    int    `json:"context,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

type GrepToolDetails struct {
	Truncation        *TruncationDetails `json:"truncation,omitempty"`
	MatchLimitReached int                `json:"matchLimitReached,omitempty"`
	LinesTruncated    bool               `json:"linesTruncated,omitempty"`
}

type FindToolInput struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

type FindToolDetails struct {
	Truncation         *TruncationDetails `json:"truncation,omitempty"`
	ResultLimitReached int                `json:"resultLimitReached,omitempty"`
}

type LsToolInput struct {
	Path  string `json:"path,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

type LsToolDetails struct {
	Truncation        *TruncationDetails `json:"truncation,omitempty"`
	EntryLimitReached int                `json:"entryLimitReached,omitempty"`
}

type ToolResult struct {
	ToolName string      `json:"toolName"`
	Content  []any       `json:"content"`
	Details  interface{} `json:"details,omitempty"`
	IsError  bool        `json:"isError"`
}

type RunEvent struct {
	Type      RunEventType    `json:"type"`
	Timestamp int64           `json:"timestamp"`
	SessionID string          `json:"sessionId,omitempty"`
	ToolName  string          `json:"toolName,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	Result    *ToolResult     `json:"result,omitempty"`
	Error     string          `json:"error,omitempty"`
}
