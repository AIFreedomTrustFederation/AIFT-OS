import { motion } from "framer-motion";
import { useState } from "react";

const LAYERS = [
  {
    id: "VII",
    name: "Exploration",
    color: "bg-indigo-500/20 text-indigo-400 border-indigo-500/30",
    glow: "shadow-[0_0_30px_rgba(99,102,241,0.2)]",
    items: "Scientific research, Distributed computation, Robotics, Autonomous laboratories, Advanced visualization, Digital twins, Space infrastructure, Simulation environments",
  },
  {
    id: "VI",
    name: "Applications",
    color: "bg-emerald-500/20 text-emerald-400 border-emerald-500/30",
    glow: "shadow-[0_0_30px_rgba(16,185,129,0.2)]",
    items: "DynastyLink, Forge, Capital City Provisions, Publishing systems, Websites, Mobile apps, Community portals, Dashboards, Business platforms",
  },
  {
    id: "V",
    name: "Economy",
    color: "bg-amber-500/20 text-amber-400 border-amber-500/30",
    glow: "shadow-[0_0_30px_rgba(245,158,11,0.2)]",
    items: "Aether Coin, Token systems, Trust accounting, Marketplace, Cooperative exchanges, Asset registries, Treasury systems, Economic governance",
  },
  {
    id: "IV",
    name: "Federation",
    color: "bg-teal-500/20 text-teal-400 border-teal-500/30",
    glow: "shadow-[0_0_30px_rgba(20,184,166,0.2)]",
    quote: "No centralized authority. Every participant remains sovereign.",
    items: "Trust relationships, Federation registries, Service contracts, Capability publication, Synchronization, Discovery, Permissions, Shared governance",
  },
  {
    id: "III",
    name: "Intelligence",
    color: "bg-violet-500/20 text-violet-400 border-violet-500/30",
    glow: "shadow-[0_0_30px_rgba(139,92,246,0.2)]",
    quote: "AI never invents reality. It assists discovery.",
    items: "Local LLM providers, Prompt registries, Agent orchestration, Planning engines, Code analysis, Architecture reasoning, Diagnostics, Knowledge synthesis",
  },
  {
    id: "II",
    name: "Knowledge",
    color: "bg-amber-600/20 text-amber-500 border-amber-600/30",
    glow: "shadow-[0_0_30px_rgba(217,119,6,0.2)]",
    items: "BookSmith AI, Documentation generation, Manuals, API references, Knowledge graphs, Research, Educational content, Publications, Evidence registries",
  },
  {
    id: "I",
    name: "Sovereign Foundation",
    color: "bg-stone-500/20 text-stone-400 border-stone-500/30",
    glow: "shadow-[0_0_30px_rgba(120,113,108,0.2)]",
    items: "Identity, Local storage, Security, Trust, Discovery, Registry, Configuration, Diagnostics, Runtime, Verification",
  },
];

export function LivingLayers() {
  const [hoveredLayer, setHoveredLayer] = useState<string | null>(null);

  return (
    <section className="py-32 px-6 max-w-6xl mx-auto">
      <motion.div
        initial={{ opacity: 0, y: 40 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
        transition={{ duration: 1 }}
        className="text-center mb-24"
      >
        <h2 className="text-4xl md:text-5xl font-bold mb-6 font-sans">The Seven Living Layers</h2>
        <p className="text-muted-foreground font-mono max-w-2xl mx-auto">
          An interactive vertical stack. Every layer is built upon the sovereign truth of the layers beneath it.
        </p>
      </motion.div>

      <div className="flex flex-col gap-4 relative">
        {/* Connection line behind */}
        <div className="absolute left-8 top-0 bottom-0 w-px bg-border z-0" />

        {LAYERS.map((layer, idx) => {
          const isHovered = hoveredLayer === layer.id;
          return (
            <motion.div
              key={layer.id}
              initial={{ opacity: 0, x: -20 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: idx * 0.1 }}
              onHoverStart={() => setHoveredLayer(layer.id)}
              onHoverEnd={() => setHoveredLayer(null)}
              className="relative z-10"
            >
              <div 
                className={`
                  p-6 md:p-8 rounded-lg border transition-all duration-500 cursor-default backdrop-blur-md
                  flex flex-col md:flex-row gap-6 items-start md:items-center
                  ${isHovered ? `bg-card ${layer.glow} border-opacity-50` : 'bg-card/30 border-border'}
                `}
              >
                <div className={`w-16 h-16 shrink-0 flex items-center justify-center rounded-full border font-mono text-xl font-bold transition-colors duration-500 ${isHovered ? layer.color : 'bg-background border-border text-muted-foreground'}`}>
                  {layer.id}
                </div>
                
                <div className="flex-1">
                  <h3 className={`text-2xl font-bold mb-2 transition-colors duration-500 ${isHovered ? layer.color.split(' ')[1] : 'text-foreground'}`}>
                    {layer.name}
                  </h3>
                  
                  {layer.quote && (
                    <p className="text-sm italic text-muted-foreground mb-3 font-serif">"{layer.quote}"</p>
                  )}
                  
                  <motion.div 
                    animate={{ height: isHovered ? "auto" : "0px", opacity: isHovered ? 1 : 0.5 }}
                    className="overflow-hidden"
                  >
                    <p className="font-mono text-sm text-muted-foreground/80 leading-relaxed mt-2">
                      {layer.items}
                    </p>
                  </motion.div>
                </div>
              </div>
            </motion.div>
          );
        })}
      </div>
    </section>
  );
}
