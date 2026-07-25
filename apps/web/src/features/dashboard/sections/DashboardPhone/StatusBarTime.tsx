"use client";

import { useEffect, useState } from "react";

export function formatPhoneTime(date = new Date()) {
  return new Intl.DateTimeFormat("en-US", {
    hour: "numeric",
    minute: "2-digit",
    hour12: true,
  }).format(date);
}

export function StatusBarTime() {
  const [time, setTime] = useState(() => formatPhoneTime());

  useEffect(() => {
    const tick = () => setTime(formatPhoneTime());
    tick();

    const intervalId = window.setInterval(tick, 15_000);
    return () => window.clearInterval(intervalId);
  }, []);

  return <span>{time}</span>;
}
