import { addDays, format, isToday } from "date-fns";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useState } from "react";

import {
  createTask,
  deleteTask,
  fetchTaskDay,
  reorderTaskOccurrences,
  updateTaskOccurrence,
} from "./Tasks.api";
import type { TaskDayResponse, TaskOccurrence } from "./Tasks.types";

export const taskQueryKeys = {
  all: ["tasks"] as const,
  day: (date: string) => ["tasks", "day", date] as const,
  history: (from: string, to: string) => ["tasks", "history", from, to] as const,
};

export function formatJournalDate(date: Date): string {
  return format(date, "yyyy-MM-dd");
}

export function useTaskJournal() {
  const queryClient = useQueryClient();
  const [selectedDate, setSelectedDate] = useState(() => new Date());
  const [draftTitle, setDraftTitle] = useState("");

  const dateKey = formatJournalDate(selectedDate);

  const dayQuery = useQuery({
    queryKey: taskQueryKeys.day(dateKey),
    queryFn: ({ signal }) => fetchTaskDay(dateKey, signal),
  });

  const occurrences = dayQuery.data?.occurrences ?? [];
  const statistics = dayQuery.data?.statistics;

  const invalidateDay = useCallback(async () => {
    await queryClient.invalidateQueries({ queryKey: taskQueryKeys.day(dateKey) });
  }, [queryClient, dateKey]);

  const createMutation = useMutation({
    mutationFn: (title: string) => createTask({ title, date: dateKey }),
    onSuccess: async () => {
      setDraftTitle("");
      await invalidateDay();
    },
  });

  const completeMutation = useMutation({
    mutationFn: ({ id, completed }: { id: string; completed: boolean }) =>
      updateTaskOccurrence(id, completed),
    onSuccess: async () => {
      await invalidateDay();
    },
  });

  const reorderMutation = useMutation({
    mutationFn: (ids: string[]) =>
      reorderTaskOccurrences({ date: dateKey, occurrence_ids: ids }),
    onMutate: async (ids) => {
      await queryClient.cancelQueries({ queryKey: taskQueryKeys.day(dateKey) });
      const previous = queryClient.getQueryData<TaskDayResponse>(
        taskQueryKeys.day(dateKey),
      );
      if (previous) {
        const byId = new Map(
          previous.occurrences.map((occurrence) => [occurrence.id, occurrence]),
        );
        const nextOccurrences = ids
          .map((id) => byId.get(id))
          .filter((occurrence): occurrence is TaskOccurrence => Boolean(occurrence))
          .map((occurrence, index) => ({ ...occurrence, sort_order: index }));
        queryClient.setQueryData<TaskDayResponse>(taskQueryKeys.day(dateKey), {
          ...previous,
          occurrences: nextOccurrences,
        });
      }
      return { previous };
    },
    onError: (_error, _ids, context) => {
      if (context?.previous) {
        queryClient.setQueryData(taskQueryKeys.day(dateKey), context.previous);
      }
    },
    onSettled: async () => {
      await invalidateDay();
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (taskId: string) => deleteTask(taskId),
    onSuccess: async () => {
      await invalidateDay();
    },
  });

  const goToday = useCallback(() => {
    setSelectedDate(new Date());
  }, []);

  const goPrevDay = useCallback(() => {
    setSelectedDate((d) => addDays(d, -1));
  }, []);

  const goNextDay = useCallback(() => {
    setSelectedDate((d) => addDays(d, 1));
  }, []);

  const selectDay = useCallback((day: Date) => {
    setSelectedDate(day);
  }, []);

  const addTask = useCallback(
    (titleOverride?: string) => {
      const title = (titleOverride ?? draftTitle).trim();
      if (!title || createMutation.isPending) {
        return;
      }
      createMutation.mutate(title);
    },
    [createMutation, draftTitle],
  );

  const toggleComplete = useCallback(
    (occurrence: TaskOccurrence) => {
      completeMutation.mutate({
        id: occurrence.id,
        completed: !occurrence.completed,
      });
    },
    [completeMutation],
  );

  const reorder = useCallback(
    (fromId: string, toId: string) => {
      if (fromId === toId) {
        return;
      }
      const ids = occurrences.map((o) => o.id);
      const fromIndex = ids.indexOf(fromId);
      const toIndex = ids.indexOf(toId);
      if (fromIndex < 0 || toIndex < 0) {
        return;
      }
      const next = [...ids];
      const [moved] = next.splice(fromIndex, 1);
      next.splice(toIndex, 0, moved!);
      reorderMutation.mutate(next);
    },
    [occurrences, reorderMutation],
  );

  const removeTask = useCallback(
    (occurrence: TaskOccurrence) => {
      deleteMutation.mutate(occurrence.task_id);
    },
    [deleteMutation],
  );

  const titleLabel = isToday(selectedDate)
    ? `Today · ${format(selectedDate, "MMMM d, yyyy")}`
    : format(selectedDate, "EEEE, MMMM d, yyyy");

  return {
    selectedDate,
    dateKey,
    titleLabel,
    occurrences,
    statistics,
    draftTitle,
    setDraftTitle,
    isLoading: dayQuery.isLoading,
    isError: dayQuery.isError,
    isSaving:
      createMutation.isPending ||
      completeMutation.isPending ||
      reorderMutation.isPending ||
      deleteMutation.isPending,
    goToday,
    goPrevDay,
    goNextDay,
    selectDay,
    addTask,
    toggleComplete,
    reorder,
    removeTask,
  };
}

export function sourceGlyph(source: TaskOccurrence["source"]): string | null {
  switch (source) {
    case "carry_forward":
      return "↻";
    case "recurring":
      return "🔁";
    case "ai":
      return "🤖";
    case "calendar":
      return "📅";
    default:
      return null;
  }
}
