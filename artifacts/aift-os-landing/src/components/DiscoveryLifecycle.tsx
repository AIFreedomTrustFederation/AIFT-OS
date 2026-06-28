import { motion } from "framer-motion";

const STEPS = [
  "Inspect", "Discover", "Understand", "Verify", 
  "Model", "Plan", "Execute", "Validate", 
  "Document", "Publish", "Learn"
];

export function DiscoveryLifecycle() {
  return (
    <section className="py-32 px-6 max-w-6xl mx-auto flex flex-col items-center">
      <div className="text-center mb-24">
        <h2 className="text-4xl md:text-5xl font-bold mb-6 font-sans">The Discovery Lifecycle</h2>
        <p className="text-xl text-muted-foreground font-mono">
          The operating system never skips steps.
        </p>
      </div>

      <div className="relative w-[300px] h-[300px] md:w-[500px] md:h-[500px]">
        {STEPS.map((step, idx) => {
          const angle = (idx / STEPS.length) * Math.PI * 2 - Math.PI / 2;
          const radius = 180; // responsive radius would be better, keeping fixed for simplicity
          const x = Math.cos(angle) * radius;
          const y = Math.sin(angle) * radius;
          
          return (
            <motion.div
              key={step}
              className="absolute left-1/2 top-1/2 flex flex-col items-center justify-center -ml-12 -mt-12 w-24 h-24"
              animate={{ 
                x, y,
              }}
              initial={{ x: 0, y: 0, opacity: 0 }}
              whileInView={{ opacity: 1 }}
              transition={{ duration: 1, delay: idx * 0.1, type: "spring" }}
            >
              <motion.div 
                className="w-3 h-3 rounded-full bg-secondary shadow-[0_0_15px_rgba(20,184,166,0.8)] mb-2 relative"
                animate={{ scale: [1, 1.5, 1], opacity: [0.5, 1, 0.5] }}
                transition={{ duration: 2, repeat: Infinity, delay: idx * (2 / STEPS.length) }}
              >
                {/* Connecting arc hint - CSS rotation trick */}
              </motion.div>
              <span className="font-mono text-xs text-foreground/80 whitespace-nowrap">
                {step}
              </span>
            </motion.div>
          );
        })}
        
        {/* Center pulsing core */}
        <div className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2">
          <motion.div
            animate={{ rotate: 360 }}
            transition={{ duration: 20, repeat: Infinity, ease: "linear" }}
            className="w-32 h-32 rounded-full border border-secondary/30 border-dashed flex items-center justify-center"
          >
            <div className="w-16 h-16 rounded-full bg-secondary/10 flex items-center justify-center blur-sm" />
          </motion.div>
        </div>
      </div>
    </section>
  );
}
