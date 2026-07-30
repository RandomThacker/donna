import { describe, expect, it } from "vitest";

import type { DonnaNotification } from "./Notifications.types";
import {
  buildStatusTimeline,
  filterMatchesStatus,
  filterNotifications,
  groupNotifications,
  matchesSearch,
  unreadCount,
} from "./Notifications.utils";

function n(
  partial: Partial<DonnaNotification> & Pick<DonnaNotification, "id" | "title">,
): DonnaNotification {
  return {
    public_id: `ntf_${partial.id}`,
    body: partial.body ?? "",
    status: partial.status ?? "SENT",
    delivery_channels: partial.delivery_channels ?? ["CHAT"],
    created_at: partial.created_at ?? "2026-07-30T10:00:00.000Z",
    updated_at: partial.updated_at ?? "2026-07-30T10:00:00.000Z",
    ...partial,
  };
}

describe("filterMatchesStatus", () => {
  it("maps unread to SENT", () => {
    expect(filterMatchesStatus("unread", "SENT")).toBe(true);
    expect(filterMatchesStatus("unread", "READ")).toBe(false);
    expect(filterMatchesStatus("unread", "PENDING")).toBe(false);
  });

  it("matches explicit statuses", () => {
    expect(filterMatchesStatus("pending", "PENDING")).toBe(true);
    expect(filterMatchesStatus("failed", "FAILED")).toBe(true);
    expect(filterMatchesStatus("dismissed", "DISMISSED")).toBe(true);
    expect(filterMatchesStatus("all", "READ")).toBe(true);
  });
});

describe("matchesSearch", () => {
  it("searches title and body", () => {
    const item = n({
      id: "1",
      title: "Guitar Class",
      body: "Starts in 15 minutes",
    });
    expect(matchesSearch(item, "guitar")).toBe(true);
    expect(matchesSearch(item, "15 minutes")).toBe(true);
    expect(matchesSearch(item, "water")).toBe(false);
    expect(matchesSearch(item, "")).toBe(true);
  });
});

describe("unreadCount", () => {
  it("counts SENT only", () => {
    const items = [
      n({ id: "1", title: "A", status: "SENT" }),
      n({ id: "2", title: "B", status: "READ" }),
      n({ id: "3", title: "C", status: "SENT" }),
      n({ id: "4", title: "D", status: "PENDING" }),
    ];
    expect(unreadCount(items)).toBe(2);
  });
});

describe("groupNotifications", () => {
  it("buckets into today / yesterday / earlier / older", () => {
    const now = new Date("2026-07-30T15:00:00");
    const items = [
      n({
        id: "1",
        title: "Today",
        scheduled_for: "2026-07-30T12:00:00.000Z",
        created_at: "2026-07-30T12:00:00.000Z",
      }),
      n({
        id: "2",
        title: "Yesterday",
        scheduled_for: "2026-07-29T12:00:00.000Z",
        created_at: "2026-07-29T12:00:00.000Z",
      }),
      n({
        id: "3",
        title: "Old",
        scheduled_for: "2026-07-01T12:00:00.000Z",
        created_at: "2026-07-01T12:00:00.000Z",
      }),
    ];
    const groups = groupNotifications(items, now);
    const byKey = Object.fromEntries(groups.map((g) => [g.key, g.items.map((i) => i.id)]));
    expect(byKey.today).toEqual(["1"]);
    expect(byKey.yesterday).toEqual(["2"]);
    expect(byKey.older).toEqual(["3"]);
  });
});

describe("filterNotifications", () => {
  it("filters unread and sorts newest first", () => {
    const items = [
      n({
        id: "old",
        title: "Old sent",
        status: "SENT",
        scheduled_for: "2026-07-30T08:00:00.000Z",
      }),
      n({
        id: "new",
        title: "New sent",
        status: "SENT",
        scheduled_for: "2026-07-30T18:00:00.000Z",
      }),
      n({ id: "read", title: "Read", status: "READ" }),
    ];
    const filtered = filterNotifications(items, "unread", "");
    expect(filtered.map((i) => i.id)).toEqual(["new", "old"]);
  });
});

describe("buildStatusTimeline", () => {
  it("marks created and queued for pending", () => {
    const steps = buildStatusTimeline(
      n({
        id: "1",
        title: "Pending",
        status: "PENDING",
        created_at: "2026-07-30T10:00:00.000Z",
      }),
    );
    expect(steps.find((s) => s.id === "created")?.done).toBe(true);
    expect(steps.find((s) => s.id === "sent")?.done).toBe(false);
  });

  it("includes sent and read for READ status", () => {
    const steps = buildStatusTimeline(
      n({
        id: "1",
        title: "Read",
        status: "READ",
        created_at: "2026-07-30T10:00:00.000Z",
        sent_at: "2026-07-30T10:00:02.000Z",
        read_at: "2026-07-30T10:01:00.000Z",
      }),
    );
    expect(steps.find((s) => s.id === "sent")?.done).toBe(true);
    expect(steps.find((s) => s.id === "read")?.done).toBe(true);
    expect(steps.find((s) => s.id === "dismissed")?.done).toBe(false);
  });

  it("stops at failed for FAILED status", () => {
    const steps = buildStatusTimeline(
      n({
        id: "1",
        title: "Failed",
        status: "FAILED",
        created_at: "2026-07-30T10:00:00.000Z",
      }),
    );
    expect(steps.map((s) => s.id)).toEqual(["created", "queued", "sent"]);
    expect(steps.find((s) => s.id === "sent")?.label).toContain("Failed");
  });
});
