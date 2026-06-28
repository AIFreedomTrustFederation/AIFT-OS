import { motion } from "framer-motion";
import { useState } from "react";

// Rainbow spectrum: I=red/foundation → VII=violet/exploration
const LAYERS = [
  {
    id: "VII",
    num: 7,
    name: "Exploration",
    spectrum: "#9922ff",
    spectrumRgb: "153,34,255",
    phase: "Violet · Convergence · Highest Integration",
    quote: "The frontier of the known, reached only by those who first understood the foundation.",
    items: ["Scientific research", "Distributed computation", "Robotics", "Autonomous laboratories", "Advanced visualization", "Digital twins", "Space infrastructure", "Simulation environments"],
  },
  {
    id: "VI",
    num: 6,
    name: "Applications",
    spectrum: "#3355ff",
    spectrumRgb: "51,85,255",
    phase: "Blue · Truth · Higher Mind",
    quote: "Software as civilization. Applications that serve sovereign beings.",
    items: ["DynastyLink", "AIFT Forge", "Capital City Provisions", "Publishing systems", "Websites", "Mobile apps", "Community portals", "Business platforms"],
  },
  {
    id: "V",
    num: 5,
    name: "Economy",
    spectrum: "#00ccff",
    spectrumRgb: "0,204,255",
    phase: "Cyan · Harmony · Communication",
    quote: "Wealth measured in truth, contribution, and verified capability.",
    items: ["Aether Coin", "Token systems", "Trust accounting", "Marketplace", "Cooperative exchanges", "Asset registries", "Treasury systems", "Economic governance"],
  },
  {
    id: "IV",
    num: 4,
    name: "Federation",
    spectrum: "#00dd55",
    spectrumRgb: "0,221,85",
    phase: "Green · Equilibrium · The Observer · Present Moment",
    quote: "No centralized authority. Every participant remains sovereign.",
    items: ["Trust relationships", "Federation registries", "Service contracts", "Capability publication", "Synchronization", "Discovery", "Permissions", "Shared governance"],
  },
  {
    id: "III",
    num: 3,
    name: "Intelligence",
    spectrum: "#ffdd00",
    spectrumRgb: "255,221,0",
    phase: "Yellow · Intellect · Discernment",
    quote: "AI never invents reality. It assists discovery.",
    items: ["Local LLM providers", "Prompt registries", "Agent orchestration", "Planning engines", "Code analysis", "Architecture reasoning", "Diagnostics", "Knowledge synthesis"],
  },
  {
    id: "II",
    num: 2,
    name: "Knowledge",
    spectrum: "#ff8800",
    spectrumRgb: "255,136,0",
    phase: "Orange · Creation · Memory",
    quote: "Documentation is not overhead. Documentation is how the system remembers.",
    items: ["BookSmith AI", "Documentation generation", "Manuals", "API references", "Knowledge graphs", "Research", "Educational content", "Publications", "Evidence registries"],
  },
  {
    id: "I",
    num: 1,
    name: "Sovereign Foundation",
    spectrum: "#ff2200",
    spectrumRgb: "255,34,0",
    phase: "Red · Foundation · Survival · Root",
    quote: "Nothing above can stand without what is verified below.",
    items: ["Identity", "Local storage", "Security", "Trust", "Discovery", "Registry", "Configuration", "Diagnostics", "Runtime", "Verification"],
  },
];

export function LivingLayers() {
  const [hoveredLayer, setHoveredLayer] = useState<string | null>(null);

  return (
    <section className="py-24 px-6 relative overflow-hidden">
      <div className="absolute inset-0 pointer-events-none">
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full h-px spectrum-divider opacity-30" />
        <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-full h-px spectrum-divider opacity-30" />
      </div>

      <div className="max-w-5xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 1 }}
          className="text-center mb-6"
        >
          <p className="font-mono text-[10px] uppercase tracking-[0.3em] text-white/30 mb-4">
            Standing Wave Nodes · Resonant Intersections of the Toroidal Field
          </p>
          <h2
            className="font-cinzel text-4xl md:text-5xl font-bold mb-4 text-glow-gold"
            style={{ color: "#e8c040" }}
          >
            The Seven Living Layers
          </h2>
          <p className="font-mono text-sm text-white/40 max-w-2xl mx-auto">
            Each layer is a stable resonance pattern in the field. The layers do not exist in linear time — they are simultaneous, self-reinforcing, and holographic.
          </p>
        </motion.div>

        <div className="spectrum-divider my-10" />

        {/* Color / phase map legend */}
        <div className="flex justify-center gap-1 mb-12 flex-wrap">
          {[...LAYERS].reverse().map((l) => (
            <div
              key={l.id}
              className="flex items-center gap-1.5 px-2 py-1 rounded font-mono text-[9px] uppercase tracking-wider"
              style={{ background: `rgba(${l.spectrumRgb},0.06)`, color: l.spectrum }}
            >
              <span
                className="w-2 h-2 rounded-full shrink-0"
                style={{ background: l.spectrum, boxShadow: `0 0 6px ${l.spectrum}` }}
              />
              {l.id} {l.name}
            </div>
          ))}
        </div>

        {/* Layer stack */}
        <div className="flex flex-col gap-3 relative">
          {/* Vertical connector */}
          <div
            className="absolute left-6 top-6 bottom-6 w-px"
            style={{
              background: "linear-gradient(to bottom, #9922ff, #3355ff, #00ccff, #00dd55, #ffdd00, #ff8800, #ff2200)",
              opacity: 0.3,
            }}
          />

          {LAYERS.map((layer, idx) => {
            const isHovered = hoveredLayer === layer.id;
            return (
              <motion.div
                key={layer.id}
                initial={{ opacity: 0, x: -20 }}
                whileInView={{ opacity: 1, x: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: idx * 0.08 }}
                onHoverStart={() => setHoveredLayer(layer.id)}
                onHoverEnd={() => setHoveredLayer(null)}
                className="relative z-10 rounded-xl cursor-default transition-all duration-500 overflow-hidden"
                style={{
                  border: `1px solid ${layer.spectrum}${isHovered ? "44" : "18"}`,
                  background: isHovered
                    ? `rgba(${layer.spectrumRgb},0.05)`
                    : "rgba(255,255,255,0.01)",
                  boxShadow: isHovered ? `0 0 40px rgba(${layer.spectrumRgb},0.12)` : "none",
                }}
              >
                {/* Spectrum left bar */}
                <div
                  className="absolute left-0 top-0 bottom-0 w-1 transition-all duration-500"
                  style={{
                    background: layer.spectrum,
                    opacity: isHovered ? 0.8 : 0.2,
                    boxShadow: isHovered ? `0 0 12px ${layer.spectrum}` : "none",
                  }}
                />

                <div className="pl-8 pr-6 py-5 flex flex-col md:flex-row gap-4 items-start md:items-center">
                  {/* ID badge */}
                  <div
                    className="shrink-0 w-14 h-14 rounded-full flex items-center justify-center font-cinzel font-bold text-lg transition-all duration-500"
                    style={{
                      border: `1px solid ${layer.spectrum}${isHovered ? "88" : "33"}`,
                      color: isHovered ? layer.spectrum : `rgba(${layer.spectrumRgb},0.4)`,
                      background: isHovered ? `rgba(${layer.spectrumRgb},0.08)` : "transparent",
                      boxShadow: isHovered ? `0 0 20px rgba(${layer.spectrumRgb},0.2), inset 0 0 20px rgba(${layer.spectrumRgb},0.05)` : "none",
                    }}
                  >
                    {layer.id}
                  </div>

                  <div className="flex-1 min-w-0">
                    {/* Header row */}
                    <div className="flex flex-wrap items-baseline gap-3 mb-1">
                      <h3
                        className="font-cinzel text-xl font-bold transition-colors duration-300"
                        style={{ color: isHovered ? layer.spectrum : "rgba(255,255,255,0.75)" }}
                      >
                        {layer.name}
                      </h3>
                      <span
                        className="font-mono text-[9px] uppercase tracking-widest"
                        style={{ color: `rgba(${layer.spectrumRgb},${isHovered ? 0.6 : 0.25})` }}
                      >
                        {layer.phase}
                      </span>
                    </div>

                    {/* Quote */}
                    <p
                      className="font-mono text-xs italic mb-3 transition-colors duration-300"
                      style={{ color: isHovered ? `rgba(${layer.spectrumRgb},0.7)` : "rgba(255,255,255,0.2)" }}
                    >
                      "{layer.quote}"
                    </p>

                    {/* Items — expanded on hover */}
                    <motion.div
                      animate={{ height: isHovered ? "auto" : 0, opacity: isHovered ? 1 : 0 }}
                      transition={{ duration: 0.35 }}
                      className="overflow-hidden"
                    >
                      <div className="flex flex-wrap gap-2 pt-1">
                        {layer.items.map((item) => (
                          <span
                            key={item}
                            className="font-mono text-[10px] px-2 py-0.5 rounded"
                            style={{
                              background: `rgba(${layer.spectrumRgb},0.08)`,
                              border: `1px solid rgba(${layer.spectrumRgb},0.2)`,
                              color: `rgba(${layer.spectrumRgb},0.8)`,
                            }}
                          >
                            {item}
                          </span>
                        ))}
                      </div>
                    </motion.div>
                  </div>
                </div>
              </motion.div>
            );
          })}
        </div>

        {/* Bottom tagline */}
        <div className="mt-10 text-center">
          <p className="font-mono text-[10px] uppercase tracking-widest text-white/20">
            Expansion (Redshift) ←— Foundation to Exploration —→ Convergence (Blueshift)
          </p>
        </div>
      </div>
    </section>
  );
}
