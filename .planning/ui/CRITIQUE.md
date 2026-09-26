# UI Critique — CU SeatFinder frontend (pre-redesign)

Grounded in full read of `src/app` + `src/components`, and live screenshots (localhost:3000, customer session) at desktop viewport (browser resize to 390 did not change the captured viewport, so mobile-specific layout wasn't visually confirmed — the flex/`md:` breakpoints were verified from code instead).

## Priority findings

| # | Finding | Severity | Fix direction |
|---|---|---|---|
| 1 | UC-01's core answer (how full is each library) is a small badge with `%` + word + dot glyphs; the actual seat numbers (`69 / 800`) are muted, tiny gray text and there's no visual bar. Nothing states **free seats**, which is what a student actually needs before walking over. | 🔴 Critical | Add a capacity bar + explicit "X ที่นั่งว่าง" free-seat count to every list row and the detail header. |
| 2 | Filter card (two full-width range sliders + amenity chips) renders open and above the fold on mobile, pushing the actual library list below the first screen — directly hurts "minimize steps to check occupancy" (NFR). | 🔴 Critical | Collapse filters behind a native `<details>` disclosure, compact summary line shows active filter count; list is visible immediately. |
| 3 | Body text falls back to hard-coded `Arial, Helvetica, sans-serif` in `globals.css` — the loaded Geist variable font is never actually applied to `body`, and there is no Thai-optimized typeface despite Thai being the default language (NFR: Thai default). | 🟡 Moderate | Load a Google font with real Thai support via `next/font`, apply as `--font-sans`, drop the dead Geist variables. |
| 4 | Primary touch targets (buttons, chips) are ~32–36px tall (`px-4 py-2 text-sm`, `px-3 py-1.5`) — under the 44px minimum for primary actions (book, confirm, submit, accept/reject). | 🟡 Moderate | Bump `Button` base to `min-h-11`; bump bottom-nav tap area similarly. |
| 5 | Density badge only has 3 bands (low/medium/high); a 100%-full room and an 80%-full room read identically as "high", which matters most exactly when a student is deciding whether to bother going. | 🟡 Moderate | Add a 4th "full" band with its own token/copy, still paired with text + bar (never color-only). |
| 6 | "My reservations" QR/ticket state is a plain `Card` with a small QR and a line of text — doesn't read as a ticket, and the countdown is visually equal-weight to everything else despite being the most time-sensitive info on that screen (UC-02 postcondition: "booking confirmation ticket... with check-in countdown"). | 🟡 Moderate | Redesign the accepted-reservation card as a boarding-pass-style ticket: dashed stub divider, large reference code, prominent countdown. |
| 7 | Client-side loading states are bare `"กำลังโหลด..."` text (reservations list, admin dashboard, reserve-room card) even though initial paint elsewhere is server-rendered — inconsistent and produces a flash of unstyled text. | 🟢 Minor | Swap for lightweight skeleton blocks (respecting `prefers-reduced-motion`). |
| 8 | Cards on dark background use only a 1px neutral border against near-black — very low separation/elevation; palette reads as generic gray, not tied to the CU brand beyond link/button color. | 🟢 Minor | Warm stone-gray surfaces + slightly stronger elevation; brand pink used sparingly as accent (already mostly true — keep, refine). |
| 9 | Admin dashboard uses full `Card` padding per row for what's fundamentally a list to triage quickly — inefficient use of vertical space for an admin power-user flow. | 🟢 Minor | Convert to a tighter divided-row list; keep accept/reject inline. |
| 10 | Nav label for Home is "หอสมุด" (="the library/libraries") rather than something that reads as "home" — minor copy nit, not wrong, revisit in copy pass. | 🟢 Minor | ux-copy pass. |

## What already works well (keep)

- Occupancy list is already sorted least→most crowded server-side per API contract — satisfies UC-03 ranking without extra client logic.
- Density is already never color-only (dot glyphs + text label alongside color) — good accessibility instinct, just needs a bar and a 4th band.
- Solid accessibility bones already in place: skip-to-content link, `aria-current`, roving-tabindex tablist on login, `role="radiogroup"` star input, `aria-busy` on buttons, `:focus-visible` ring. Redesign must not regress any of this.
- i18n is a clean flat dictionary with full th/en parity — easy to extend.
- Reserve flow is already minimal (room → time/duration → confirm → done), matches "fewest steps" NFR structurally; only needs a visual pass.
