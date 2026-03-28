/**
 * Playwright script to record a web UI demo GIF that mirrors the TUI demo.
 *
 * Flow: sidebar search → select deploy-app → fill parameter form →
 *       risk confirmation → watch execution output → back to runbook.
 *
 * Usage:
 *   npx playwright test tapes/web-demo.ts
 *   # or via the wrapper:
 *   ./tapes/record-web-demo.sh
 */
import { test, expect, type Page } from "@playwright/test";

// Base URL is set in playwright.config.ts — page.goto("/") uses it.

// Slow typing speed to look natural in the recording.
async function typeSlowly(page: Page, selector: string, text: string, delay = 80) {
  await page.click(selector);
  for (const char of text) {
    await page.keyboard.type(char, { delay });
  }
}

test("web UI demo", async ({ page }) => {
  // ── 1. Dashboard ──────────────────────────────────────────────────
  await page.goto("/");
  await page.waitForSelector("text=dops");
  await page.waitForTimeout(2500);

  // ── 2. Search sidebar ─────────────────────────────────────────────
  const searchInput = page.locator('aside input[type="text"]');
  await searchInput.click();
  await page.waitForTimeout(500);
  await searchInput.pressSequentially("deploy-app", { delay: 120 });
  await page.waitForTimeout(1500);

  // ── 3. Select deploy-app ──────────────────────────────────────────
  await page.locator("button", { hasText: "deploy-app" }).first().click();
  await page.waitForTimeout(2000);

  // ── 4. Fill parameter form ────────────────────────────────────────

  // Environment: select "staging" (should be default, but click to show)
  const envSelect = page.locator("select").first();
  await envSelect.selectOption("staging");
  await page.waitForTimeout(1000);

  // Version: type "v1.2.3"
  const versionInput = page.locator('input[placeholder="version"]');
  await versionInput.click();
  await versionInput.pressSequentially("v1.2.3", { delay: 120 });
  await page.waitForTimeout(1200);

  // Multi-select features: click logging, monitoring, tracing
  // These are chip buttons inside the main form (not the saved section).
  const mainForm = page.locator("form");
  await mainForm.locator("button", { hasText: "logging" }).click();
  await page.waitForTimeout(500);
  await mainForm.locator("button", { hasText: "monitoring" }).click();
  await page.waitForTimeout(500);
  await mainForm.locator("button", { hasText: "tracing" }).click();
  await page.waitForTimeout(1000);

  // ── 4b. Expand saved values section ─────────────────────────────
  // api_token and dry_run have saved values — show them for the demo.
  const savedToggle = page.locator("button", { hasText: "Saved values" });
  if (await savedToggle.isVisible()) {
    await savedToggle.click();
    await page.waitForTimeout(2000);
  }

  // ── 5. Scroll down to show Execute button ─────────────────────────
  await page.locator("button", { hasText: "Execute" }).first().scrollIntoViewIfNeeded();
  await page.waitForTimeout(1200);

  // ── 6. Click Execute → risk confirmation dialog ───────────────────
  await page.locator("button", { hasText: "Execute" }).first().click();
  await page.waitForTimeout(2500);

  // ── 7. Confirm execution (high risk = click Execute in dialog) ────
  // The confirmation dialog should be visible now.
  const confirmBtn = page.locator(".fixed button", { hasText: "Execute" });
  await confirmBtn.click();
  await page.waitForTimeout(2000);

  // ── 8. Watch execution output ─────────────────────────────────────
  // Wait for at least a few lines of output.
  await page.waitForSelector(".font-mono span >> nth=5", { timeout: 15000 }).catch(() => {});
  await page.waitForTimeout(5000);

  // ── 9. Wait for completion ────────────────────────────────────────
  // Wait for the "Completed" or "Failed" status pill (or timeout).
  await page
    .locator("text=Completed")
    .or(page.locator("text=Failed"))
    .first()
    .waitFor({ timeout: 30000 })
    .catch(() => {});
  await page.waitForTimeout(3000);

  // ── 10. Click back to runbook ─────────────────────────────────────
  const backBtn = page.locator("button", { hasText: "Back to runbook" });
  if (await backBtn.isVisible()) {
    await backBtn.click();
    await page.waitForTimeout(2000);
  }
});
