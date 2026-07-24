"use client";

import { Icon } from "@/components/common";

import { useTheme } from "./ThemeProvider";

export function ThemeToggle({ className }: { className?: string }) {
  const { theme, toggleTheme } = useTheme();
  const next = theme === "dark" ? "light" : "dark";

  return (
    <button
      type="button"
      onClick={toggleTheme}
      className={className}
      aria-label={`Switch to ${next} theme`}
      title={`Switch to ${next} theme`}
    >
      <Icon name={theme === "dark" ? "sunrise" : "moon"} className="h-4 w-4" />
      <span>{theme === "dark" ? "Light mode" : "Dark mode"}</span>
    </button>
  );
}
