(() => {
  "use strict";

  const path = window.location.pathname;
  if (path !== "/tree" && path !== "/world") return;

  const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)");
  const chroma = ["#00e5e5", "#00aeb5", "#147df5", "#00e676", "#ff4f46", "#ff256e", "#ff7a35", "#ffd21f", "#9b5cff"];\n  const fire = ["#d93628", "#ff7a35", "#ffd21f", "#f4fbff", "#00e5e5", "#147df5"];
  const state = {
    geometry: null,
    quests: new Map(),
    selected: null,
    width: 0,
    height: 0,
    dpr: 1,
    frame: 0,
    start: performance.now(),
    visible: true
  };

  const style = document.createElement("style");
  style.textContent = `
    :root{--aift-ink:#050609;--aift-charcoal:#101116;--aift-cyan:#00e5e5;--aift-teal:#00aeb5;--aift-blue:#147df5;--aift-green:#00e676;--aift-coral:#ff4f46;--aift-pink:#ff256e;--aift-orange:#ff7a35;--aift-yellow:#ffd21f;--aift-violet:#9b5cff;--aift-white:#f4fbff}
    html,body{background:var(--aift-ink)!important}
    .edge.life,.tree-edge.life{stroke:var(--aift-cyan)!important}
    .edge.knowledge,.tree-edge.knowledge{stroke:var(--aift-blue)!important}
    .edge.root,.edge.living-layer,.tree-edge.root,.tree-edge.living-layer{stroke:var(--aift-coral)!important}
    .node.ready circle,.tree-node.ready circle,.world-node.ready circle{stroke:var(--aift-green)!important}
    .node.detected circle,.tree-node.detected circle,.world-node.detected circle{stroke:var(--aift-orange)!important}
    .node.blocked circle,.tree-node.blocked circle,.world-node.blocked circle{stroke:var(--aift-pink)!important}
    .filter.active,.tree-filter.active{color:#fff!important;background:linear-gradient(90deg,rgba(0,174,181,.32),rgba(20,125,245,.3),rgba(255,37,110,.25))!important}
    .quest-rune{color:var(--aift-coral)!important}
    #livingGeometry{position:fixed;inset:0;z-index:0;width:100%;height:100%;pointer-events:none;opacity:.72}
    #geometryTruth{position:fixed;z-index:12;right:max(14px,env(safe-area-inset-right));top:50%;transform:translateY(-50%);width:min(278px,calc(100vw - 28px));padding:13px 14px;border:1px solid rgba(0,229,229,.22);border-radius:18px;background:rgba(5,6,9,.76);backdrop-filter:blur(16px);box-shadow:0 18px 45px rgba(0,0,0,.28);color:#eaf6ff;font:11px/1.45 Inter,system-ui,sans-serif;transition:opacity .2s,transform .2s}
    #geometryTruth[hidden]{display:none}
    #geometryTruth .geometry-kicker{color:var(--aift-cyan);font-size:8px;letter-spacing:.18em;text-transform:uppercase}
    #geometryTruth h3{margin:4px 0 2px;color:#ff8a58;font-size:17px;line-height:1.1}
    #geometryTruth .geometry-role{color:#9cadc3;font-size:9px}
    #geometryTruth .geometry-grid{display:grid;grid-template-columns:1fr 1fr;gap:7px;margin:10px 0}
    #geometryTruth .geometry-stat{padding:7px;border-radius:10px;background:rgba(255,255,255,.045)}
    #geometryTruth small{display:block;color:#71859f;font-size:7px;letter-spacing:.08em;text-transform:uppercase}
    #geometryTruth strong{display:block;margin-top:2px;color:#eff7ff;font-size:11px}
    #geometryTruth .geometry-cycle{color:#9bb0c7;font-size:8px;overflow-wrap:anywhere}
    #geometryTruth a{display:block;margin-top:10px;padding:9px 11px;border-radius:999px;background:linear-gradient(90deg,rgba(0,229,229,.2),rgba(20,125,245,.2),rgba(255,37,110,.2));color:#e8ffff;text-align:center;text-decoration:none;font-size:9px;font-weight:800;letter-spacing:.05em}
    #geometryTruth button{position:absolute;right:8px;top:7px;border:0;background:transparent;color:#8498b1;font-size:18px}
    #geometryLegend{position:fixed;z-index:11;left:50%;bottom:max(12px,env(safe-area-inset-bottom));transform:translateX(-50%);padding:6px 10px;border-radius:999px;background:rgba(5,6,9,.68);backdrop-filter:blur(10px);color:#8195ac;font:8px/1.2 Inter,system-ui,sans-serif;letter-spacing:.08em;pointer-events:none;white-space:nowrap}
    @media(max-width:700px){#geometryTruth{left:12px;right:12px;top:auto;bottom:max(78px,calc(env(safe-area-inset-bottom) + 66px));width:auto;transform:none}#geometryLegend{display:none}}
    @media(prefers-reduced-motion:reduce){#livingGeometry{opacity:.55}}
  `;
  document.head.append(style);

  const canvas = document.createElement("canvas");
  canvas.id = "livingGeometry";
  canvas.setAttribute("aria-hidden", "true");
  document.body.prepend(canvas);
  const ctx = canvas.getContext("2d", {alpha: true});

  const panel = document.createElement("aside");
  panel.id = "geometryTruth";
  panel.hidden = true;
  panel.innerHTML = `
    <button type="button" aria-label="Close geometry details">×</button>
    <div class="geometry-kicker">Living Federation Geometry</div>
    <h3></h3>
    <div class="geometry-role"></div>
    <div class="geometry-grid"></div>
    <div class="geometry-cycle"></div>
    <a href="/">Open restoration quest in MoBox</a>
  `;
  document.body.append(panel);
  panel.querySelector("button").addEventListener("click", () => {
    panel.hidden = true;
    state.selected = null;
  });

  const legend = document.createElement("div");
  legend.id = "geometryLegend";
  legend.textContent = "Mandelbrot identity · Fibonacci space · torus flow · evidence governs light";
  document.body.append(legend);

  function resize() {
    state.width = window.innerWidth;
    state.height = window.innerHeight;
    state.dpr = Math.min(window.devicePixelRatio || 1, 1.5);
    canvas.width = Math.max(1, Math.floor(state.width * state.dpr));
    canvas.height = Math.max(1, Math.floor(state.height * state.dpr));
    canvas.style.width = state.width + "px";
    canvas.style.height = state.height + "px";
    ctx.setTransform(state.dpr, 0, 0, state.dpr, 0, 0);
  }

  function repositoryName(target) {
    const interactive = target.closest(".node,.world-node,.tree-node,[data-repository]");
    if (!interactive) return "";
    const declared = interactive.dataset && interactive.dataset.repository;
    if (declared) return declared;
    const label = interactive.getAttribute("aria-label") || "";
    return label.split(",")[0].trim();
  }

  function selectByName(name) {
    if (!state.geometry || !name) return;
    const node = state.geometry.nodes.find(item => item.name === name);
    if (!node) return;
    state.selected = node;
    const quest = state.quests.get(node.repository_id);
    panel.querySelector("h3").textContent = node.name;
    panel.querySelector(".geometry-role").textContent =
      `${node.sacred_form} · symmetry ${node.symmetry} · ${node.role}`;
    const values = [
      ["Status", node.status],
      ["Coherence", node.coherence + "%"],
      ["Mandelbrot", node.mandelbrot.bounded ? "bounded" : `escape ${node.mandelbrot.iterations}`],
      ["Evidence", node.evidence_count]
    ];
    const grid = panel.querySelector(".geometry-grid");
    grid.replaceChildren();
    for (const [label, value] of values) {
      const item = document.createElement("div");
      item.className = "geometry-stat";
      const small = document.createElement("small");
      const strong = document.createElement("strong");
      small.textContent = label;
      strong.textContent = String(value);
      item.append(small, strong);
      grid.append(item);
    }
    panel.querySelector(".geometry-cycle").textContent =
      quest ? `${quest.title} · ${quest.status} · ${node.torus.cycle}` : node.torus.cycle;
    const link = panel.querySelector("a");
    link.href = `/?inspect=${encodeURIComponent(node.name)}`;
    link.textContent = quest && quest.status !== "complete"
      ? "Enter restoration quest in MoBox"
      : "Inspect evidence in MoBox";
    panel.hidden = false;
  }

  document.addEventListener("click", event => {
    const name = repositoryName(event.target);
    if (name) selectByName(name);
  }, true);

  function project(node, rotation) {
    const p = node.position;
    const cos = Math.cos(rotation), sin = Math.sin(rotation);
    const x = p.x * cos - p.z * sin;
    const z = p.x * sin + p.z * cos;
    const perspective = 1 / (1.7 - z * .42);
    const radius = Math.min(state.width, state.height) * (path === "/tree" ? .38 : .44);
    return {
      x: state.width / 2 + x * radius * perspective,
      y: state.height / 2 + p.y * radius * perspective,
      z,
      perspective
    };
  }

  function color(node, alpha) {
    if (node.status === "blocked") return `rgba(255,120,126,${alpha})`;
    if (node.status === "ready") return `rgba(136,242,172,${alpha})`;
    return `rgba(242,199,106,${alpha})`;
  }

  function polygon(x, y, radius, sides, rotation) {
    ctx.beginPath();
    for (let i = 0; i < sides; i++) {
      const angle = rotation + i * Math.PI * 2 / sides;
      const px = x + Math.cos(angle) * radius;
      const py = y + Math.sin(angle) * radius;
      if (i === 0) ctx.moveTo(px, py); else ctx.lineTo(px, py);
    }
    ctx.closePath();
  }

  function drawTorus(node, point, flow) {
    const coherence = Math.max(.08, node.coherence / 100);
    const base = (8 + 16 * coherence) * point.perspective;
    const sides = Math.max(3, Math.min(13, node.symmetry || 6));
    ctx.save();
    ctx.translate(point.x, point.y);
    ctx.rotate(flow);
    ctx.strokeStyle = harmonicColor(node, 0, state.selected === node ? .96 : .5);
    ctx.lineWidth = state.selected === node ? 2.4 : 1.15;
    ctx.shadowColor = harmonicColor(node, 1, .78);
    ctx.shadowBlur = state.selected === node ? 18 : 7;
    ctx.beginPath();
    ctx.ellipse(0, 0, base * 1.7, base * .62, node.torus.identity_phase, 0, Math.PI * 2);
    ctx.stroke();
    ctx.strokeStyle = harmonicColor(node, 2, state.selected === node ? .92 : .58);
    polygon(0, 0, base, sides, -flow * .5);
    ctx.stroke();
    ctx.strokeStyle = color(node, .5);
    ctx.beginPath();
    ctx.ellipse(0, 0, base * 1.18, base * .38, -node.torus.identity_phase, 0, Math.PI * 2);
    ctx.stroke();
    ctx.globalAlpha = .72;
    ctx.shadowColor = energyColor(node, .95);
    ctx.shadowBlur = 12 + 18 * coherence;
    ctx.beginPath();
    ctx.arc(0, 0, Math.max(2.1, base * .16), 0, Math.PI * 2);
    ctx.fillStyle = energyColor(node, .98);
    ctx.fill();
    ctx.globalAlpha = .9;
    ctx.beginPath();
    ctx.arc(-base * .035, -base * .035, Math.max(.7, base * .045), 0, Math.PI * 2);
    ctx.fillStyle = "rgba(255,255,255,.96)";
    ctx.fill();
    ctx.restore();
  }

  function render(now) {
    state.frame = requestAnimationFrame(render);
    if (!state.geometry || !state.visible) return;
    ctx.clearRect(0, 0, state.width, state.height);
    const elapsed = reducedMotion.matches ? 0 : (now - state.start) / 1000;
    const rotation = elapsed * .035;
    const points = state.geometry.nodes.map(node => ({
      node,
      point: project(node, rotation + node.torus.identity_phase * .08)
    })).sort((a, b) => a.point.z - b.point.z);

    ctx.save();
    ctx.lineWidth = .7;
    for (let i = 1; i < points.length; i++) {
      const a = points[i - 1], b = points[i];
      ctx.strokeStyle = harmonicColor(b.node, 1, .12);
      ctx.beginPath();
      ctx.moveTo(a.point.x, a.point.y);
      ctx.lineTo(b.point.x, b.point.y);
      ctx.stroke();
    }
    ctx.restore();

    for (const item of points) {
      const speed = .08 + item.node.coherence / 1000;
      drawTorus(item.node, item.point, item.node.torus.identity_phase + elapsed * speed);
    }
  }

  async function load() {
    const [geometryResponse, treeResponse] = await Promise.all([
      fetch("/v1/federation/geometry", {cache: "no-store"}),
      fetch("/v1/federation/tree", {cache: "no-store"})
    ]);
    if (!geometryResponse.ok) throw new Error(`Geometry API ${geometryResponse.status}`);
    if (!treeResponse.ok) throw new Error(`Tree API ${treeResponse.status}`);
    state.geometry = await geometryResponse.json();
    const tree = await treeResponse.json();
    for (const quest of tree.quests || []) {
      const existing = state.quests.get(quest.repository_id);
      if (!existing || (existing.status === "complete" && quest.status !== "complete")) {
        state.quests.set(quest.repository_id, quest);
      }
    }
    legend.textContent =
      `${state.geometry.nodes.length} living nodes · ${state.geometry.law.temporal_model}`;
  }

  document.addEventListener("visibilitychange", () => {
    state.visible = !document.hidden;
  });
  window.addEventListener("resize", resize, {passive: true});
  reducedMotion.addEventListener?.("change", () => {
    state.start = performance.now();
  });

  resize();
  requestAnimationFrame(render);
  load().catch(error => {
    legend.textContent = "Geometry blocked · " + error.message;
    legend.style.color = "#ff9b9b";
  });
})();
