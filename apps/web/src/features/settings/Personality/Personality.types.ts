export type PersonalityId = "professional" | "casual" | "flirty" | string;

export type PersonalityLevel = "none" | "low" | "medium" | "high" | string;

export type PersonalityDefinition = {
  id: PersonalityId;
  name: string;
  description: string;
};

export type PersonalityCatalogResponse = {
  personalities: PersonalityDefinition[];
};

export type PersonalityProfile = {
  personality_id: PersonalityId;
  display_name: string;
  nickname: string;
  emoji_level: PersonalityLevel;
  humor_level: PersonalityLevel;
  greeting_style: string;
  encouragement_level: PersonalityLevel;
  response_style: string;
};

export type UpdatePersonalityInput = {
  personality_id?: string;
  display_name?: string;
  nickname?: string;
  emoji_level?: string;
  humor_level?: string;
  greeting_style?: string;
  encouragement_level?: string;
  response_style?: string;
};
