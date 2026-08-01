package model

import "github.com/RandomThacker/donna/services/api/internal/personality"

// PersonalityProfileResponse is GET/PATCH /settings/personality.
type PersonalityProfileResponse struct {
	PersonalityID      string `json:"personality_id"`
	DisplayName        string `json:"display_name"`
	Nickname           string `json:"nickname"`
	EmojiLevel         string `json:"emoji_level"`
	HumorLevel         string `json:"humor_level"`
	GreetingStyle      string `json:"greeting_style"`
	EncouragementLevel string `json:"encouragement_level"`
	ResponseStyle      string `json:"response_style"`
}

// UpdatePersonalityRequest is PATCH /settings/personality.
type UpdatePersonalityRequest struct {
	PersonalityID      *string `json:"personality_id"`
	DisplayName        *string `json:"display_name"`
	Nickname           *string `json:"nickname"`
	EmojiLevel         *string `json:"emoji_level"`
	HumorLevel         *string `json:"humor_level"`
	GreetingStyle      *string `json:"greeting_style"`
	EncouragementLevel *string `json:"encouragement_level"`
	ResponseStyle      *string `json:"response_style"`
}

// PersonalityDefinitionResponse is one catalog entry.
type PersonalityDefinitionResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// PersonalityPreviewRequest is POST /settings/personality/preview.
type PersonalityPreviewRequest struct {
	Timezone           string  `json:"timezone"`
	PersonalityID      *string `json:"personality_id"`
	DisplayName        *string `json:"display_name"`
	Nickname           *string `json:"nickname"`
	EmojiLevel         *string `json:"emoji_level"`
	HumorLevel         *string `json:"humor_level"`
	GreetingStyle      *string `json:"greeting_style"`
	EncouragementLevel *string `json:"encouragement_level"`
	ResponseStyle      *string `json:"response_style"`
}

// PersonalityPreviewResponse is live preview samples.
type PersonalityPreviewResponse struct {
	Greeting      string `json:"greeting"`
	Reminder      string `json:"reminder"`
	TaskComplete  string `json:"task_complete"`
	Error         string `json:"error"`
	Notification  string `json:"notification"`
	Automation    string `json:"automation"`
	MorningBrief  string `json:"morning_brief"`
	Chat          string `json:"chat"`
}

// PersonalityProfileFromDomain maps a profile to the API response.
func PersonalityProfileFromDomain(p personality.Profile) PersonalityProfileResponse {
	return PersonalityProfileResponse{
		PersonalityID:      p.PersonalityID,
		DisplayName:        p.DisplayName,
		Nickname:           p.Nickname,
		EmojiLevel:         string(p.EmojiLevel),
		HumorLevel:         string(p.HumorLevel),
		GreetingStyle:      p.GreetingStyle,
		EncouragementLevel: string(p.EncouragementLevel),
		ResponseStyle:      p.ResponseStyle,
	}
}

// PersonalityDefinitionsFromDomain maps catalog entries.
func PersonalityDefinitionsFromDomain(defs []personality.Definition) []PersonalityDefinitionResponse {
	out := make([]PersonalityDefinitionResponse, 0, len(defs))
	for _, d := range defs {
		out = append(out, PersonalityDefinitionResponse{
			ID:          d.ID,
			Name:        d.Name,
			Description: d.Description,
		})
	}
	return out
}

// PersonalityPreviewFromMap maps preview samples.
func PersonalityPreviewFromMap(m map[string]string) PersonalityPreviewResponse {
	return PersonalityPreviewResponse{
		Greeting:     m["greeting"],
		Reminder:     m["reminder"],
		TaskComplete: m["task_complete"],
		Error:        m["error"],
		Notification: m["notification"],
		Automation:   m["automation"],
		MorningBrief: m["morning_brief"],
		Chat:         m["chat"],
	}
}
