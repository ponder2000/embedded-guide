/* app.js — tiny vanilla helpers. Alpine handles interactive components;
   this file only does ambient scroll-reveal and the reading-progress bar. */
(function () {
  "use strict";

  const reduceMotion = window.matchMedia(
    "(prefers-reduced-motion: reduce)"
  ).matches;

  // Reveal <section>s as they scroll into view.
  function initReveal() {
    const sections = document.querySelectorAll("article section");
    if (!sections.length) return;

    if (reduceMotion || !("IntersectionObserver" in window)) {
      sections.forEach((s) => s.classList.add("revealed"));
      return;
    }

    const io = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add("revealed");
            io.unobserve(entry.target);
          }
        });
      },
      { threshold: 0.12, rootMargin: "0px 0px -8% 0px" }
    );
    sections.forEach((s) => {
      io.observe(s);
      s.classList.add("reveal-ready");
    });
  }

  // Native details keeps chapter navigation usable even without JavaScript.
  function initChapterMenu() {
    const menu = document.querySelector(".chapter-menu");
    if (!menu) return;
    const desktop = window.matchMedia("(min-width: 1080px)");
    function syncMenu() { menu.open = desktop.matches; }
    syncMenu();
    desktop.addEventListener("change", syncMenu);
  }

  // Reading-progress bar in the top bar reflects scroll position.
  function initProgress() {
    const fill = document.getElementById("progress-fill");
    if (!fill) return;
    let ticking = false;
    function update() {
      const doc = document.documentElement;
      const max = doc.scrollHeight - doc.clientHeight;
      const pct = max > 0 ? (doc.scrollTop / max) * 100 : 0;
      fill.style.width = pct.toFixed(1) + "%";
      ticking = false;
    }
    window.addEventListener(
      "scroll",
      () => {
        if (!ticking) {
          window.requestAnimationFrame(update);
          ticking = true;
        }
      },
      { passive: true }
    );
    update();
  }

  if (document.readyState !== "loading") {
    initChapterMenu();
    initReveal();
    initProgress();
  } else {
    document.addEventListener("DOMContentLoaded", function () {
      initChapterMenu();
      initReveal();
      initProgress();
    });
  }
})();
