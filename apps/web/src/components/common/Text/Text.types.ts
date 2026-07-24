export type TextTone = "cream" | "muted" | "copper";
export type TextAs = "p" | "span" | "h1" | "h2" | "h3";

export type TextProps = {
  children: React.ReactNode;
  as?: TextAs;
  tone?: TextTone;
  className?: string;
  id?: string;
};
