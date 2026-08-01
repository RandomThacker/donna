package provider

import (
	"context"
	"strings"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/logger"
)

// EstBytesPerRow* are engineering estimates for Sprint 1A metrics (not Neon-measured).
const (
	EstBytesPerRowCalendarScheduler = 500  // was ~3–15 KB with provider_payload
	EstBytesPerRowDonnaEvent        = 400  // was ~0.3–2 KB with color/timestamps
	EstBytesPerRowDonnaReminder     = 300
)

func logSchedulerQuery(
	ctx context.Context,
	log *logger.Logger,
	providerName string,
	columns []string,
	estBytesPerRow int,
	rows int,
	duration time.Duration,
) {
	if log == nil {
		return
	}
	log.Info(ctx, "occurrence provider query",
		"provider", providerName,
		"columns_selected", strings.Join(columns, ","),
		"column_count", len(columns),
		"est_bytes_per_row", estBytesPerRow,
		"rows_returned", rows,
		"est_bytes_total", estBytesPerRow*rows,
		"duration_ms", duration.Milliseconds(),
	)
}
