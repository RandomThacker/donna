import { Card, Icon } from "@/components/common";

import { landingCapabilitiesStyles as styles } from "./LandingCapabilities.styles";
import type { CapabilityCardProps } from "./LandingCapabilities.types";

export function CapabilityCard({ item }: CapabilityCardProps) {
  return (
    <Card interactive>
      <span className={styles.iconWrap}>
        <Icon name={item.icon} className="h-6 w-6" />
      </span>
      <h3 className={styles.title}>{item.title}</h3>
      <p className={styles.body}>{item.description}</p>
    </Card>
  );
}
