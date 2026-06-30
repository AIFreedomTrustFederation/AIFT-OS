import { motion } from "framer-motion";
import { useState } from "react";

// Spectrum colors for governance layers (inner=violet/high-coherence → outer=red/expansion)
const RING_COLORS = [
  { stroke: "#cc88ff", glow: "rgba(200,136,255,0.6)", label: "255,255,255" },   // Core — white/unity
  { stroke: "#8855ee", glow: "rgba(136,85,238,0.5)", label: "136,85,238" },     // Council — violet
  { stroke: "#4499ff", glow: "rgba(68,153,255,0.5)", label: "68,153,255" },     // Validators — blue
  { stroke: "#00ddaa", glow: "rgba(0,221,170,0.5)", label: "0,221,170" },       // Runtime — cyan/green
  { stroke: "#88dd00", glow: "rgba(136,221,0,0.4)", label: "136,221,0" },       // Nodes — yellow-green
  { stroke: "#ff9900", glow: "rgba(255,153,0,0.4)", label: "255,153,0" },       // Sovereign — amber
];

const LAYERS = [
  { id: 0, name: "Constitutional Core",          radius: 55,  strokeWidth: 2.5, dashArray: "0" },
  { id: 1, name: "Council of 24 Elders",         radius: 110, strokeWidth: 1.5, dashArray: "4 4",  nodes: 24 },
  { id: 2, name: "12 Validator Orders",           radius: 175, strokeWidth: 2,   dashArray: "0",    nodes: 12 },
  { id: 3, name: "Federation Runtime & Services", radius: 250, strokeWidth: 1,   dashArray: "3 7",  nodes: 10 },
  { id: 4, name: "Sovereign Federation Nodes",    radius: 335, strokeWidth: 1.5, dashArray: "10 5", nodes: 6 },
  { id: 5, name: "Individual Sovereign Trees",    radius: 430, strokeWidth: 0.5, dashArray: "1 10", nodes: 32 },
];

const LAYER_DETAILS: Record<number, { title: string; sub: string; desc: string; color: string }> = {
  0: {
    title: "Constitutional Core",
    sub: "Unity · Maximum Coherence",
    desc: "Truth before authority. Evidence before claims. The stable resonant center — where all spectra merge into white light.",
    color: "#e0c060",
  },
  1: {
    title: "Council of 24 Elders",
    sub: "Violet · Spiritual Integration",
    desc: "Constitutional review and stewardship. Role-based, transparent, evidence-driven. Human + AI participants maintain the living covenant.",
    color: "#bb88ff",
  },
  2: {
    title: "12 Validator Orders",
    sub: "Blue · Truth · Order",
    desc: "Vision · Wisdom · Reconciliation · Treasury · Ledger · Judgment · Timekeeping · Equity · Faith · Renewal · Intercession · Sovereignty. Cryptographic identity trails.",
    color: "#66aaff",
  },
  3: {
    title: "Federation Runtime & Services",
    sub: "Cyan-Green · Equilibrium · Communication",
    desc: "Discovery · Registry · Runtime · Scheduler · Event Bus · Documentation · Intelligence · Patch Engine · Architecture · Diagnostics.",
    color: "#00ddaa",
  },
  4: {
    title: "Sovereign Federation Nodes",
    sub: "Yellow-Green · Intellect · Illumination",
    desc: "BookSmith AI · AIFT Forge · Aether Coin · DynastyLink · Capital City Provisions · AIFT-OS. Each node sovereign, each node federated.",
    color: "#aaee44",
  },
  5: {
    title: "Individual Sovereign Trees",
    sub: "Amber · Creation · Energy",
    desc: "Each person owns a complete local operating system. Private. Shared. Public. No centralized authority can revoke sovereignty.",
    color: "#ffaa33",
  },
};

export function ArchitectureRings() {
  const [activeLayer, setActiveLayer] = useState<number | null>(null);
  const center = 500;

  return (
    <section className="py-24 px-6 relative overflow-hidden flex flex-col items-center">
      {/* Ambient background */}
      <div className="absolute inset-0 pointer-events-none">
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[900px] h-[900px] rounded-full bg-violet-900/5 blur-[150px]" />
      </div>

      {/* Section header */}
      <motion.div
        initial={{ opacity: 0, y: 30 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
        transition={{ duration: 1 }}
        className="max-w-4xl text-center mb-8 relative z-10"
      >
        <p className="font-mono text-[10px] uppercase tracking-[0.3em] text-white/30 mb-4">
          Wheel Within Wheels · Ezekiel 1:16
        </p>
        <h2
          className="font-cinzel text-4xl md:text-5xl font-bold mb-5 text-glow-gold"
          style={{ color: "#e8c040" }}
        >
          The Governance Architecture
        </h2>
        <p className="font-mono text-sm text-white/40 max-w-xl mx-auto">
          From constitutional coherence at the center outward to sovereign expansion. Click any ring to inspect its domain.
        </p>
      </motion.div>

      <div className="spectrum-divider w-full max-w-2xl mb-12" />

      {/* Diagram area */}
      <div className="relative w-full max-w-[1000px] flex flex-col items-center">
        {/* Info panel */}
        <motion.div
          animate={{ opacity: activeLayer !== null ? 1 : 0, y: activeLayer !== null ? 0 : 10 }}
          transition={{ duration: 0.3 }}
          className="absolute top-4 right-0 md:-right-4 z-20 w-72 pointer-events-none"
          style={{ maxWidth: "90vw" }}
        >
          {activeLayer !== null && (() => {
            const d = LAYER_DETAILS[activeLayer];
            const c = RING_COLORS[activeLayer];
            return (
              <div
                className="p-5 rounded-xl backdrop-blur-xl pointer-events-auto"
                style={{
                  border: `1px solid ${c.stroke}44`,
                  background: `rgba(0,0,0,0.7)`,
                  boxShadow: `0 0 30px ${c.glow}33`,
                }}
              >
                <div className="font-mono text-[10px] uppercase tracking-widest mb-1" style={{ color: d.color }}>
                  Layer {activeLayer}
                </div>
                <h3 className="font-cinzel text-lg font-bold mb-1" style={{ color: d.color }}>
                  {d.title}
                </h3>
                <p className="font-mono text-[10px] mb-3" style={{ color: d.color + "99" }}>
                  {d.sub}
                </p>
                <p className="text-sm text-white/60 leading-relaxed">
                  {d.desc}
                </p>
              </div>
            );
          })()}
        </motion.div>

        {/* SVG diagram */}
        <svg
          viewBox="0 0 1000 1000"
          className="w-full max-w-[900px]"
          style={{ maxHeight: "90vh" }}
        >
          <defs>
            <filter id="ringGlow" x="-50%" y="-50%" width="200%" height="200%">
              <feGaussianBlur stdDeviation="6" result="coloredBlur" />
              <feMerge>
                <feMergeNode in="coloredBlur" />
                <feMergeNode in="SourceGraphic" />
              </feMerge>
            </filter>
            <filter id="coreGlow" x="-100%" y="-100%" width="300%" height="300%">
              <feGaussianBlur stdDeviation="12" result="coloredBlur" />
              <feMerge>
                <feMergeNode in="coloredBlur" />
                <feMergeNode in="SourceGraphic" />
              </feMerge>
            </filter>
            {/* Radial gradient for core */}
            <radialGradient id="coreRadial" cx="50%" cy="50%" r="50%">
              <stop offset="0%" stopColor="#ffffff" stopOpacity="0.6" />
              <stop offset="40%" stopColor="#ccaaff" stopOpacity="0.3" />
              <stop offset="100%" stopColor="#6622cc" stopOpacity="0" />
            </radialGradient>
          </defs>

          {/* Starfield dots */}
          {Array.from({ length: 40 }).map((_, i) => {
            const angle = (i / 40) * Math.PI * 2;
            const r = 480 - (i % 5) * 20;
            const x = center + Math.cos(angle) * r;
            const y = center + Math.sin(angle) * r;
            return <circle key={i} cx={x} cy={y} r={0.8} fill="white" opacity={0.15 + (i % 3) * 0.1} />;
          })}

          {/* Core ambient glow */}
          <circle cx={center} cy={center} r={130} fill="url(#coreRadial)" />

          {/* Rings */}
          {LAYERS.map((layer, index) => {
            const isActive = activeLayer === layer.id;
            const rc = RING_COLORS[index];

            // Node positions
            const nodes: React.ReactNode[] = [];
            if (layer.nodes) {
              const step = (Math.PI * 2) / layer.nodes;
              for (let i = 0; i < layer.nodes; i++) {
                const a = i * step;
                const nx = center + Math.cos(a) * layer.radius;
                const ny = center + Math.sin(a) * layer.radius;
                nodes.push(
                  <circle
                    key={`n-${layer.id}-${i}`}
                    cx={nx} cy={ny}
                    r={isActive ? 5 : 2.5}
                    fill={isActive ? rc.stroke : `rgba(${rc.label},0.4)`}
                    className="transition-all duration-300"
                  />
                );
              }
            }

            return (
              <g
                key={layer.id}
                className="cursor-pointer"
                onClick={() => setActiveLayer(isActive ? null : layer.id)}
                onMouseEnter={() => setActiveLayer(layer.id)}
                onMouseLeave={() => setActiveLayer(null)}
              >
                {/* Wide hit area */}
                <circle
                  cx={center} cy={center} r={layer.radius + 20}
                  fill="transparent" stroke="transparent" strokeWidth={40}
                />

                <motion.g
                  animate={{ rotate: index % 2 === 0 ? 360 : -360 }}
                  transition={{ duration: 80 + index * 25, repeat: Infinity, ease: "linear" }}
                  style={{ originX: `${center}px`, originY: `${center}px` }}
                >
                  {/* Glow halo when active */}
                  {isActive && (
                    <circle
                      cx={center} cy={center} r={layer.radius}
                      fill="transparent"
                      stroke={rc.stroke}
                      strokeWidth={14}
                      strokeOpacity={0.07}
                      filter="url(#ringGlow)"
                    />
                  )}

                  <circle
                    cx={center} cy={center} r={layer.radius}
                    fill="transparent"
                    stroke={isActive ? rc.stroke : `rgba(${rc.label},${index === 0 ? 0.5 : 0.2})`}
                    strokeWidth={isActive ? layer.strokeWidth * 2.5 : layer.strokeWidth}
                    strokeDasharray={layer.dashArray}
                    filter={isActive ? "url(#ringGlow)" : ""}
                    className="transition-all duration-500"
                  />

                  {nodes}
                </motion.g>

                {/* Label on ring */}
                {!layer.nodes && (
                  <text
                    x={center + layer.radius + 8}
                    y={center}
                    fill={`rgba(${rc.label},0.4)`}
                    fontSize="9"
                    fontFamily="Space Mono, monospace"
                    className="pointer-events-none"
                  >
                    {layer.name.toUpperCase()}
                  </text>
                )}
              </g>
            );
          })}

          {/* Core center dot */}
          <circle cx={center} cy={center} r={12} fill="#ffffff" filter="url(#coreGlow)" opacity={0.9} />
          <circle cx={center} cy={center} r={6}  fill="#ffffff" opacity={1} />

          {/* Labels for each ring at fixed angle */}
          {LAYERS.map((layer, index) => {
            const rc = RING_COLORS[index];
            const labelAngle = -Math.PI / 2 - 0.3 + index * 0.08;
            const lx = center + Math.cos(labelAngle) * (layer.radius + 14);
            const ly = center + Math.sin(labelAngle) * (layer.radius + 14);
            return (
              <text
                key={`lbl-${layer.id}`}
                x={lx} y={ly}
                fill={`rgba(${rc.label},0.5)`}
                fontSize="8"
                fontFamily="Space Mono, monospace"
                textAnchor="middle"
                className="pointer-events-none select-none"
              >
                {layer.name}
              </text>
            );
          })}
        </svg>

        {/* Legend row */}
        <div className="flex flex-wrap justify-center gap-4 mt-4 max-w-2xl">
          {LAYERS.map((layer, i) => (
            <button
              key={layer.id}
              onClick={() => setActiveLayer(activeLayer === layer.id ? null : layer.id)}
              className="flex items-center gap-2 px-3 py-1.5 rounded-full transition-all duration-300"
              style={{
                border: `1px solid ${RING_COLORS[i].stroke}${activeLayer === layer.id ? "88" : "22"}`,
                background: activeLayer === layer.id ? `${RING_COLORS[i].stroke}11` : "transparent",
              }}
            >
              <span className="w-2 h-2 rounded-full" style={{ background: RING_COLORS[i].stroke, boxShadow: `0 0 6px ${RING_COLORS[i].stroke}` }} />
              <span className="font-mono text-[10px] text-white/50">{layer.name}</span>
            </button>
          ))}
        </div>
      </div>
    </section>
  );
}
