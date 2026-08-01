"use client";

import { useCallback, useEffect, useState } from "react";

import {
  fetchPersonalityCatalog,
  fetchPersonalityProfile,
  previewPersonality,
  updatePersonalityProfile,
} from "./Personality.api";
import {
  defaultBrowserTimezone,
  EMOJI_LEVELS,
  PREVIEW_LABELS,
} from "./Personality.logic";
import { personalityStyles as styles } from "./Personality.styles";
import type {
  PersonalityDefinition,
  PersonalityPreview,
  PersonalityProfile,
} from "./Personality.types";

const EMPTY_PREVIEW: PersonalityPreview = {
  greeting: "",
  reminder: "",
  task_complete: "",
  error: "",
  notification: "",
  automation: "",
  morning_brief: "",
  chat: "",
};

export function PersonalityPanel() {
  const [catalog, setCatalog] = useState<PersonalityDefinition[]>([]);
  const [form, setForm] = useState<PersonalityProfile | null>(null);
  const [preview, setPreview] = useState<PersonalityPreview>(EMPTY_PREVIEW);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const timezone = defaultBrowserTimezone();

  const refreshPreview = useCallback(
    async (next: PersonalityProfile, signal?: AbortSignal) => {
      try {
        const samples = await previewPersonality({
          timezone,
          personality_id: next.personality_id,
          display_name: next.display_name,
          nickname: next.nickname,
          emoji_level: next.emoji_level,
          humor_level: next.humor_level,
          greeting_style: next.greeting_style,
          encouragement_level: next.encouragement_level,
          response_style: next.response_style,
        });
        if (signal?.aborted) return;
        setPreview(samples);
      } catch (err) {
        if (signal?.aborted) return;
        setError(err instanceof Error ? err.message : "Couldn’t preview personality");
      }
    },
    [timezone],
  );

  useEffect(() => {
    const controller = new AbortController();
    void (async () => {
      setLoading(true);
      setError(null);
      try {
        const [profile, catalogData] = await Promise.all([
          fetchPersonalityProfile(controller.signal),
          fetchPersonalityCatalog(controller.signal),
        ]);
        if (controller.signal.aborted) return;
        setForm(profile);
        setCatalog(catalogData.personalities ?? []);
        await refreshPreview(profile, controller.signal);
      } catch (err) {
        if (controller.signal.aborted) return;
        setError(err instanceof Error ? err.message : "Couldn’t load personality");
      } finally {
        if (!controller.signal.aborted) {
          setLoading(false);
        }
      }
    })();
    return () => controller.abort();
  }, [refreshPreview]);

  function patchForm(partial: Partial<PersonalityProfile>) {
    setForm((prev) => {
      if (!prev) return prev;
      const next = { ...prev, ...partial };
      void refreshPreview(next);
      return next;
    });
  }

  async function handleSave() {
    if (!form) return;
    setSaving(true);
    setError(null);
    try {
      const saved = await updatePersonalityProfile({
        personality_id: form.personality_id,
        display_name: form.display_name,
        nickname: form.nickname,
        emoji_level: form.emoji_level,
        humor_level: form.humor_level,
        greeting_style: form.greeting_style,
        encouragement_level: form.encouragement_level,
        response_style: form.response_style,
      });
      setForm(saved);
      await refreshPreview(saved);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Couldn’t save personality");
    } finally {
      setSaving(false);
    }
  }

  if (loading || !form) {
    return <p className={styles.hint}>Loading personality…</p>;
  }

  const selected = catalog.find((item) => item.id === form.personality_id);

  return (
    <div className={styles.root}>
      <div className={styles.grid}>
        <div className={styles.field}>
          <label className={styles.label} htmlFor="personality-id">
            Personality
          </label>
          <select
            id="personality-id"
            className={styles.select}
            value={form.personality_id}
            onChange={(event) => patchForm({ personality_id: event.target.value })}
          >
            {catalog.map((item) => (
              <option key={item.id} value={item.id}>
                {item.name}
              </option>
            ))}
          </select>
          {selected ? (
            <p className={styles.optionHint}>{selected.description}</p>
          ) : null}
        </div>
        <div className={styles.field}>
          <label className={styles.label} htmlFor="emoji-level">
            Emoji Level
          </label>
          <select
            id="emoji-level"
            className={styles.select}
            value={form.emoji_level}
            onChange={(event) => patchForm({ emoji_level: event.target.value })}
          >
            {EMOJI_LEVELS.map((level) => (
              <option key={level} value={level}>
                {level.charAt(0).toUpperCase() + level.slice(1)}
              </option>
            ))}
          </select>
        </div>
        <div className={styles.field}>
          <label className={styles.label} htmlFor="preferred-name">
            Preferred Name
          </label>
          <input
            id="preferred-name"
            className={styles.input}
            value={form.display_name}
            onChange={(event) => patchForm({ display_name: event.target.value })}
            placeholder="Aryan"
          />
        </div>
        <div className={styles.field}>
          <label className={styles.label} htmlFor="nickname">
            Nickname
          </label>
          <input
            id="nickname"
            className={styles.input}
            value={form.nickname}
            onChange={(event) => patchForm({ nickname: event.target.value })}
            placeholder="Rockstar"
          />
        </div>
      </div>

      <div className={styles.actions}>
        <button
          type="button"
          className={styles.saveBtn}
          disabled={saving}
          onClick={() => void handleSave()}
        >
          {saving ? "Saving…" : "Save personality"}
        </button>
        <p className={styles.hint}>Preview updates as you type.</p>
      </div>

      <div className={styles.previewGrid}>
        {(
          [
            "greeting",
            "morning_brief",
            "reminder",
            "task_complete",
            "chat",
            "automation",
            "notification",
            "error",
          ] as const
        ).map((key) => (
          <div key={key} className={styles.previewCard}>
            <p className={styles.previewLabel}>{PREVIEW_LABELS[key]}</p>
            <p className={styles.previewBody}>{preview[key] || "—"}</p>
          </div>
        ))}
      </div>

      {error ? <p className={styles.error}>{error}</p> : null}
    </div>
  );
}
