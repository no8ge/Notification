package types

type MarkdownMsg struct {
	Msgtype  string            `json:"msgtype"`
	Markdown map[string]string `json:"markdown"`
}
