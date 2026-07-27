import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useMemo, useState } from "react";

import { createNote, deleteNote, fetchNotes, updateNote } from "./Notes.api";
import type { CreateNoteInput, Note, NoteColor, UpdateNoteInput } from "./Notes.types";

export const noteQueryKeys = {
  all: ["notes"] as const,
  list: () => ["notes", "list"] as const,
};

export function useNotes() {
  const queryClient = useQueryClient();
  const [composerTitle, setComposerTitle] = useState("");
  const [composerContent, setComposerContent] = useState("");
  const [composerColor, setComposerColor] = useState<NoteColor>("default");
  const [composerOpen, setComposerOpen] = useState(false);
  const [activeNoteId, setActiveNoteId] = useState<string | null>(null);
  const [draftTitle, setDraftTitle] = useState("");
  const [draftContent, setDraftContent] = useState("");
  const [draftColor, setDraftColor] = useState<NoteColor>("default");

  const listQuery = useQuery({
    queryKey: noteQueryKeys.list(),
    queryFn: ({ signal }) => fetchNotes(signal),
  });

  const notes = listQuery.data?.notes ?? [];

  const invalidate = useCallback(async () => {
    await queryClient.invalidateQueries({ queryKey: noteQueryKeys.list() });
  }, [queryClient]);

  const createMutation = useMutation({
    mutationFn: (input: CreateNoteInput) => createNote(input),
    onSuccess: async () => {
      setComposerTitle("");
      setComposerContent("");
      setComposerColor("default");
      setComposerOpen(false);
      await invalidate();
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateNoteInput }) =>
      updateNote(id, input),
    onSuccess: async () => {
      await invalidate();
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteNote(id),
    onSuccess: async () => {
      setActiveNoteId(null);
      await invalidate();
    },
  });

  const pinnedNotes = useMemo(
    () => notes.filter((note) => note.pinned),
    [notes],
  );
  const otherNotes = useMemo(
    () => notes.filter((note) => !note.pinned),
    [notes],
  );

  const activeNote = useMemo(
    () => notes.find((note) => note.id === activeNoteId) ?? null,
    [notes, activeNoteId],
  );

  const openComposer = useCallback(() => {
    setComposerOpen(true);
  }, []);

  const closeComposer = useCallback(() => {
    setComposerOpen(false);
    setComposerTitle("");
    setComposerContent("");
    setComposerColor("default");
  }, []);

  const submitComposer = useCallback(() => {
    const title = composerTitle.trim();
    const content = composerContent.trim();
    if (!title && !content) {
      return;
    }
    createMutation.mutate({
      title,
      content,
      color: composerColor,
    });
  }, [composerTitle, composerContent, composerColor, createMutation]);

  const openNote = useCallback((note: Note) => {
    setActiveNoteId(note.id);
    setDraftTitle(note.title);
    setDraftContent(note.content);
    setDraftColor(note.color);
  }, []);

  const closeNote = useCallback(() => {
    setActiveNoteId(null);
  }, []);

  const saveActiveNote = useCallback(() => {
    if (!activeNoteId) {
      return;
    }
    const title = draftTitle.trim();
    const content = draftContent.trim();
    if (!title && !content) {
      return;
    }
    updateMutation.mutate({
      id: activeNoteId,
      input: { title, content, color: draftColor },
    });
  }, [activeNoteId, draftTitle, draftContent, draftColor, updateMutation]);

  const togglePin = useCallback(
    (note: Note) => {
      updateMutation.mutate({
        id: note.id,
        input: { pinned: !note.pinned },
      });
    },
    [updateMutation],
  );

  const removeNote = useCallback(
    (id: string) => {
      deleteMutation.mutate(id);
    },
    [deleteMutation],
  );

  return {
    notes,
    pinnedNotes,
    otherNotes,
    isLoading: listQuery.isLoading,
    isError: listQuery.isError,
    isSaving:
      createMutation.isPending ||
      updateMutation.isPending ||
      deleteMutation.isPending,
    composerOpen,
    composerTitle,
    composerContent,
    composerColor,
    setComposerTitle,
    setComposerContent,
    setComposerColor,
    openComposer,
    closeComposer,
    submitComposer,
    activeNote,
    draftTitle,
    draftContent,
    draftColor,
    setDraftTitle,
    setDraftContent,
    setDraftColor,
    openNote,
    closeNote,
    saveActiveNote,
    togglePin,
    removeNote,
  };
}
