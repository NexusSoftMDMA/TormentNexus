interface TriadLogoProps {
  className?: string;
}

export function TriadLogo({ className }: TriadLogoProps) {
  return (
    <svg
      viewBox="0 0 32 32"
      aria-hidden="true"
      className={className}
      fill="currentColor"
    >
      <circle cx="11" cy="20" r="3.2" fillOpacity="0.55" />
      <circle cx="21" cy="20" r="3.2" fillOpacity="0.55" />
      <circle cx="16" cy="11" r="3.2" />
    </svg>
  );
}
