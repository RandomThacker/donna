"use client";

import { usePathname } from "next/navigation";
import { useState } from "react";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";
import { useAuth } from "@/features/auth";
import { navItemsForPath } from "@/features/dashboard/dashboardNav";
import { DashboardSidebar } from "@/features/dashboard/sections/DashboardSidebar";

import { useNotes } from "./Notes.logic";
import {
  useMasonryColumnCount,
  usePackedNoteColumns,
} from "./Notes.masonry";
import { noteColorClass, noteColorDot, notesStyles as styles } from "./Notes.styles";
import { NOTE_COLORS, type Note, type NoteColor } from "./Notes.types";

function initialsFrom(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) {
    return "D";
  }
  if (parts.length === 1) {
    return parts[0]!.slice(0, 2).toUpperCase();
  }
  return `${parts[0]![0] ?? ""}${parts[1]![0] ?? ""}`.toUpperCase();
}

function ColorPicker({
  value,
  onChange,
}: {
  value: NoteColor;
  onChange: (color: NoteColor) => void;
}) {
  return (
    <div className={styles.colorRow} role="group" aria-label="Note color">
      {NOTE_COLORS.map((color) => (
        <button
          key={color}
          type="button"
          className={cn(
            styles.colorDot,
            noteColorDot[color],
            value === color && styles.colorDotActive,
          )}
          aria-label={color}
          aria-pressed={value === color}
          onClick={() => onChange(color)}
        />
      ))}
    </div>
  );
}

function NoteCard({
  note,
  onOpen,
}: {
  note: Note;
  onOpen: (note: Note) => void;
}) {
  return (
    <button
      type="button"
      className={cn(styles.card, noteColorClass[note.color] ?? noteColorClass.default)}
      onClick={() => onOpen(note)}
    >
      {note.pinned ? (
        <span className={styles.cardPin} aria-label="Pinned">
          <Icon name="pin" className="h-3.5 w-3.5" />
        </span>
      ) : null}
      {note.title ? <p className={styles.cardTitle}>{note.title}</p> : null}
      {note.content ? <p className={styles.cardBody}>{note.content}</p> : null}
      {!note.title && !note.content ? (
        <p className={styles.cardBody}>Empty note</p>
      ) : null}
    </button>
  );
}

function NotesMasonry({
  notes,
  onOpen,
}: {
  notes: Note[];
  onOpen: (note: Note) => void;
}) {
  const [container, setContainer] = useState<HTMLDivElement | null>(null);
  const columnCount = useMasonryColumnCount(container);
  const columns = usePackedNoteColumns(notes, columnCount);

  return (
    <div ref={setContainer} className={styles.masonry}>
      {columns.map((column, index) => (
        <div key={index} className={styles.masonryColumn}>
          {column.map((note) => (
            <NoteCard key={note.id} note={note} onOpen={onOpen} />
          ))}
        </div>
      ))}
    </div>
  );
}

export function Notes() {
  const pathname = usePathname();
  const { user } = useAuth();
  const notes = useNotes();
  const nav = navItemsForPath(pathname);

  const profileName =
    user?.display_name?.trim() || user?.email?.split("@")[0] || "You";
  const profileInitials = initialsFrom(profileName);

  const closeAndSave = () => {
    if (notes.activeNote) {
      const dirty =
        notes.draftTitle.trim() !== notes.activeNote.title ||
        notes.draftContent.trim() !== notes.activeNote.content ||
        notes.draftColor !== notes.activeNote.color;
      if (dirty) {
        notes.saveActiveNote();
      }
    }
    notes.closeNote();
  };

  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <DashboardSidebar
          items={nav}
          profileName={profileName}
          profileInitials={profileInitials}
          profileEmail={user?.email}
          profileAvatarUrl={user?.avatar_url}
        />
        <main className={styles.workspace}>
          <div className={styles.workspaceInner}>
            <header className={styles.header}>
              <h1 className={styles.title}>Notes</h1>
              <p className={styles.subtitle}>
                Capture thoughts, lists, and scraps — keep them handy.
              </p>
            </header>

            <div className={styles.composer}>
              {!notes.composerOpen ? (
                <button
                  type="button"
                  className="w-full px-1 py-1.5 text-left text-sm text-donna-muted"
                  onClick={notes.openComposer}
                >
                  Take a note…
                </button>
              ) : (
                <>
                  <input
                    className={styles.composerTitle}
                    placeholder="Title"
                    value={notes.composerTitle}
                    onChange={(event) => notes.setComposerTitle(event.target.value)}
                    autoFocus
                  />
                  <textarea
                    className={styles.composerBody}
                    placeholder="Take a note…"
                    rows={3}
                    value={notes.composerContent}
                    onChange={(event) => notes.setComposerContent(event.target.value)}
                  />
                  <div className={styles.composerFooter}>
                    <ColorPicker
                      value={notes.composerColor}
                      onChange={notes.setComposerColor}
                    />
                    <div className={styles.composerActions}>
                      <button
                        type="button"
                        className={styles.ghostBtn}
                        onClick={notes.closeComposer}
                      >
                        Close
                      </button>
                      <button
                        type="button"
                        className={styles.primaryBtn}
                        disabled={notes.isSaving}
                        onClick={notes.submitComposer}
                      >
                        Done
                      </button>
                    </div>
                  </div>
                </>
              )}
            </div>

            {notes.isLoading ? (
              <p className={styles.state}>Loading notes…</p>
            ) : null}
            {notes.isError ? (
              <p className={styles.state}>Couldn&apos;t load notes.</p>
            ) : null}
            {!notes.isLoading && !notes.isError && notes.notes.length === 0 ? (
              <p className={styles.empty}>
                No notes yet. Tap above and jot something down.
              </p>
            ) : null}

            {notes.pinnedNotes.length > 0 ? (
              <section className={styles.section} aria-label="Pinned">
                <p className={styles.sectionLabel}>Pinned</p>
                <NotesMasonry notes={notes.pinnedNotes} onOpen={notes.openNote} />
              </section>
            ) : null}

            {notes.otherNotes.length > 0 ? (
              <section className={styles.section} aria-label="Others">
                {notes.pinnedNotes.length > 0 ? (
                  <p className={styles.sectionLabel}>Others</p>
                ) : null}
                <NotesMasonry notes={notes.otherNotes} onOpen={notes.openNote} />
              </section>
            ) : null}
          </div>
        </main>
      </div>

      {notes.activeNote ? (
        <div className={styles.editorRoot} role="presentation">
          <button
            type="button"
            className={styles.editorBackdrop}
            aria-label="Close note"
            onClick={closeAndSave}
          />
          <div
            role="dialog"
            aria-modal="true"
            aria-label="Edit note"
            className={cn(
              styles.editorPanel,
              noteColorClass[notes.draftColor] ?? noteColorClass.default,
            )}
          >
            <input
              className={styles.editorTitle}
              placeholder="Title"
              value={notes.draftTitle}
              onChange={(event) => notes.setDraftTitle(event.target.value)}
              autoFocus
            />
            <textarea
              className={styles.editorBody}
              placeholder="Note"
              value={notes.draftContent}
              onChange={(event) => notes.setDraftContent(event.target.value)}
            />
            <div className={styles.editorFooter}>
              <ColorPicker
                value={notes.draftColor}
                onChange={notes.setDraftColor}
              />
              <div className={styles.composerActions}>
                <button
                  type="button"
                  className={styles.ghostBtn}
                  aria-label={notes.activeNote.pinned ? "Unpin" : "Pin"}
                  onClick={() => notes.togglePin(notes.activeNote!)}
                >
                  <Icon name="pin" className="h-4 w-4" />
                </button>
                <button
                  type="button"
                  className={styles.ghostBtn}
                  aria-label="Delete note"
                  onClick={() => notes.removeNote(notes.activeNote!.id)}
                >
                  <Icon name="trash" className="h-4 w-4" />
                </button>
                <button
                  type="button"
                  className={styles.primaryBtn}
                  disabled={notes.isSaving}
                  onClick={closeAndSave}
                >
                  Done
                </button>
              </div>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
