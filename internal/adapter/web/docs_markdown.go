package web

import (
	_ "embed"
	"net/http"
)

//go:embed docs_api.md
var docsAPIMarkdown []byte

// handleDocsMarkdown 返回可供下载与渲染的 API Markdown 文档。
func (s *Server) handleDocsMarkdown(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(docsAPIMarkdown)
}
