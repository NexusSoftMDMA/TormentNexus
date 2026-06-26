import styles from "./Footer.module.css";

export function Footer() {
  return (
    <footer className={styles.foot}>
      <div className={styles.row}>
        <a href="#top" className={styles.mark}>
          AGENTMEMORY
        </a>
        <nav className={styles.links} aria-label="Footer">
          <a
            href="https://github.com/rohitg00/agentmemory"
            target="_blank"
            rel="noopener"
          >
            SOURCE
          </a>
          <a
            href="https://github.com/rohitg00/agentmemory/blob/main/CHANGELOG.md"
            target="_blank"
            rel="noopener"
          >
            CHANGELOG
          </a>
          <a href="https://iii.dev" target="_blank" rel="noopener">
            RUNS ON iii
          </a>
          <a
            href="https://github.com/rohitg00/agentmemory/blob/main/LICENSE"
            target="_blank"
            rel="noopener"
          >
            APACHE-2.0
          </a>
        </nav>
      </div>
      <div className={styles.fine}>© 2026 AGENTMEMORY · BUILT IN THE OPEN</div>
    </footer>
  );
}
