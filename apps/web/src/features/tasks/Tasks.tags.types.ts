export const TAG_COLOR_PRESETS = [
  "#c9a87c",
  "#7eb8da",
  "#86c79a",
  "#e8a87c",
  "#b794f4",
  "#f687b3",
  "#f6d860",
  "#ff8b94",
] as const;

export type TaskTag = {
  id: string;
  public_id: string;
  name: string;
  color: string;
  updated_at?: string;
};

export type TaskTagsResponse = {
  tags: TaskTag[];
};
