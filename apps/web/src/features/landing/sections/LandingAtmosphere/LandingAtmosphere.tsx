import { atmosphereStyles as styles } from "./LandingAtmosphere.styles";

export function LandingAtmosphere() {
  return (
    <div className={styles.root} aria-hidden>
      <div className={styles.base} />
      <div className={styles.bloomTop} />
      <div className={styles.bloomLeft} />
      <div className={styles.bloomRight} />
      <div className={styles.vignette} />
      <div className={styles.grain} />
    </div>
  );
}
