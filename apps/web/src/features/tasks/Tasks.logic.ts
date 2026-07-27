import {
  addDays,
  addMonths,
  endOfMonth,
  format,
  isSameMonth,
  isToday,
  startOfMonth,
  subMonths,
} from "date-fns";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useMemo, useState } from "react";

import {
  createTask,
  deleteTask,
  fetchTaskDay,
  fetchTaskHistory,
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

export function buildMonthGrid(cursor: Date): Date[] {
  const start = startOfMonth(cursor);
  const end = endOfMonth(cursor);
  const gridStart = new Date(start);
  gridStart.setDate(gridStart.getDate() - gridStart.getDay());
  const gridEnd = new Date(end);
  gridEnd.setDate(gridEnd.getDate() + (6 - gridEnd.getDay()));
  const days: Date[] = [];
  let current = gridStart;
  while (current <= gridEnd) {
    days.push(new Date(current));
    current = addDays(current, 1);
  }
  return days;
}

export function useTaskJournal() {
  const queryClient = useQueryClient();
  const [selectedDate, setSelectedDate] = useState(() => new Date());
  const [miniMonth, setMiniMonth] = useState(() => new Date());
  const [draftTitle, setDraftTitle] = useState("");

  const dateKey = formatJournalDate(selectedDate);
  const historyFrom = formatJournalDate(startOfMonth(miniMonth));
  const historyTo = formatJournalDate(endOfMonth(miniMonth));

  const dayQuery = useQuery({
    queryKey: taskQueryKeys.day(dateKey),
    queryFn: ({ signal }) => fetchTaskDay(dateKey, signal),
  });

  const historyQuery = useQuery({
    queryKey: taskQueryKeys.history(historyFrom, historyTo),
    queryFn: ({ signal }) => fetchTaskHistory(historyFrom, historyTo, signal),
  });

  const historyByDate = useMemo(() => {
    const map = new Map<string, { total: number; completed: number; pending: number }>();
    for (const day of historyQuery.data?.days ?? []) {
      map.set(day.date, day);
    }
    return map;
  }, [historyQuery.data?.days]);

  const occurrences = dayQuery.data?.occurrences ?? [];
  const statistics = dayQuery.data?.statistics;

  const invalidateDay = useCallback(async () => {
    await queryClient.invalidateQueries({ queryKey: taskQueryKeys.day(dateKey) });
    await queryClient.invalidateQueries({
      queryKey: taskQueryKeys.history(historyFrom, historyTo),
    });
  }, [queryClient, dateKey, historyFrom, historyTo]);

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
    const today = new Date();
    setSelectedDate(today);
    setMiniMonth(today);
  }, []);

  const goPrevDay = useCallback(() => {
    setSelectedDate((d) => addDays(d, -1));
  }, []);

  const goNextDay = useCallback(() => {
    setSelectedDate((d) => addDays(d, 1));
  }, []);

  const selectDay = useCallback((day: Date) => {
    setSelectedDate(day);
    if (!isSameMonth(day, miniMonth)) {
      setMiniMonth(day);
    }
  }, [miniMonth]);

  const shiftMiniMonth = useCallback((dir: -1 | 1) => {
    setMiniMonth((m) => (dir === 1 ? addMonths(m, 1) : subMonths(m, 1)));
  }, []);

  const addTask = useCallback(() => {
    const title = draftTitle.trim();
    if (!title) {
      return;
    }
    createMutation.mutate(title);
  }, [createMutation, draftTitle]);

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
    miniMonth,
    dateKey,
    titleLabel,
    occurrences,
    statistics,
    draftTitle,
    setDraftTitle,
    historyByDate,
    miniDays: buildMonthGrid(miniMonth),
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
    shiftMiniMonth,
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
