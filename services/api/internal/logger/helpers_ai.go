package logger

import (
	"context"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
)

// AIUsageEvent captures one LLM invocation for usage tracking and future dashboards.
type AIUsageEvent struct {
	Model                 string
	Provider              string
	InputTokens           int
	OutputTokens          int
	Latency               time.Duration
	EstimatedCostUSD      float64
	PromptVersion         string
	ToolsUsed             []string
	MemoryRetrievalCount  int
	ConversationID        string
	UserID                string
}

// AIUsage logs a structured AI usage record at INFO, WARNing if over the AI budget.
func (l *Logger) AIUsage(ctx context.Context, e AIUsageEvent) {
	if e.ConversationID != "" || e.UserID != "" {
		ctx = WithFields(ctx, Fields{
			ConversationID: e.ConversationID,
			UserID:         e.UserID,
		})
	}

	args := []any{
		constant.LogAttrEvent, "ai_usage",
		constant.LogAttrModel, e.Model,
		constant.LogAttrProvider, e.Provider,
		constant.LogAttrInputTokens, e.InputTokens,
		constant.LogAttrOutputTokens, e.OutputTokens,
		constant.LogAttrLatencyMS, e.Latency.Milliseconds(),
		constant.LogAttrEstimatedCostUSD, e.EstimatedCostUSD,
		constant.LogAttrPromptVersion, e.PromptVersion,
		constant.LogAttrToolsUsed, e.ToolsUsed,
		constant.LogAttrMemoryRetrievalCount, e.MemoryRetrievalCount,
	}

	if e.Latency >= constant.BudgetAIRequest {
		l.Warn(ctx, "ai request exceeded budget", args...)
		return
	}
	l.Info(ctx, "ai usage", args...)
}
