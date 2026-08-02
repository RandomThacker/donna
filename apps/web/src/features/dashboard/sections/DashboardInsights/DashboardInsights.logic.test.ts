import { describe, expect, it } from "vitest";

import { buildDashboardInsights } from "./DashboardInsights.logic";

describe("buildDashboardInsights", () => {
  it("surfaces focus window, task progress, and busy tomorrow", () => {
    const now = new Date("2026-08-02T04:00:00.000Z"); // 09:30 IST
    const insights = buildDashboardInsights({
      now,
      timeZone: "Asia/Kolkata",
      todayEvents: [
        {
          title: "Standup",
          all_day: false,
          start_time: "2026-08-02T04:00:00.000Z", // 09:30
          end_time: "2026-08-02T04:30:00.000Z",
        },
        {
          title: "Client",
          all_day: false,
          start_time: "2026-08-02T10:00:00.000Z", // 15:30
          end_time: "2026-08-02T11:00:00.000Z",
        },
      ],
      tomorrowEvents: [
        { all_day: false, start_time: "2026-08-03T04:00:00.000Z", end_time: "2026-08-03T05:00:00.000Z" },
        { all_day: false, start_time: "2026-08-03T06:00:00.000Z", end_time: "2026-08-03T07:00:00.000Z" },
        { all_day: false, start_time: "2026-08-03T08:00:00.000Z", end_time: "2026-08-03T09:00:00.000Z" },
        { all_day: false, start_time: "2026-08-03T10:00:00.000Z", end_time: "2026-08-03T11:00:00.000Z" },
      ],
      todayTasks: [
        { completed: true },
        { completed: true },
        { completed: false },
        { completed: true },
        { completed: false },
      ],
    });

    expect(insights.map((item) => item.id)).toEqual([
      "focus-window",
      "tomorrow-busy",
      "tasks-progress",
    ]);
    expect(insights[0]?.text).toMatch(/focus window/i);
    expect(insights[1]?.text).toMatch(/Tomorrow looks busy/i);
    expect(insights[2]?.text).toMatch(/60% of today's tasks done/);
  });

  it("flags a meeting starting soon", () => {
    const now = new Date("2026-08-02T08:00:00.000Z");
    const insights = buildDashboardInsights({
      now,
      timeZone: "Asia/Kolkata",
      todayEvents: [
        {
          title: "Design review",
          all_day: false,
          start_time: "2026-08-02T08:25:00.000Z",
          end_time: "2026-08-02T09:00:00.000Z",
        },
      ],
      tomorrowEvents: [],
      todayTasks: [],
    });

    expect(insights[0]).toEqual({
      id: "meeting-soon",
      text: "Design review starts in 25 min.",
    });
  });

  it("falls back when the day is quiet", () => {
    const now = new Date("2026-08-02T08:00:00.000Z");
    const insights = buildDashboardInsights({
      now,
      timeZone: "Asia/Kolkata",
      todayEvents: [],
      tomorrowEvents: [],
      todayTasks: [],
    });

    expect(insights.some((item) => item.id === "clear-day")).toBe(true);
    expect(insights.some((item) => item.id === "tomorrow-clear")).toBe(true);
  });
});
