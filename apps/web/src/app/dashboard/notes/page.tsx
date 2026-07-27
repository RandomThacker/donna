"use client";

import { RequireAuth } from "@/features/auth";
import { Notes } from "@/features/notes";

export default function NotesPage() {
  return (
    <RequireAuth>
      <Notes />
    </RequireAuth>
  );
}
