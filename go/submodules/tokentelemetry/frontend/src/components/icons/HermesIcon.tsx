import { forwardRef } from "react";
import type { LucideIcon, LucideProps } from "lucide-react";

const HermesIcon = forwardRef<SVGSVGElement, LucideProps>((props, ref) => (
  <svg
    ref={ref}
    xmlns="http://www.w3.org/2000/svg"
    width={props.size ?? 24}
    height={props.size ?? 24}
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth={props.strokeWidth ?? 2}
    strokeLinecap="round"
    strokeLinejoin="round"
    {...props}
  >
    {/* Central staff */}
    <line x1="12" y1="3" x2="12" y2="21" />
    {/* Wings */}
    <path d="M8 5L4.5 2.5L6.5 6" />
    <path d="M8 5L5.5 6.5L7.5 8" />
    <path d="M16 5L19.5 2.5L17.5 6" />
    <path d="M16 5L18.5 6.5L16.5 8" />
    {/* Snake coils */}
    <path d="M9 9C7.5 11 7.5 13 9 15" />
    <path d="M9 15C10 17 11 17 12 16" />
    <path d="M15 9C16.5 11 16.5 13 15 15" />
    <path d="M15 15C14 17 13 17 12 16" />
  </svg>
));
HermesIcon.displayName = "HermesIcon";

export default HermesIcon as LucideIcon;
