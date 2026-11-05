// UI NFRs (Proposal: Operational, Usability, Cultural) against a running app:
//   `make up` (or pnpm dev with the backend running), then  pnpm test:nfr
//   - mobile: every page fits a 375 px phone screen, no horizontal scroll
//   - Thai UI by default, English via the EN toggle
//   - steps: user actions from sign-in to a confirmed reservation (recorded)
// Screenshots go to e2e/screenshots/. Exits 1 if any check fails.
import { chromium } from "playwright";
import { mkdir } from "node:fs/promises";

const APP = process.env.FRONTEND_URL ?? "http://localhost:3000";
const API = process.env.GATEWAY_URL ?? "http://localhost:8080";
const PHONE = { viewport: { width: 375, height: 812 }, isMobile: true, hasTouch: true };
const THAI = /[฀-๿]/;
const SHOTS = new URL("./screenshots/", import.meta.url).pathname;

let failed = 0;
function check(ok, msg) {
  console.log(`${ok ? "PASS" : "FAIL"} ${msg}`);
  if (!ok) failed++;
}

const api = async (path, body) => {
  const res = await fetch(API + path, body && { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(body) });
  if (!res.ok) throw new Error(`${path}: ${res.status}`);
  return res.json();
};

async function fitsPhone(page, path, name) {
  await page.goto(APP + path, { waitUntil: "networkidle" });
  const { scroll, width } = await page.evaluate(() => ({ scroll: document.documentElement.scrollWidth, width: window.innerWidth }));
  await page.screenshot({ path: `${SHOTS}${name}.png`, fullPage: true });
  check(scroll <= width, `mobile: ${path} fits ${width}px (content ${scroll}px)`);
}

await mkdir(SHOTS, { recursive: true });
const [lib] = await api("/api/libraries");
const [room] = await api(`/api/libraries/${lib.id}/rooms`);
const { token: adminToken } = await api("/api/auth/admin/login", { email: "admin@cu.ac.th", password: "admin123" });

const browser = await chromium.launch();
try {
  // --- Cultural: Thai by default, English optional.
  {
    const page = await (await browser.newContext(PHONE)).newPage();
    await page.goto(APP + "/", { waitUntil: "networkidle" });
    check((await page.getAttribute("html", "lang")) === "th", 'cultural: default <html lang="th">');
    check(THAI.test(await page.locator("nav").first().innerText()), "cultural: navigation is in Thai by default");
    await Promise.all([page.waitForEvent("load"), page.getByRole("button", { name: "Switch to English" }).click()]);
    check((await page.getAttribute("html", "lang")) === "en", 'cultural: EN toggle switches to <html lang="en">');
    check(!THAI.test(await page.locator("nav").first().innerText()), "cultural: navigation is in English after toggle");
  }

  // --- Usability: count user actions from sign-in to a confirmed reservation.
  const ctx = await browser.newContext(PHONE);
  const page = await ctx.newPage();
  let actions = 0;
  const act = async (fn) => {
    actions++;
    await fn();
  };
  await page.goto(APP + "/login", { waitUntil: "networkidle" });
  await act(() => page.locator("#panel-cu input").first().fill(`ux${Date.now()}`));
  await act(() => page.locator('#panel-cu button[type="submit"]').click());
  await page.waitForURL((u) => !u.pathname.startsWith("/login"));
  const landing = await page.locator("main").innerText();
  check(/\d+\s*%/.test(landing), "usability: occupancy is visible on the landing page with 0 extra clicks");
  await act(() => page.locator(`a[href="/libraries/${lib.id}"]`).first().click());
  await act(() => page.locator(`a[href="/reserve/${room.id}"]`).first().click());
  // A random future slot so repeated runs don't collide.
  const d = new Date(Date.now() + (10 + Math.floor(Math.random() * 60)) * 86400e3);
  d.setHours(9 + Math.floor(Math.random() * 8), 0, 0, 0);
  const local = new Date(d.getTime() - d.getTimezoneOffset() * 60e3).toISOString().slice(0, 16);
  await act(() => page.locator('input[type="datetime-local"]').fill(local));
  await act(() => page.locator('form button[type="submit"]').click());
  const created = page.waitForResponse((r) => r.request().method() === "POST" && r.url().endsWith("/api/reservations"));
  await act(() => page.locator("main button").last().click()); // confirm
  const status = (await created).status();
  check(status === 201, `usability: reservation confirmed through the UI (POST /api/reservations -> ${status})`);
  console.log(`INFO usability: ${actions} user actions from sign-in to confirmed reservation`);

  // --- Operational: mobile display, signed in as a customer, then as an admin.
  await fitsPhone(page, "/", "home");
  await fitsPhone(page, `/libraries/${lib.id}`, "library");
  await fitsPhone(page, `/reserve/${room.id}`, "reserve");
  await fitsPhone(page, "/me/reservations", "my-reservations");
  const anon = await (await browser.newContext(PHONE)).newPage();
  await fitsPhone(anon, "/login", "login");
  const adminCtx = await browser.newContext(PHONE);
  await adminCtx.addCookies([{ name: "cusf_token", value: adminToken, url: APP }]);
  await fitsPhone(await adminCtx.newPage(), "/admin", "admin");
} finally {
  await browser.close();
}
process.exit(failed ? 1 : 0);
