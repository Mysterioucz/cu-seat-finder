export default function Chip({
  label,
  pressed,
  onClick,
}: {
  label: string;
  pressed: boolean;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      aria-pressed={pressed}
      onClick={onClick}
      className={`min-h-11 rounded-full border px-3.5 py-2 text-sm transition-colors ${
        pressed
          ? "border-brand bg-brand text-brand-foreground"
          : "border-border bg-surface text-foreground hover:border-brand"
      }`}
    >
      {label}
    </button>
  );
}
