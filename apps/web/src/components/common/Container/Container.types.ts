export type ContainerWidth = "narrow" | "default" | "wide";

export type ContainerProps = {
  children: React.ReactNode;
  width?: ContainerWidth;
  className?: string;
  as?: "div" | "section" | "header" | "footer" | "main";
};
