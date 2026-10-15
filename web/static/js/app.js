/* Static-first enhancements. IndexedDB owns preferences and section progress. */
(function () {
  "use strict";
  const $ = (s, root = document) => root.querySelector(s);
  const $$ = (s, root = document) => [...root.querySelectorAll(s)];
  const defaults = { theme: "system", sound: false, tracking: true };
  let preferences = { ...defaults }, progress = {}, db = null, audio = null;
  let storageReady = false, channel = null, lastSound = 0, resetUntil = 0;
  const motion = matchMedia("(prefers-reduced-motion: reduce)");
  const systemTheme = matchMedia("(prefers-color-scheme: dark)");
  const article = $("article[data-chapter]");
  const sections = article ? $$("section", article) : [];
  let toastTimer;

  function toast(message) {
    const el = $("#app-toast");
    if (!el) return;
    el.textContent = message;
    el.classList.add("visible");
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => el.classList.remove("visible"), 2800);
  }
  function storageFailure() {
    const status = $("#storage-status");
    if (status) status.textContent = "Storage unavailable. Changes last only for this page.";
  }
  function openDB() {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open("embedded-visual-guide", 1);
      const timer = setTimeout(() => reject(new Error("Storage timeout")), 1800);
      request.onupgradeneeded = () => request.result.createObjectStore("state", { keyPath: "key" });
      request.onerror = () => { clearTimeout(timer); reject(request.error); };
      request.onblocked = () => { clearTimeout(timer); reject(new Error("Storage blocked")); };
      request.onsuccess = () => {
        clearTimeout(timer);
        resolve(request.result);
      };
    });
  }
  function records() {
    return new Promise((resolve, reject) => {
      const tx = db.transaction("state", "readonly");
      const request = tx.objectStore("state").getAll();
      tx.oncomplete = () => resolve(request.result);
      tx.onabort = tx.onerror = () => reject(tx.error);
    });
  }
  function persist(key, value) {
    if (!db) return Promise.resolve();
    return new Promise((resolve, reject) => {
      const tx = db.transaction("state", "readwrite");
      tx.objectStore("state").put({ key, value });
      tx.oncomplete = resolve;
      tx.onabort = tx.onerror = () => reject(tx.error);
    }).catch(() => { storageFailure(); });
  }
  function broadcast(data) { if (channel) channel.postMessage(data); }
  function validPreferences(value) {
    return {
      theme: ["system", "light", "dark"].includes(value?.theme) ? value.theme : "system",
      sound: value?.sound === true,
      tracking: value?.tracking !== false
    };
  }
  function applyPreferences() {
    const theme = preferences.theme === "system" ? (systemTheme.matches ? "dark" : "light") : preferences.theme;
    document.documentElement.dataset.theme = theme;
    try {
      if (preferences.theme === "system") localStorage.removeItem("embedded-guide-theme");
      else localStorage.setItem("embedded-guide-theme", preferences.theme);
    } catch (_) { /* IndexedDB remains authoritative. */ }
    if ($("#theme-choice")) $("#theme-choice").value = preferences.theme;
    if ($("#sound-choice")) $("#sound-choice").checked = preferences.sound;
    if ($("#tracking-choice")) $("#tracking-choice").checked = preferences.tracking;
    document.documentElement.classList.toggle("tracking-paused", !preferences.tracking);
    renderProgress();
  }

  // All tones are synthesized locally; playback requires an explicit opt-in
  // and an initial gesture. Never start audio from scrolling alone.
  function sound(kind) {
    if (!preferences.sound || !audio || audio.state !== "running") return;
    const now = performance.now();
    if (now - lastSound < (kind === "scroll" ? 350 : 55)) return;
    lastSound = now;
    const notes = { click: [480], scroll: [190], correct: [523, 659, 784], incorrect: [260, 196] }[kind] || [480];
    notes.forEach((frequency, i) => {
      const start = audio.currentTime + i * .075;
      const oscillator = audio.createOscillator(), gain = audio.createGain();
      oscillator.type = "sine";
      oscillator.frequency.value = frequency;
      gain.gain.setValueAtTime(0, start);
      gain.gain.linearRampToValueAtTime(kind === "scroll" ? .012 : .045, start + .008);
      gain.gain.exponentialRampToValueAtTime(.0001, start + .09);
      oscillator.connect(gain).connect(audio.destination);
      oscillator.start(start); oscillator.stop(start + .1);
    });
  }
  function unlockAudio() {
    if (!preferences.sound) return;
    try {
      const Context = window.AudioContext || window.webkitAudioContext;
      if (!Context) return;
      audio ||= new Context();
      if (audio.state === "suspended") audio.resume().catch(() => {});
    } catch (_) { /* Audio is optional. */ }
  }

  function initDrawer() {
    const sidebar = $("#guide-sidebar"), trigger = $(".drawer-trigger"), backdrop = $(".drawer-backdrop");
    if (!sidebar || !trigger || !backdrop) return;
    const desktop = matchMedia("(min-width: 1080px)"), menu = $(".chapter-menu");
    const main = $("#main-content"), topbar = $(".topbar"), footer = $("footer");
    const others = [main, topbar, footer].filter(Boolean);
    let opened = false, previousFocus;
    function close(restore = true) {
      opened = false;
      document.body.classList.remove("drawer-open");
      trigger.setAttribute("aria-expanded", "false");
      sidebar.removeAttribute("role"); sidebar.removeAttribute("aria-modal");
      backdrop.hidden = true;
      others.forEach(el => { el.inert = false; });
      sidebar.inert = !desktop.matches;
      if (restore && previousFocus?.isConnected) previousFocus.focus();
    }
    function open() {
      previousFocus = document.activeElement;
      opened = true; menu.open = true; sidebar.inert = false;
      document.body.classList.add("drawer-open");
      sidebar.setAttribute("role", "dialog"); sidebar.setAttribute("aria-modal", "true");
      trigger.setAttribute("aria-expanded", "true"); backdrop.hidden = false;
      others.forEach(el => { el.inert = true; });
      $(".drawer-close", sidebar).focus();
    }
    trigger.hidden = false;
    document.documentElement.classList.add("drawer-ready");
    menu.open = true;
    close(false);
    trigger.addEventListener("click", open);
    $(".drawer-close", sidebar).addEventListener("click", () => close());
    backdrop.addEventListener("click", () => close());
    desktop.addEventListener("change", () => close(false));
    sidebar.addEventListener("click", e => { if (e.target.closest("a")) close(false); });
    document.addEventListener("keydown", e => {
      if (!opened) return;
      if (e.key === "Escape") { e.preventDefault(); close(); }
      if (e.key === "Tab") {
        const focusable = $$("a[href], button:not([disabled]), select", sidebar);
        const first = focusable[0], last = focusable.at(-1);
        if (e.shiftKey && document.activeElement === first) { e.preventDefault(); last.focus(); }
        else if (!e.shiftKey && document.activeElement === last) { e.preventDefault(); first.focus(); }
      }
    });
  }

  function renderProgress() {
    let complete = 0;
    const seen = new Set();
    $$("[data-chapter-link]").forEach(link => {
      const slug = link.dataset.chapterLink, total = Number(link.dataset.sections);
      const read = new Set((progress[slug] || []).filter(n => Number.isInteger(n) && n > 0 && n <= total));
      const state = read.size >= total && total > 0 ? "read" : read.size ? "partial" : "unread";
      link.dataset.readState = state;
      if (!seen.has(slug) && state === "read") complete++;
      seen.add(slug);
      const label = state === "read" ? "Read" : state === "partial" ? "In progress" : "Not read";
      const status = $(".chapter-status", link);
      if (status) { status.textContent = state === "read" ? "✓" : state === "partial" ? "◐" : "○"; status.setAttribute("aria-label", label); }
      const text = $(".index-section-progress", link);
      if (text) text.textContent = `${read.size} / ${total} sections read · ${label}`;
      $$("[data-section-pip]", link).forEach(pip => pip.classList.toggle("read", read.has(Number(pip.dataset.sectionPip))));
    });
    const overall = $(".overall-progress");
    if (overall) overall.textContent = `${complete} / ${seen.size} chapters read${preferences.tracking ? "" : " · tracking paused"}`;
    if (article) {
      const read = new Set(progress[article.dataset.chapter] || []);
      sections.forEach((section, i) => {
        const button = $(".section-read-toggle", section), done = read.has(i + 1);
        section.classList.toggle("section-read", done);
        if (button) { button.textContent = done ? "✓ Read · mark unread" : "○ Mark section read"; button.setAttribute("aria-pressed", String(done)); button.disabled = !preferences.tracking; }
      });
      const summary = $("[data-reading-summary]");
      if (summary) summary.textContent = preferences.tracking ? `${read.size} / ${sections.length} sections read · saved on this browser` : "Reading tracking is paused in Settings.";
    }
  }
  async function markSection(number, done, audible = false) {
    if (!storageReady || !preferences.tracking || !article) return;
    const slug = article.dataset.chapter;
    const read = new Set(progress[slug] || []);
    if (done) read.add(number); else read.delete(number);
    progress[slug] = [...read].sort((a, b) => a - b);
    renderProgress();
    if (audible && done) sound("correct");
    broadcast({ type: "progress", slug, value: progress[slug] });
    await persist("progress/" + slug, progress[slug]);
  }
  const dwell = new Map(), manuallyUnread = new Set();
  function initReading() {
    sections.forEach((section, i) => {
      section.id ||= "section-" + (i + 1);
      const button = document.createElement("button");
      button.type = "button"; button.className = "section-read-toggle";
      button.addEventListener("click", () => {
        const done = !(progress[article.dataset.chapter] || []).includes(i + 1);
        if (!done) manuallyUnread.add(i + 1); else manuallyUnread.delete(i + 1);
        markSection(i + 1, done, true);
      });
      section.append(button);
    });
    // A section is auto-read only after 3 seconds in view AND reaching its end.
    setInterval(() => {
      if (document.hidden || !storageReady || !preferences.tracking || Date.now() < resetUntil) return;
      sections.forEach((section, i) => {
        if (manuallyUnread.has(i + 1) || (progress[article.dataset.chapter] || []).includes(i + 1)) return;
        const rect = section.getBoundingClientRect();
        if (rect.top < innerHeight - 30 && rect.bottom > 90) {
          const seconds = (dwell.get(i) || 0) + 1; dwell.set(i, seconds);
          if (seconds >= 3 && rect.bottom <= innerHeight + 12) markSection(i + 1, true);
        }
      });
    }, 1000);
    renderProgress();
  }
  async function resetProgress(reload = false) {
    // The synchronous marker survives even if the browser handles its reload
    // shortcut before an asynchronous IndexedDB transaction has completed.
    try { sessionStorage.setItem("embedded-guide-reset", "1"); } catch (_) {}
    progress = {}; dwell.clear(); resetUntil = Date.now() + 5000;
    renderProgress();
    if (db) {
      try {
        await new Promise((resolve, reject) => {
          const tx = db.transaction("state", "readwrite");
          const request = tx.objectStore("state").openCursor();
          request.onsuccess = () => { const cursor = request.result; if (cursor) { if (cursor.key.startsWith("progress/")) cursor.delete(); cursor.continue(); } };
          tx.oncomplete = resolve; tx.onabort = tx.onerror = () => reject(tx.error);
        });
        try { sessionStorage.removeItem("embedded-guide-reset"); } catch (_) {}
      } catch (_) { storageFailure(); }
    }
    broadcast({ type: "reset" });
    toast("Reading progress reset. Preferences kept.");
    if (reload) location.reload();
  }

  function initMotion() {
    const items = $$("article section, .landing-section, .circuit-board");
    if (motion.matches || !("IntersectionObserver" in window)) { items.forEach(el => el.classList.add("revealed")); return; }
    const observer = new IntersectionObserver(entries => entries.forEach(entry => {
      if (entry.isIntersecting) { entry.target.classList.add("revealed"); observer.unobserve(entry.target); }
    }), { threshold: 0, rootMargin: "0px 0px -20px 0px" });
    items.forEach((el, i) => { el.style.setProperty("--reveal-delay", `${Math.min(i % 3, 2) * 45}ms`); observer.observe(el); el.classList.add("reveal-ready"); });
    motion.addEventListener("change", () => { if (motion.matches) { observer.disconnect(); items.forEach(el => el.classList.add("revealed")); } });
  }
  function initInteractions() {
    document.addEventListener("pointerdown", unlockAudio, { passive: true });
    document.addEventListener("keydown", unlockAudio);
    document.addEventListener("click", e => {
      if (e.target.closest("a, button, summary, select, .bit") && !e.target.closest("[data-answer]")) sound("click");
    });
    $$("[data-answer]").forEach(button => button.addEventListener("click", () => {
      const correct = button.dataset.answer === "correct", box = button.closest(".quick-check");
      box.dataset.result = correct ? "correct" : "incorrect";
      $$("[data-answer]", box).forEach(el => el.setAttribute("aria-pressed", String(el === button)));
      $(".answer-feedback", box).textContent = correct ? "Exactly. 1 << 5 = 32 = 0x20. One bit, one pin." : "Not quite. Count from bit 0: bit 5 is 2⁵ = 32. Try another mask.";
      sound(correct ? "correct" : "incorrect");
    }));
    $(".circuit-pulse")?.addEventListener("click", () => {
      const board = $(".circuit-board"); board.classList.remove("signal-sent");
      void board.offsetWidth; board.classList.add("signal-sent");
      toast("Signal sent: bit 5 → GPIO → LED.");
    });
    let pending = false, lastStep = -1, lastY = scrollY;
    addEventListener("scroll", () => {
      if (pending) return; pending = true;
      requestAnimationFrame(() => {
        const max = document.documentElement.scrollHeight - innerHeight;
        const percentage = max > 0 ? Math.min(100, Math.max(0, scrollY / max * 100)) : 0;
        const fill = $("#progress-fill"); if (fill) fill.style.width = percentage + "%";
        const step = Math.floor(percentage / 10);
        if (lastStep >= 0 && step !== lastStep && Math.abs(scrollY - lastY) > 8 && !document.body.classList.contains("drawer-open")) sound("scroll");
        lastStep = step; lastY = scrollY; pending = false;
      });
    }, { passive: true });
    document.addEventListener("guide:feedback", e => { if (["correct", "incorrect"].includes(e.detail?.type)) sound(e.detail.type); });
    document.addEventListener("keydown", e => {
      if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key.toLowerCase() === "r") { e.preventDefault(); resetProgress(true); }
    });
    $("#reset-progress")?.addEventListener("click", () => resetProgress());
    $("#theme-choice")?.addEventListener("change", e => updatePreferences({ theme: e.target.value }));
    $("#sound-choice")?.addEventListener("change", e => { updatePreferences({ sound: e.target.checked }); unlockAudio(); });
    $("#tracking-choice")?.addEventListener("change", e => updatePreferences({ tracking: e.target.checked }));
    systemTheme.addEventListener("change", applyPreferences);
    document.addEventListener("click", e => { const settings = $(".preferences"); if (settings?.open && !settings.contains(e.target)) settings.open = false; });
    document.addEventListener("keydown", e => { if (e.key === "Escape") { const settings = $(".preferences"); if (settings?.open) { settings.open = false; $("summary", settings).focus(); } } });
  }
  function updatePreferences(change) {
    preferences = validPreferences({ ...preferences, ...change });
    applyPreferences(); persist("preferences", preferences);
    broadcast({ type: "preferences", value: preferences });
  }
  async function initStorage() {
    try {
      db = await openDB();
      db.onversionchange = () => { db.close(); db = null; storageFailure(); };
      for (const record of await records()) {
        if (record.key === "preferences") preferences = validPreferences(record.value);
        else if (record.key.startsWith("progress/") && Array.isArray(record.value)) progress[record.key.slice(9)] = record.value.filter(Number.isInteger);
      }
      let reset = false;
      try { reset = sessionStorage.getItem("embedded-guide-reset") === "1"; } catch (_) {}
      if (reset) await resetProgress();
    } catch (_) { db = null; storageFailure(); }
    storageReady = true;
    applyPreferences();
    if ("BroadcastChannel" in window) {
      channel = new BroadcastChannel("embedded-guide-state");
      channel.onmessage = e => {
        if (e.data.type === "preferences") { preferences = validPreferences(e.data.value); applyPreferences(); }
        if (e.data.type === "progress" && typeof e.data.slug === "string" && Array.isArray(e.data.value)) { progress[e.data.slug] = e.data.value.filter(Number.isInteger); renderProgress(); }
        if (e.data.type === "reset") { progress = {}; dwell.clear(); resetUntil = Date.now() + 5000; renderProgress(); }
      };
    }
  }
  initDrawer(); initReading(); initMotion(); initInteractions(); initStorage();
})();
