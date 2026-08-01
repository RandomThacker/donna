import { describe, expect, it } from "vitest";

import {
  buildCreatePayloadFromTemplate,
  commandLabel,
  defaultBrowserTimezone,
  formatLocalTimeForInput,
  formatSchedule,
} from "./Automations.logic";
import type { Automation, AutomationTemplate } from "./Automations.types";

describe("automations", () => {
  it("builds create payload from a catalog template", () => {
    const template: AutomationTemplate = {
      id: "morning_brief",
      name: "Morning Brief",
      description: "Agenda plus tasks",
      commands: [
        { command: "todays_agenda", variables: { range: "today" }, label: "Today's Agenda" },
        { command: "tasks_due", variables: { priority: "all" }, label: "Tasks Due" },
      ],
      default_schedule: { type: "daily", time: "09:00" },
    };
    const payload = buildCreatePayloadFromTemplate(
      template,
      defaultBrowserTimezone(),
      "09:30",
    );
    expect(payload.template_id).toBe("morning_brief");
    expect(payload.trigger).toEqual({ type: "daily", time: "09:30" });
    expect(payload.commands).toHaveLength(2);
    expect(payload.commands[0]).toEqual({
      command: "todays_agenda",
      variables: { range: "today" },
    });
    expect(payload.delivery.channels).toEqual(["chat"]);
    expect(payload.enabled).toBe(true);
  });

  it("formats schedule and readable command labels", () => {
    const auto: Automation = {
      id: "1",
      public_id: "atm_x",
      name: "Task Review",
      enabled: true,
      trigger: { type: "daily", time: "10:00:00" },
      timezone: "UTC",
      commands: [{ command: "tasks_due", variables: { priority: "all" } }],
      delivery: { channels: ["chat"] },
      created_at: "",
      updated_at: "",
    };
    expect(formatLocalTimeForInput(auto.trigger.time)).toBe("10:00");
    expect(formatSchedule(auto)).toBe("Daily at 10:00");
    expect(commandLabel(auto.commands[0]!)).toBe("Tasks Due");
  });
});
