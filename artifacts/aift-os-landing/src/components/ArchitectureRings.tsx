import { motion } from "framer-motion";
import { useState } from "react";

const LAYERS = [
  { id: 0, name: "Constitutional Core", radius: 50, strokeWidth: 2, dashArray: "0" },
  { id: 1, name: "Council of 24 Elders", radius: 100, strokeWidth: 1, dashArray: "4 4", nodes: 24 },
  { id: 2, name: "12 Validator Orders", radius: 160, strokeWidth: 2, dashArray: "0", nodes: 12 },
  { id: 3, name: "Federation Runtime & Services", radius: 230, strokeWidth: 1, dashArray: "2 6", nodes: 10 },
  { id: 4, name: "Sovereign Federation Nodes", radius: 310, strokeWidth: 1.5, dashArray: "10 5", nodes: 6 },
  { id: 5, name: "Individual Sovereign Trees", radius: 400, strokeWidth: 0.5, dashArray: "1 8", nodes: 36 },
];

const LAYER_DETAILS = {
  0: { title: "Constitutional Core", desc: "Truth before authority. Evidence before claims. Discoverable governance." },
  1: { title: "Council of 24 Elders", desc: "Constitutional review and stewardship. Human + AI participants. Membership is role-based, transparent, evidence-driven." },
  2: { title: "12 Validator Orders", desc: "Vision · Wisdom · Reconciliation · Treasury · Ledger · Judgment · Timekeeping · Equity · Faith · Renewal · Intercession · Sovereignty. Cryptographic identity and evidence trail." },
  3: { title: "Federation Runtime & Services", desc: "Discovery · Registry · Runtime · Scheduler · Event Bus · Documentation · Intelligence · Patch Engine · Architecture · Diagnostics" },
  4: { title: "Sovereign Federation Nodes", desc: "BookSmith AI · AIFT Forge · Aether Coin · DynastyLink · Capital City Provisions · AIFT-OS" },
  5: { title: "Individual Sovereign Trees", desc: "Each user owns a complete local operating system. Private, shared, or public." },
};

export function ArchitectureRings() {
  const [activeLayer, setActiveLayer] = useState<number | null>(null);

  const center = 500;

  return (
    <section className="py-32 px-6 relative overflow-hidden min-h-[1000px] flex flex-col items-center">
      <div className="max-w-4xl text-center mb-16 relative z-10">
        <h2 className="text-4xl md:text-5xl font-bold mb-6 text-glow-teal text-foreground">The Governance Architecture</h2>
        <p className="text-xl text-muted-foreground font-mono">
          Wheel within wheels. Inspect the layers of the federation.
        </p>
      </div>

      <div className="relative w-[1000px] h-[1000px] max-w-full flex-shrink-0 flex items-center justify-center">
        {/* Info Panel Overlay */}
        <div className="absolute top-0 right-0 md:-right-20 z-20 w-80 pointer-events-none">
          <motion.div
            animate={{ opacity: activeLayer !== null ? 1 : 0, y: activeLayer !== null ? 0 : 20 }}
            className="p-6 bg-card/80 backdrop-blur-xl border border-primary/30 rounded-xl shadow-[0_0_30px_rgba(245,166,35,0.1)] pointer-events-auto"
          >
            {activeLayer !== null && (
              <>
                <div className="text-primary font-mono text-xs mb-2">LAYER {activeLayer}</div>
                <h3 className="text-xl font-bold mb-3">{LAYER_DETAILS[activeLayer as keyof typeof LAYER_DETAILS].title}</h3>
                <p className="text-muted-foreground text-sm leading-relaxed">
                  {LAYER_DETAILS[activeLayer as keyof typeof LAYER_DETAILS].desc}
                </p>
              </>
            )}
          </motion.div>
        </div>

        <svg viewBox="0 0 1000 1000" className="w-full h-full absolute inset-0">
          <defs>
            <radialGradient id="coreGlow" cx="50%" cy="50%" r="50%">
              <stop offset="0%" stopColor="hsl(var(--primary))" stopOpacity="0.4" />
              <stop offset="100%" stopColor="hsl(var(--primary))" stopOpacity="0" />
            </radialGradient>
            <filter id="glow">
              <feGaussianBlur stdDeviation="4" result="coloredBlur"/>
              <feMerge>
                <feMergeNode in="coloredBlur"/>
                <feMergeNode in="SourceGraphic"/>
              </feMerge>
            </filter>
          </defs>

          {/* Ambient Core Glow */}
          <circle cx={center} cy={center} r={100} fill="url(#coreGlow)" />

          {LAYERS.map((layer, index) => {
            const isActive = activeLayer === layer.id;
            const isHovered = activeLayer === layer.id;
            
            // Generate nodes for this ring
            const nodes = [];
            if (layer.nodes) {
              const angleStep = (Math.PI * 2) / layer.nodes;
              for (let i = 0; i < layer.nodes; i++) {
                const angle = i * angleStep;
                const x = center + Math.cos(angle) * layer.radius;
                const y = center + Math.sin(angle) * layer.radius;
                nodes.push(
                  <circle
                    key={`node-${layer.id}-${i}`}
                    cx={x}
                    cy={y}
                    r={isActive ? 4 : 2}
                    fill={isActive ? "hsl(var(--primary))" : "hsl(var(--muted-foreground))"}
                    className="transition-all duration-300"
                  />
                );
              }
            }

            return (
              <g 
                key={layer.id} 
                className="cursor-pointer group"
                onClick={() => setActiveLayer(isActive ? null : layer.id)}
                onMouseEnter={() => setActiveLayer(layer.id)}
                onMouseLeave={() => setActiveLayer(null)}
              >
                {/* Invisible hit area */}
                <circle
                  cx={center}
                  cy={center}
                  r={layer.radius + 20}
                  fill="transparent"
                  stroke="transparent"
                  strokeWidth="40"
                />
                
                <motion.g
                  animate={{ rotate: index % 2 === 0 ? 360 : -360 }}
                  transition={{ duration: 100 + (index * 20), repeat: Infinity, ease: "linear" }}
                  style={{ originX: "500px", originY: "500px" }}
                >
                  <circle
                    cx={center}
                    cy={center}
                    r={layer.radius}
                    fill="transparent"
                    stroke={isActive ? "hsl(var(--primary))" : "hsl(var(--border))"}
                    strokeWidth={isActive ? layer.strokeWidth * 2 : layer.strokeWidth}
                    strokeDasharray={layer.dashArray}
                    className="transition-all duration-500"
                    filter={isActive ? "url(#glow)" : ""}
                  />
                  {nodes}
                </motion.g>
              </g>
            );
          })}
        </svg>
      </div>
    </section>
  );
}
