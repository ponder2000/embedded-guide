/* Explicit, pausable teaching timelines. Real MHz/baud is intentionally slowed. */
(function () {
  "use strict";
  if (!window.Alpine) return;
  const NS = "http://www.w3.org/2000/svg";
  function svgNode(name, attributes) {
    const node = document.createElementNS(NS, name);
    Object.entries(attributes).forEach(([key, value]) => node.setAttribute(key, value));
    return node;
  }
  document.querySelectorAll("[data-simulation]").forEach(root => {
    const model = Alpine.$data(root), kind = root.dataset.simulation;
    let playing = false, frame = 0, previous = 0, elapsed = 0, tick = 0;
    const panel = document.createElement("div"); panel.className = "simulation-player";
    panel.innerHTML = '<div class="simulation-heading"></div><div class="simulation-stage"></div><div class="simulation-controls"><button type="button" class="btn btn-accent sim-play" aria-pressed="false">▶ Play</button><button type="button" class="btn sim-step">Step →</button><button type="button" class="btn sim-reset">Reset</button><label>Speed <select class="sim-speed"><option value="0.5">0.5×</option><option value="1" selected>1×</option><option value="2">2×</option></select></label></div><p class="simulation-readout"></p><p class="simulation-announcement" role="status" aria-live="polite"></p>';
    panel.querySelector(".simulation-heading").textContent = root.dataset.simulationName + " · slowed down to see the signal";
    const stage = panel.querySelector(".simulation-stage"), play = panel.querySelector(".sim-play");
    const readout = panel.querySelector(".simulation-readout"), announcement = panel.querySelector(".simulation-announcement");
    const speed = panel.querySelector(".sim-speed");
    let svg, cursor, paths = [], bitCells = [];
    if (kind !== "steps") {
      svg = svgNode("svg", { viewBox: kind === "clocks" ? "0 0 600 156" : "0 0 600 96", role: "img", "aria-label": root.dataset.simulationName + " waveform and playback cursor" });
      const lanes = kind === "clocks" ? ["SYSCLK", "HCLK", "PCLK"] : [kind === "uart" ? "TX" : "PWM"];
      lanes.forEach((name, i) => {
        const label = svgNode("text", { x: 4, y: 35 + i * 46, fill: "var(--ink-soft)", "font-size": 10, "font-family": "IBM Plex Mono" }); label.textContent = name; svg.append(label);
        const path = svgNode("path", { fill: "none", stroke: i === 2 ? "var(--green)" : "var(--accent)", "stroke-width": 2 }); paths.push(path); svg.append(path);
      });
      cursor = svgNode("line", { x1: 80, x2: 80, y1: 8, y2: kind === "clocks" ? 150 : 80, stroke: "var(--blue)", "stroke-width": 2, "stroke-dasharray": "3 3" }); svg.append(cursor); stage.append(svg);
      if (kind === "uart") {
        const bits = document.createElement("div"); bits.className = "uart-playback-bits";
        Array.from({ length: 10 }, (_, i) => {
          const cell = document.createElement("span"); cell.dataset.bit = i; bits.append(cell); bitCells.push(cell);
        }); stage.append(bits);
      }
    } else {
      stage.classList.add("step-playback");
    }
    root.append(panel);
    const frameBits = () => [0, ...Array.from({ length: 8 }, (_, i) => model.dataBit(i)), 1];
    const lastTick = () => kind === "uart" ? 9 : kind === "steps" ? model.steps.length - 1 : kind === "pwm" ? 99 : 63;
    function waveform(values, lane) {
      const width = 510 / values.length;
      let d = "";
      values.forEach((value, i) => {
        const x = 80 + i * width, y = (value ? 16 : 43) + lane * 46;
        d += (i === 0 ? "M" : "L") + x + " " + y + " L" + (x + width) + " " + y + " ";
      }); return d;
    }
    function render() {
      if (kind === "steps") {
        tick = model.step;
        stage.textContent = `${tick + 1} / ${model.steps.length} · ${model.steps[tick].title}`;
        readout.textContent = "Play advances one stage at a time. Pause to inspect; use Step for manual control.";
      } else if (kind === "uart") {
        const bits = frameBits(); paths[0].setAttribute("d", waveform(bits, 0));
        bitCells.forEach((cell, i) => { cell.textContent = `${i === 0 ? "START" : i === 9 ? "STOP" : "D" + (i - 1)}: ${bits[i]}`; cell.classList.toggle("active", i === tick); });
        cursor.setAttribute("x1", 80 + (tick + .5) * 51); cursor.setAttribute("x2", 80 + (tick + .5) * 51);
        readout.textContent = `${tick === 0 ? "Start bit" : tick === 9 ? "Stop bit" : "Data bit " + (tick - 1)} · TX = ${bits[tick]} · ${bits[tick] ? "HIGH" : "LOW"} · ${model.hex()} (LSB first)`;
      } else if (kind === "clocks") {
        const divisors = [1, model.ahbPre, model.ahbPre * model.apbPre];
        paths.forEach((path, i) => path.setAttribute("d", waveform(Array.from({ length: 64 }, (_, n) => Math.floor(n / divisors[i]) % 2 === 0 ? 1 : 0), i)));
        cursor.setAttribute("x1", 80 + (tick + .5) / 64 * 510); cursor.setAttribute("x2", 80 + (tick + .5) / 64 * 510);
        readout.textContent = `Tick ${tick + 1} · SYSCLK ${model.sysclk()} MHz → HCLK ${model.ahbClk()} MHz → PCLK ${model.apbClk()} MHz. Dividers change the relative edge spacing.`;
      } else if (kind === "pwm") {
        paths[0].setAttribute("d", waveform(Array.from({ length: 100 }, (_, n) => n < model.duty ? 1 : 0), 0));
        cursor.setAttribute("x1", 80 + tick / 100 * 510); cursor.setAttribute("x2", 80 + tick / 100 * 510);
        readout.textContent = `Counter ${tick} / 99 · CCR ${model.onTime()} · output ${tick < model.duty ? "HIGH" : "LOW"} · duty ${model.duty}%`;
      }
    }
    function pause(message = "Paused. Your position is kept.") {
      playing = false; cancelAnimationFrame(frame); previous = 0;
      play.textContent = "▶ Play"; play.setAttribute("aria-pressed", "false");
      panel.dataset.playing = "false"; announcement.textContent = message;
    }
    function advance() {
      if (kind === "steps") tick = model.step;
      if (tick >= lastTick()) {
        if (kind === "uart" || kind === "steps") { pause("Playback complete. Press Play to replay."); return; }
        tick = 0;
      } else tick++;
      if (kind === "steps") model.step = tick;
      render();
    }
    function animate(time) {
      if (!playing) return;
      if (previous) elapsed += Math.min(time - previous, 100);
      previous = time;
      const interval = (kind === "steps" ? 1400 : kind === "uart" ? 500 : kind === "pwm" ? 45 : 130) / Number(speed.value);
      if (elapsed >= interval) { elapsed %= interval; advance(); }
      if (playing) frame = requestAnimationFrame(animate);
    }
    play.addEventListener("click", () => {
      if (playing) { pause(); return; }
      if (kind === "steps") tick = model.step;
      if ((kind === "uart" || kind === "steps") && tick >= lastTick()) { tick = 0; if (kind === "steps") model.step = 0; }
      playing = true; elapsed = 0; previous = 0;
      play.textContent = "Ⅱ Pause"; play.setAttribute("aria-pressed", "true"); panel.dataset.playing = "true";
      announcement.textContent = "Playing at " + speed.options[speed.selectedIndex].text + ".";
      render(); frame = requestAnimationFrame(animate);
    });
    panel.querySelector(".sim-step").addEventListener("click", () => { pause("Advanced one step."); advance(); });
    panel.querySelector(".sim-reset").addEventListener("click", () => { pause("Simulation reset."); tick = 0; elapsed = 0; if (kind === "steps") model.step = 0; render(); });
    root.addEventListener("input", e => { if (!panel.contains(e.target)) { if (kind === "uart") { pause(); tick = 0; } queueMicrotask(render); } });
    root.addEventListener("change", e => {
      if (panel.contains(e.target)) return;
      queueMicrotask(() => {
        render();
        if (kind === "clocks") document.dispatchEvent(new CustomEvent("guide:feedback", { detail: { type: model.sysclk() > 72 ? "incorrect" : "correct" } }));
      });
    });
    root.addEventListener("click", e => {
      if (panel.contains(e.target) || !e.target.closest("button")) return;
      if (kind === "steps" || kind === "uart") { pause(); if (kind === "uart") tick = 0; }
      queueMicrotask(render);
    });
    document.addEventListener("visibilitychange", () => { if (document.hidden && playing) pause("Paused while the tab is hidden."); });
    window.addEventListener("pagehide", () => pause());
    render();
  });
})();
