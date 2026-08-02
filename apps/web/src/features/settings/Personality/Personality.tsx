"use client";

import { useEffect, useState } from "react";

import {
  fetchPersonalityCatalog,
  fetchPersonalityProfile,
  updatePersonalityProfile,
} from "./Personality.api";
import { EMOJI_LEVELS } from "./Personality.logic";
import { personalityStyles as styles } from "./Personality.styles";
import type {
  PersonalityDefinition,
  PersonalityProfile,
} from "./Personality.types";

export function PersonalityPanel() {
  const [catalog, setCatalog] = useState<PersonalityDefinition[]>([]);
  const [form, setForm] = useState<PersonalityProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [savedNotice, setSavedNotice] = useState(false);

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
  }, []);

  function patchForm(partial: Partial<PersonalityProfile>) {
    setForm((prev) => (prev ? { ...prev, ...partial } : prev));
    setSavedNotice(false);
  }

  async function handleSave() {
    if (!form) return;
    setSaving(true);
    setError(null);
    setSavedNotice(false);
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
      setSavedNotice(true);
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
        <p className={styles.hint}>
          {savedNotice
            ? "Saved — chat and automations will use this."
            : "Hit Save so chat and automations pick this up."}
        </p>
      </div>

      {error ? <p className={styles.error}>{error}</p> : null}
    </div>
  );
}
