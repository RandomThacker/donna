export type AutomationTrigger = {
  type: string;
  time: string;
  days?: string[];
};

export type AutomationDelivery = {
  channels: string[];
};

export type AutomationExecutionStatus =
  | "RUNNING"
  | "SUCCESS"
  | "PARTIAL_SUCCESS"
  | "FAILED"
  | "CANCELLED"
  | string;

export type AutomationCommandStatus = "SUCCESS" | "FAILED" | "SKIPPED" | string;

export type AutomationCommand = {
  command: string;
  variables?: Record<string, string>;
  label?: string;
};

export type Automation = {
  id: string;
  public_id: string;
  name: string;
  description?: string | null;
  enabled: boolean;
  trigger: AutomationTrigger;
  timezone: string;
  commands: AutomationCommand[];
  delivery: AutomationDelivery;
  template_id?: string | null;
  last_run_at?: string | null;
  next_run_at?: string | null;
  last_status?: string | null;
  success_rate?: number | null;
  average_duration_ms?: number | null;
  last_commands_total?: number | null;
  last_commands_success?: number | null;
  total_executions?: number;
  created_at: string;
  updated_at: string;
};

export type AutomationsListResponse = {
  automations: Automation[];
};

export type AutomationTemplate = {
  id: string;
  name: string;
  description: string;
  commands: AutomationCommand[];
  default_schedule: AutomationTrigger;
};

export type AutomationTemplatesResponse = {
  templates: AutomationTemplate[];
};

export type CreateAutomationInput = {
  name?: string;
  description?: string;
  enabled?: boolean;
  trigger?: AutomationTrigger;
  timezone: string;
  commands?: Array<AutomationCommand | string>;
  delivery?: AutomationDelivery;
  template_id?: string;
};

export type UpdateAutomationInput = {
  name?: string;
  description?: string;
  enabled?: boolean;
  trigger?: AutomationTrigger;
  timezone?: string;
  commands?: Array<AutomationCommand | string>;
  delivery?: AutomationDelivery;
};

export type AutomationCommandExecution = {
  id: string;
  public_id: string;
  order_index: number;
  command: string;
  command_type?: string | null;
  started_at: string;
  completed_at?: string | null;
  status: AutomationCommandStatus;
  duration_ms?: number | null;
  response?: string | null;
  error?: string | null;
};

export type AutomationExecution = {
  id: string;
  public_id: string;
  automation_id: string;
  automation_name?: string | null;
  started_at: string;
  completed_at?: string | null;
  status: AutomationExecutionStatus;
  duration_ms?: number | null;
  commands_total: number;
  commands_success: number;
  commands_failed: number;
  trigger_source: string;
  delivery: AutomationDelivery;
  delivery_status?: string | null;
  response?: string | null;
  error?: string | null;
  commands?: AutomationCommandExecution[];
  created_at: string;
  updated_at: string;
  debug?: Record<string, unknown>;
};

export type AutomationHistoryResponse = {
  executions: AutomationExecution[];
};

export type AutomationAnalytics = {
  total_executions: number;
  success_rate: number;
  failure_rate: number;
  average_duration_ms?: number | null;
  average_commands_per_run?: number | null;
  most_frequent_automation_id?: string | null;
  most_frequent_automation_name?: string | null;
};

export type AutomationRunCommandResult = {
  order_index: number;
  command: string;
  command_key?: string;
  command_type?: string;
  status: AutomationCommandStatus;
  duration_ms: number;
  response?: string;
  error?: string;
};

export type AutomationRunResult = {
  response: string;
  status: AutomationExecutionStatus;
  delivery_status: string;
  commands_total: number;
  commands_success: number;
  commands_failed: number;
  duration_ms: number;
  trigger_source: string;
  commands: AutomationRunCommandResult[];
  execution_id?: string | null;
  dry_run: boolean;
};
