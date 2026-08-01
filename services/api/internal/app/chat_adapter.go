package app

import (
	"context"

	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/chat"
)

// chatCommandAdapter adapts chat.Executor to business.ChatCommandExecutor
// without creating a business ↔ chat import cycle.
type chatCommandAdapter struct {
	exec *chat.Executor
}

func (a chatCommandAdapter) Execute(ctx context.Context, in business.ChatCommandInput) business.ChatCommandResult {
	if a.exec == nil {
		return business.ChatCommandResult{}
	}
	out := a.exec.Execute(ctx, chat.ExecuteInput{
		UserID:          in.UserID,
		Timezone:        in.Timezone,
		Now:             in.Now,
		Message:         in.Message,
		DisplayName:     in.DisplayName,
		DryRun:          in.DryRun,
		SkipPersonality: in.SkipPersonality,
	})
	return business.ChatCommandResult{
		Reply:  out.Reply,
		Intent: string(out.Intent),
		Error:  out.Error,
	}
}
