/* Optional production-build checks. Counter requests are mocked, never counted.
   Build with SITE_URL=https://example.com VISITOR_BADGE=true, then serve it.
   SITE_URL here is the local preview URL, including any base path. */
const assert = require("node:assert/strict");
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || "playwright");
const base = (process.env.SITE_URL || "http://localhost:8097").replace(/\/$/, "");
(async () => {
  const browser = await chromium.launch({ channel: "chromium", headless: true });
  try {
    const context = await browser.newContext();
    const requests = [];
    await context.route("https://hits.sh/**", route => {
      requests.push(route.request().url());
      return route.fulfill({ contentType: "image/svg+xml", body: '<svg xmlns="http://www.w3.org/2000/svg" width="88" height="20"><rect width="88" height="20" fill="#b8521e"/><text x="5" y="14" fill="white" font-size="11">visits | 123</text></svg>' });
    });
    const page = await context.newPage();
    for (const width of [320, 390, 768, 1440]) {
      await page.setViewportSize({ width, height: 900 });
      for (const theme of ["light", "dark"]) {
        await page.emulateMedia({ colorScheme: theme });
        await page.goto(base + "/");
        await page.locator(".visitor-badge img").waitFor({ state: "visible" });
        assert.ok(await page.locator(".visitor-badge img").evaluate(el => el.complete && el.naturalWidth > 0));
        assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), `Overflow at ${width}px, ${theme}`);
        for (const selector of [".brand", ".visitor-badge", ".preferences summary"]) {
          const box = await page.locator(selector).boundingBox();
          assert.ok(box && box.x >= 0 && box.x + box.width <= width, `${selector} clipped at ${width}px`);
        }
      }
    }
    const namespace = await page.locator(".visitor-badge img").getAttribute("src");
    await page.goto(base + "/chapter/timers-pwm/");
    assert.equal(await page.locator(".visitor-badge img").getAttribute("src"), namespace);
    assert.match(await page.title(), /Timers.*PWM/);
    assert.equal(await page.locator('link[rel="canonical"]').count(), 1);
    const schema = JSON.parse(await page.locator('script[type="application/ld+json"]').textContent());
    assert.equal(schema["@type"], "LearningResource");
    await context.unroute("https://hits.sh/**");
    await context.route("https://hits.sh/**", route => route.abort());
    await page.goto(base + "/");
    await page.locator(".visitor-badge-fallback").waitFor({ state: "visible" });
    assert.equal(await page.locator(".visitor-badge img").isVisible(), false);
    assert.ok(requests.length > 0);
    const noJS = await browser.newContext({ javaScriptEnabled: false });
    await noJS.route("https://hits.sh/**", route => route.fulfill({ contentType: "image/svg+xml", body: '<svg xmlns="http://www.w3.org/2000/svg" width="88" height="20"/>' }));
    const staticPage = await noJS.newPage();
    await staticPage.goto(base + "/");
    assert.equal(await staticPage.locator(".visitor-badge img").isVisible(), true);
    assert.equal(await staticPage.locator("h1").count(), 1);
    console.log("SEO and visitor-badge browser checks passed (counter mocked).");
  } finally {
    await browser.close();
  }
})().catch(error => { console.error(error); process.exitCode = 1; });
