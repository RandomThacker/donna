package webpush

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	webpushlib "github.com/SherClockHolmes/webpush-go"
)

// Payload is the JSON body delivered to the service worker.
type Payload struct {
	Title          string `json:"title"`
	Body           string `json:"body"`
	OccurrenceID   string `json:"occurrenceId,omitempty"`
	TimelineType   string `json:"timelineType,omitempty"`
	Source         string `json:"source,omitempty"`
	StartTime      string `json:"startTime,omitempty"`
	DeepLink       string `json:"deepLink"`
	NotificationID string `json:"notificationId,omitempty"`
}

// Result describes one push attempt.
type Result struct {
	StatusCode int
	Gone       bool // endpoint should be unsubscribed (404/410)
}

// Sender delivers Web Push messages using VAPID.
type Sender interface {
	Send(ctx context.Context, sub entity.PushSubscription, payload Payload) (Result, error)
	Configured() bool
}

type vapidSender struct {
	publicKey  string
	privateKey string
	subject    string
	ttl        int
}

// NewSender builds a VAPID Web Push sender. Empty keys → Configured() false.
func NewSender(publicKey, privateKey, subject string) Sender {
	if subject == "" {
		subject = "mailto:donna@localhost"
	}
	return &vapidSender{
		publicKey:  publicKey,
		privateKey: privateKey,
		subject:    subject,
		ttl:        constant.WebPushTTL,
	}
}

func (s *vapidSender) Configured() bool {
	return s != nil && s.publicKey != "" && s.privateKey != ""
}

func (s *vapidSender) Send(ctx context.Context, sub entity.PushSubscription, payload Payload) (Result, error) {
	if !s.Configured() {
		return Result{}, errors.New("web push is not configured")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Result{}, fmt.Errorf("marshal push payload: %w", err)
	}

	resp, err := webpushlib.SendNotificationWithContext(ctx, body, &webpushlib.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpushlib.Keys{
			P256dh: sub.P256dh,
			Auth:   sub.Auth,
		},
	}, &webpushlib.Options{
		Subscriber:      s.subject,
		VAPIDPublicKey:  s.publicKey,
		VAPIDPrivateKey: s.privateKey,
		TTL:             s.ttl,
		Urgency:         webpushlib.UrgencyNormal,
	})
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()

	gone := resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{StatusCode: resp.StatusCode, Gone: gone}, fmt.Errorf("push endpoint returned %d", resp.StatusCode)
	}
	return Result{StatusCode: resp.StatusCode}, nil
}

// PayloadFromNotification builds a browser payload from a queued notification.
func PayloadFromNotification(n entity.Notification) Payload {
	p := Payload{
		Title:          n.Title,
		Body:           n.Body,
		NotificationID: n.ID.String(),
		DeepLink:       constant.NotificationDeepLinkPath,
	}
	if n.OccurrenceID != nil {
		p.OccurrenceID = *n.OccurrenceID
		p.DeepLink = constant.NotificationDeepLinkPath + *n.OccurrenceID
	}
	if n.NotificationType != nil {
		p.TimelineType = *n.NotificationType
	}
	if len(n.Payload) > 0 {
		var raw map[string]any
		if err := json.Unmarshal(n.Payload, &raw); err == nil {
			if v, ok := raw["source"].(string); ok {
				p.Source = v
			}
			if v, ok := raw["type"].(string); ok && p.TimelineType == "" {
				p.TimelineType = v
			}
			if v, ok := raw["startAt"].(string); ok {
				p.StartTime = v
			}
			if v, ok := raw["deepLink"].(string); ok && v != "" {
				p.DeepLink = v
			}
			if v, ok := raw["occurrenceId"].(string); ok && p.OccurrenceID == "" {
				p.OccurrenceID = v
				if p.DeepLink == constant.NotificationDeepLinkPath {
					p.DeepLink = constant.NotificationDeepLinkPath + v
				}
			}
		}
	}
	if p.StartTime == "" && n.ScheduledFor != nil {
		p.StartTime = n.ScheduledFor.UTC().Format(time.RFC3339Nano)
	}
	return p
}
