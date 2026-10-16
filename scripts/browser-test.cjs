/* Optional browser regression checks. See README for running with Playwright. */
const assert = require("node:assert/strict");
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || "playwright");
const base = (process.env.SITE_URL || "http://localhost:8097").replace(/\/$/, "");
(async () => {
  const browser = await chromium.launch({ channel: "chromium", headless: true });
  try {
    const context = await browser.newContext({ viewport: { width: 390, height: 844 }, reducedMotion: "reduce" });
    await context.addInitScript(() => {
      window.__interfaceTones = 0;
      window.__feedbackTypes = [];
      document.addEventListener("guide:feedback", event => window.__feedbackTypes.push(event.detail?.type));
      const Audio = window.AudioContext || window.webkitAudioContext;
      if (Audio) {
        const original = Audio.prototype.createOscillator;
        Audio.prototype.createOscillator = function (...args) { window.__interfaceTones++; return original.apply(this, args); };
      }
    });
    const page = await context.newPage(), errors = [];
    page.on("pageerror", e => errors.push(e.message));
    const ready = () => page.waitForSelector('html[data-storage-ready="true"]');
    const saved = () => page.waitForSelector('html[data-saving="false"]');
    await page.goto(base + "/"); await ready();
    const chapterTotal = await page.locator(".index-chapter").count();
    assert.ok(chapterTotal >= 19, "Beginner and practical firmware lessons must appear on the learning map");
    assert.equal(await page.locator(".sidebar-part[open]").count(), 0, "Overview categories should start collapsed");
    assert.ok((await page.locator(".compact-map").boundingBox()).height < 300, "Mobile learning map should stay compact");
    assert.equal(await page.locator(".guide-sidebar").evaluate(el => el.inert), true);
    await page.locator(".drawer-trigger").click();
    assert.equal(await page.locator("#guide-sidebar").getAttribute("aria-modal"), "true");
    assert.equal(await page.locator("#main-content").evaluate(el => el.inert), true);
    await page.locator(".sidebar-part > summary").first().click();
    assert.equal(await page.locator(".sidebar-part").first().getAttribute("open"), "");
    await page.keyboard.press("Shift+Tab");
    assert.equal(await page.locator("#guide-sidebar").evaluate(el => el.contains(document.activeElement)), true);
    if (process.env.SCREENSHOT_DIR) await page.screenshot({ path: `${process.env.SCREENSHOT_DIR}/mobile-drawer.png` });
    await page.keyboard.press("Escape");
    assert.equal(await page.locator(".drawer-trigger").getAttribute("aria-expanded"), "false");
    assert.equal(await page.evaluate(() => window.__interfaceTones), 0, "Sounds must default off");
    await page.locator(".preferences summary").click();
    await page.selectOption("#theme-choice", "dark");
    await page.check("#sound-choice");
    await saved();
    assert.equal(await page.locator("html").getAttribute("data-theme"), "dark");
    await page.keyboard.press("Escape");
    await page.locator(".experiment-preview > summary").click();
    await page.locator('[data-answer="incorrect"]').first().click();
    assert.match(await page.locator(".answer-feedback").textContent(), /Not quite/);
    await page.locator('[data-answer="correct"]').click();
    assert.match(await page.locator(".answer-feedback").textContent(), /Exactly/);
    assert.ok(await page.evaluate(() => window.__interfaceTones > 0), "Opted-in feedback should synthesize sound");
    await page.goto(base + "/chapter/uart/"); await ready();
    assert.equal(await page.locator(".sidebar-part[open]").count(), 1, "Only the active chapter category should start open");
    assert.equal(await page.locator("html").getAttribute("data-theme"), "dark");
    const toggles = page.locator(".section-read-toggle");
    for (let i = 0; i < await toggles.count(); i++) {
      if (await toggles.nth(i).getAttribute("aria-pressed") !== "true") await toggles.nth(i).click();
    }
    await saved();
    await page.reload(); await ready();
    assert.match(await page.locator("[data-reading-summary]").textContent(), /^3 \/ 3/);
    if (process.env.SCREENSHOT_DIR) await page.screenshot({ path: `${process.env.SCREENSHOT_DIR}/uart-dark-mobile.png`, fullPage: true });
    await page.locator(".sim-play").click();
    await page.waitForTimeout(700);
    await page.locator(".sim-play").click();
    const position = await page.locator(".simulation-readout").textContent();
    await page.waitForTimeout(650);
    assert.equal(await page.locator(".simulation-readout").textContent(), position, "Pause must freeze the simulation");
    await page.locator(".sim-reset").click();
    assert.match(await page.locator(".simulation-readout").textContent(), /^Start bit/);
    await page.locator(".sim-step").click();
    assert.match(await page.locator(".simulation-readout").textContent(), /^Data bit 0/);
    await page.goto(base + "/"); await ready();
    assert.equal(await page.locator('.index-chapter[data-chapter-link="uart"]').getAttribute("data-read-state"), "read");
    assert.ok((await page.locator(".overall-progress").textContent()).startsWith(`1 / ${chapterTotal}`));
    await page.locator(".preferences summary").click();
    assert.equal(await page.isChecked("#sound-choice"), true);
    await page.uncheck("#tracking-choice");
    await saved();
    await page.goto(base + "/chapter/clocks/"); await ready();
    assert.equal(await page.locator(".section-read-toggle").first().isDisabled(), true);
    await page.locator(".sim-step").click();
    assert.match(await page.locator(".simulation-readout").textContent(), /Tick 2/);
    await page.locator(".sim-play").click();
    await page.waitForTimeout(220);
    await page.locator(".sim-play").click();
    assert.equal(await page.locator(".sim-play").getAttribute("aria-pressed"), "false");
    await page.locator(".preferences summary").click();
    await page.check("#tracking-choice");
    await page.click("#reset-progress");
    await page.waitForTimeout(100);
    await page.goto(base + "/"); await ready();
    assert.ok((await page.locator(".overall-progress").textContent()).startsWith(`0 / ${chapterTotal}`));
    assert.equal(await page.locator("html").getAttribute("data-theme"), "dark");
    await page.goto(base + "/chapter/uart/"); await ready();
    await page.locator(".section-read-toggle").first().click(); await saved();
    await Promise.all([page.waitForNavigation(), page.keyboard.press("Control+Shift+R")]);
    await ready();
    assert.match(await page.locator("[data-reading-summary]").textContent(), /^0 \/ 3/);
    assert.equal(await page.locator("html").getAttribute("data-theme"), "dark");
    await page.goto(base + "/"); await ready();
    await page.setViewportSize({ width: 1440, height: 1000 });
    await page.waitForFunction(() => !document.querySelector(".guide-sidebar").inert);
    assert.ok((await page.locator(".compact-map").boundingBox()).height < 200, "Desktop learning map should fit in a short strip");
    assert.equal(await page.locator(".guide-sidebar").evaluate(el => el.inert), false);
    for (const theme of ["light", "dark"]) {
      await page.locator(".preferences summary").click();
      await page.selectOption("#theme-choice", theme);
      await page.keyboard.press("Escape");
      if (process.env.SCREENSHOT_DIR) await page.screenshot({ path: `${process.env.SCREENSHOT_DIR}/home-${theme}.png`, fullPage: true });
    }
    const slugs = await page.locator(".index-chapter").evaluateAll(nodes => nodes.map(el => el.dataset.chapterLink));
    for (const width of [360, 768, 1440]) {
      await page.setViewportSize({ width, height: 900 });
      for (const slug of slugs) {
        await page.goto(base + "/chapter/" + slug + "/"); await ready();
        assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth), false, `${slug} overflows at ${width}px`);
        assert.match(await page.locator(".chapter-learning-time").textContent(), /Estimated learning time: ~\d+ min/);
        assert.equal(await page.locator(".sidebar-duration").count(), chapterTotal, "Every sidebar chapter needs a learning estimate");
        assert.equal(await page.locator('footer a[href="https://x.com/chotathanos"]').count(), 1, "Every chapter should include Jay's shared footer link");
        assert.equal(await page.locator(".chapter-checkpoint details").count(), 1, "Every chapter needs a review explanation");
        if (["boot-sequence", "interrupts", "timers-pwm"].includes(slug)) {
          await page.locator(".sim-step").click();
          assert.ok((await page.locator(".simulation-readout").textContent()).length > 0);
        }
        if (width === 360 && slug === "electronics-essentials") {
          assert.match(await page.locator("#electronics-estimate").textContent(), /3\.94 mA/);
          await page.fill("#electronics-resistance", "660");
          assert.match(await page.locator("#electronics-estimate").textContent(), /1\.97 mA/);
          await page.fill("#electronics-drop", "4");
          assert.match(await page.locator("#electronics-estimate").textContent(), /No forward-current estimate/);
          await page.fill("#electronics-drop", "2");
          await page.selectOption("#electronics-prediction", "half");
          await page.getByRole("button", { name: "Check prediction", exact: true }).click();
          assert.match(await page.locator('.interactive p[role="status"]').last().textContent(), /Correct/);
          assert.ok((await page.evaluate(() => window.__feedbackTypes)).includes("correct"));
        }
        if (width === 360 && slug === "embedded-c") {
          await page.selectOption("#embedded-c-bit", "31");
          assert.equal(await page.locator("output").textContent(), "0x80000000");
          await page.selectOption("#embedded-c-answer", "invalid");
          await page.getByRole("button", { name: "Check answer", exact: true }).click();
          assert.ok((await page.evaluate(() => window.__feedbackTypes)).includes("correct"));
        }
        if (width === 360 && slug === "bare-metal-firmware") {
          assert.equal(await page.locator("pre").count(), 3, "C, Rust, and Zig examples must all be present");
          await page.getByRole("button", { name: "Next step", exact: true }).click();
          assert.match(await page.locator("#bare-metal-step").textContent(), /Step 2/);
          await page.getByRole("button", { name: "Restart loop", exact: true }).click();
          assert.match(await page.locator("#bare-metal-step").textContent(), /Step 1/);
        }
        if (width === 360 && slug === "debouncing") {
          const next = page.getByRole("button", { name: "Next sample", exact: true });
          const status = page.locator('.interactive [role="status"]');
          for (let sample = 1; sample <= 16; sample++) {
            await next.click();
            if (sample === 8) assert.match(await status.textContent(), /Accepted released · Press events 0/);
            if (sample === 9) assert.match(await status.textContent(), /Accepted pressed · Press events 1/);
          }
          assert.match(await status.textContent(), /Accepted released · Press events 1/);
          assert.equal(await next.isDisabled(), true);
          await page.getByRole("button", { name: "Reset samples", exact: true }).click();
          assert.match(await status.textContent(), /Sample 0 \/ 16/);
          assert.equal(await next.isDisabled(), false);
        }
        if (width === 360 && slug === "glossary") {
          await page.locator('.glossary-jumps a[href="#atomicity"]').click();
          assert.ok(page.url().endsWith("#atomicity"));
          assert.ok(await page.locator("#atomicity").isVisible());
          const review = page.locator(".chapter-checkpoint details");
          await review.locator("summary").click();
          assert.equal(await review.getAttribute("open"), "");
        }
      }
    }
    assert.deepEqual(errors, [], "No browser/Alpine errors expected");
    const persisted = await page.evaluate(() => new Promise((resolve, reject) => {
      const request = indexedDB.open("embedded-visual-guide");
      request.onsuccess = () => { const db = request.result; const tx = db.transaction("state"); const read = tx.objectStore("state").get("preferences"); tx.oncomplete = () => { resolve(read.result.value); db.close(); }; tx.onerror = reject; };
    }));
    assert.equal(persisted.theme, "dark"); assert.equal(persisted.sound, true);
    const automatic = await browser.newContext({ viewport: { width: 390, height: 844 }, reducedMotion: "reduce" });
    const reading = await automatic.newPage(); await reading.goto(base + "/chapter/what-is-embedded/");
    await reading.waitForSelector('html[data-storage-ready="true"]');
    await reading.locator(".section-read-toggle").first().scrollIntoViewIfNeeded();
    await reading.evaluate(() => scrollBy(0, 70));
    await reading.waitForFunction(() => document.querySelector(".section-read-toggle").getAttribute("aria-pressed") === "true");
    assert.equal(await reading.locator(".section-read-toggle").nth(1).getAttribute("aria-pressed"), "false", "Unseen section must stay unread");
    await automatic.close();
    const noStorage = await browser.newContext({ viewport: { width: 390, height: 844 } });
    await noStorage.addInitScript(() => { Object.defineProperty(window, "indexedDB", { get() { throw new Error("Storage blocked"); } }); });
    const fallback = await noStorage.newPage();
    await fallback.goto(base + "/"); await fallback.waitForSelector('html[data-storage-ready="true"]');
    assert.match(await fallback.locator("#storage-status").textContent(), /unavailable/);
    await fallback.locator(".preferences summary").click(); await fallback.selectOption("#theme-choice", "dark");
    assert.equal(await fallback.locator("html").getAttribute("data-theme"), "dark");
    await noStorage.close();
    const noJS = await browser.newContext({ javaScriptEnabled: false, viewport: { width: 390, height: 844 } });
    const staticPage = await noJS.newPage(); await staticPage.goto(base + "/chapter/uart/");
    assert.equal(await staticPage.locator(".sidebar-link").count(), chapterTotal);
    assert.equal(await staticPage.locator("article section").first().evaluate(el => getComputedStyle(el).opacity), "1");
    await staticPage.locator(".chapter-checkpoint summary").click();
    assert.equal(await staticPage.locator(".chapter-checkpoint details").getAttribute("open"), "", "Review explanations should work without JavaScript");
    await staticPage.goto(base + "/chapter/debouncing/");
    assert.match(await staticPage.locator("noscript").textContent(), /exactly one accepted press/);
    await noJS.close();
    console.log("Browser checks passed: drawer, themes, IndexedDB, progress, sound preference, feedback, playback, all chapters at 360/768/1440px.");
    await context.close();
  } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; });
