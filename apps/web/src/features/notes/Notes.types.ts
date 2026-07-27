export type NoteColor =
  | "default"
  | "coral"
  | "sage"
  | "sky"
  | "blush"
  | "sand"
  | "lilac";

export type Note = {
  id: string;
  public_id: string;
  title: string;
  content: string;
  color: NoteColor;
  pinned: boolean;
  archived: boolean;
  created_at: string;
  updated_at: string;
};

export type NotesListResponse = {
  notes: Note[];
};

export type CreateNoteInput = {
  title?: string;
  content?: string;
  color?: NoteColor;
  pinned?: boolean;
};

export type UpdateNoteInput = {
  title?: string;
  content?: string;
  color?: NoteColor;
  pinned?: boolean;
  archived?: boolean;
};

export const NOTE_COLORS: NoteColor[] = [
  "default",
  "coral",
  "sage",
  "sky",
  "blush",
  "sand",
  "lilac",
];
