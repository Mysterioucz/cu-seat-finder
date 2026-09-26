# CU SeatFinder — Design System (visual redesign)

All tokens live in `frontend/src/app/globals.css` under `@theme` (Tailwind v4). No component library (ADR-002) — tokens back plain Tailwind utility classes plus a handful of semantic ones.

## Color

Warm neutrals (not cool gray) + CU pink used **sparingly** as an accent, not a wash.

| Token | Light | Dark | Use |
|---|---|---|---|
| `--color-background` | `#FAF9F7` | `#17140F` | page bg |
| `--color-surface` | `#FFFFFF` | `#211D18` | cards, inputs |
| `--color-surface-muted` | `#F1EEE8` | `#2A251E` | filter panel, skeletons, disabled |
| `--color-border` | `#E4DFD6` | `#3A342B` | card/input borders |
| `--color-foreground` | `#241F1C` | `#F3EFE9` | body text |
| `--color-muted-foreground` | `#6B6459` | `#B3A99A` | secondary text |
| `--color-brand` | `#C13E72` | `#F0A0C2` | buttons, links, active state — AA-safe against surface |
| `--color-brand-soft` | `#DE5C8E` (CU pink) | `#DE5C8E` | decorative accent only (dots, badges, chip fill) — not depended on for text contrast |
| `--color-brand-foreground` | `#FFFFFF` | `#241019` | text on brand-filled elements |

Brand pink (`#DE5C8E`, the CU identity color) is kept as `--color-brand-soft` for decorative accents; a slightly deepened `#C13E72` is `--color-brand` for anything carrying text (buttons/links), because `#DE5C8E` on white is ~3.5:1 — fine for large/graphical elements, not for button labels. This is the one deliberate deviation from the literal brand hex, done to pass AA.

### Density (never color-only — always paired with a text label + bar)

Four bands, not three — a 100%-full room reads differently from an 80%-full one when deciding whether to walk over.

| Band | Range | Token pair | th / en |
|---|---|---|---|
| low | 0–39% | `--color-density-low-{bg,fg}` (green) | ว่าง / Free |
| medium | 40–69% | `--color-density-medium-{bg,fg}` (amber) | ปานกลาง / Moderate |
| high | 70–94% | `--color-density-high-{bg,fg}` (orange) | แน่น / Busy |
| full | 95–100% | `--color-density-full-{bg,fg}` (red) | เต็ม / Full |

Every occupancy display shows: dot glyph (already existed) + `%` + band label + a horizontal capacity bar + free-seat count. Removing any one still leaves the other three, so color blindness / grayscale printing never loses information.

## Type

Single font, `next/font/google` **Anuphan** (variable weight 100–700, `subsets: ["thai", "latin"]`) — one typeface with native Thai shapes and full Latin coverage, so Thai and English sit at the same visual weight instead of a Thai fallback next to a Latin webfont. Replaces Geist (which was loaded but never actually applied to `body` — a pre-existing bug). Kept Geist Mono only for tabular/reference digits (countdown timer, booking reference).

Scale: Tailwind defaults only (`text-xs` … `text-2xl`), no custom scale — the app has 2 heading levels and body/caption, which Tailwind's stock scale already covers.

## Spacing / radius / elevation

Tailwind's default spacing and radius scale, used consistently:
- Cards: `rounded-xl`, `border`, `shadow-sm` (light) — enough separation from the new warm-white bg without inventing an elevation system.
- Controls (inputs, buttons): `rounded-lg`.
- Pills (badges, chips): `rounded-full`.

## Focus / motion

- Focus ring unchanged in mechanism (`:focus-visible { outline: 2px solid var(--color-brand) }`) but now sits on an AA-safe brand color in both themes.
- `prefers-reduced-motion: reduce` globally zeroes animation/transition duration — covers the new skeleton pulse and existing hover transitions.

## Components touched (existing files, no new ones except `Skeleton.tsx`)

`Button`, `Card`, `Chip`, `OccupancyBadge` (now also renders the capacity bar + free-seat count), `AppShell`, `LibrariesList`, `ReservationsList`, `AdminDashboard`, `ReserveForm`, `ReviewForm`, `StarRating`, login/detail pages. One new primitive, `Skeleton.tsx`, added because loading placeholders were duplicated ad hoc in 3+ places — same justification as the existing `Card`/`Chip`/`Button` primitives.
