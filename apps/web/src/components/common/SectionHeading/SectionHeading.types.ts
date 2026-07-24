export type SectionHeadingAlign = "left" | "center";

export type SectionHeadingProps = {
  eyebrow?: string;
  title: React.ReactNode;
  description?: string;
  align?: SectionHeadingAlign;
  id?: string;
  className?: string;
};
