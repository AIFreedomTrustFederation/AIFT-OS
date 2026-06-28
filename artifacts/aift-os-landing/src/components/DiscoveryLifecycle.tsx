import { motion } from "framer-motion";

const STEPS = [
  { name: "Inspect",    color: "#ff2200", rgb: "255,34,0" },
  { name: "Discover",   color: "#ff6600", rgb: "255,102,0" },
  { name: "Understand", color: "#ffaa00", rgb: "255,170,0" },
  { name: "Verify",     color: "#ffdd00", rgb: "255,221,0" },
  { name: "Model",      color: "#88dd00", rgb: "136,221,0" },
  { name: "Plan",       color: "#00dd88", rgb: "0,221,136" },
  { name: "Execute",    color: "#00ccff", rgb: "0,204,255" },
  { name: "Validate",   color: "#0088ff", rgb: "0,136,255" },
  { name: "Document",   color: "#4455ff", rgb: "68,85,255" },
  { name: "Publish",    color: "#8833ee", rgb: "136,51,238" },
  { name: "Learn",      color: "#cc22ff", rgb: "204,34,255" },
];

export function DiscoveryLifecycle() {
  const radius = 200;

  return (
    <section className="py-24 px-6 relative overflow-hidden">
      <div className="absolute inset-0 pointer-events-none">
        <div className="absolute top-0 w-full h-px spectrum-divider opacity-20" />
      </div>

      <div className="max-w-5xl mx-auto flex flex-col items-center">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 1 }}
          className="text-center mb-16"
        >
          <p className="font-mono text-[10px] uppercase tracking-[0.3em] text-white/30 mb-4">
            Standing Wave Resonance · The Eternal Return
          </p>
          <h2
            className="font-cinzel text-4xl md:text-5xl font-bold mb-4 text-glow-gold"
            style={{ color: "#e8c040" }}
          >
            The Discovery Lifecycle
          </h2>
          <p className="font-mono text-sm text-white/40 max-w-2xl mx-auto">
            The operating system never skips steps. Like the torus, every end point is a new beginning. No beginning. No end. Only eternal circulation.
          </p>
        </motion.div>

        {/* Circular node diagram */}
        <div
          className="relative flex items-center justify-center"
          style={{ width: `${radius * 2 + 120}px`, height: `${radius * 2 + 120}px`, maxWidth: "100vw" }}
        >
          {/* SVG background arcs */}
          <svg
            className="absolute inset-0 w-full h-full"
            viewBox={`0 0 ${radius * 2 + 120} ${radius * 2 + 120}`}
          >
            <defs>
              {/* Spectrum gradient around the ring */}
              <linearGradient id="cycleGrad" x1="0%" y1="0%" x2="100%" y2="0%" gradientUnits="objectBoundingBox">
                <stop offset="0%"   stopColor="#ff2200" />
                <stop offset="18%"  stopColor="#ff8800" />
                <stop offset="36%"  stopColor="#ffdd00" />
                <stop offset="50%"  stopColor="#00dd88" />
                <stop offset="64%"  stopColor="#00ccff" />
                <stop offset="82%"  stopColor="#4455ff" />
                <stop offset="100%" stopColor="#cc22ff" />
              </linearGradient>
              <filter id="nodeGlow">
                <feGaussianBlur stdDeviation="3" result="blur" />
                <feMerge><feMergeNode in="blur" /><feMergeNode in="SourceGraphic" /></feMerge>
              </filter>
            </defs>

            {/* Outer ring */}
            <circle
              cx={radius + 60}
              cy={radius + 60}
              r={radius}
              fill="none"
              stroke="rgba(255,255,255,0.06)"
              strokeWidth="1"
            />

            {/* Spectrum ring — animated dash-offset to show circulation */}
            <motion.circle
              cx={radius + 60}
              cy={radius + 60}
              r={radius}
              fill="none"
              stroke="url(#cycleGrad)"
              strokeWidth="1.5"
              strokeDasharray={`${2 * Math.PI * radius}`}
              initial={{ strokeDashoffset: 2 * Math.PI * radius }}
              whileInView={{ strokeDashoffset: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 3, ease: "easeOut" }}
              opacity={0.5}
            />

            {/* Inner dashed orbit */}
            <circle
              cx={radius + 60}
              cy={radius + 60}
              r={radius * 0.4}
              fill="none"
              stroke="rgba(255,255,255,0.04)"
              strokeWidth="1"
              strokeDasharray="4 8"
            />

            {/* Spoke lines from center to each node */}
            {STEPS.map((step, idx) => {
              const angle = (idx / STEPS.length) * Math.PI * 2 - Math.PI / 2;
              const nx = (radius + 60) + Math.cos(angle) * radius;
              const ny = (radius + 60) + Math.sin(angle) * radius;
              return (
                <line
                  key={step.name}
                  x1={radius + 60} y1={radius + 60}
                  x2={nx} y2={ny}
                  stroke={`rgba(${step.rgb},0.08)`}
                  strokeWidth="1"
                />
              );
            })}
          </svg>

          {/* Step nodes */}
          {STEPS.map((step, idx) => {
            const angle = (idx / STEPS.length) * Math.PI * 2 - Math.PI / 2;
            const nx = Math.cos(angle) * radius;
            const ny = Math.sin(angle) * radius;

            return (
              <motion.div
                key={step.name}
                className="absolute flex flex-col items-center"
                style={{
                  left: "50%",
                  top: "50%",
                  transform: `translate(-50%, -50%)`,
                }}
                initial={{ opacity: 0, x: 0, y: 0 }}
                whileInView={{ opacity: 1, x: nx, y: ny }}
                viewport={{ once: true }}
                transition={{ duration: 0.8, delay: idx * 0.08, type: "spring", stiffness: 60 }}
              >
                {/* Node dot */}
                <motion.div
                  className="rounded-full mb-1.5 relative"
                  animate={{
                    scale: [1, 1.4, 1],
                    opacity: [0.7, 1, 0.7],
                  }}
                  transition={{
                    duration: 2.5,
                    repeat: Infinity,
                    delay: idx * (2.5 / STEPS.length),
                    ease: "easeInOut",
                  }}
                  style={{
                    width: 10,
                    height: 10,
                    background: step.color,
                    boxShadow: `0 0 12px ${step.color}, 0 0 24px ${step.color}66`,
                  }}
                />
                {/* Label */}
                <span
                  className="font-mono text-[9px] uppercase tracking-wider whitespace-nowrap"
                  style={{ color: `rgba(${step.rgb},0.7)` }}
                >
                  {step.name}
                </span>
                {/* Step number */}
                <span className="font-mono text-[8px] text-white/30">{String(idx + 1).padStart(2, "0")}</span>
              </motion.div>
            );
          })}

          {/* Center pulsing core */}
          <div className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 flex items-center justify-center">
            <motion.div
              animate={{ rotate: 360 }}
              transition={{ duration: 30, repeat: Infinity, ease: "linear" }}
              className="absolute rounded-full"
              style={{
                width: radius * 0.8,
                height: radius * 0.8,
                border: "1px dashed rgba(255,255,255,0.06)",
              }}
            />
            <motion.div
              animate={{ rotate: -360 }}
              transition={{ duration: 20, repeat: Infinity, ease: "linear" }}
              className="absolute rounded-full"
              style={{
                width: radius * 0.55,
                height: radius * 0.55,
                border: "1px dashed rgba(255,255,255,0.04)",
              }}
            />
            <div className="flex flex-col items-center text-center">
              <div
                className="w-12 h-12 rounded-full mb-1"
                style={{
                  background: "radial-gradient(circle, rgba(0,221,136,0.2) 0%, transparent 70%)",
                  border: "1px solid rgba(0,221,136,0.15)",
                  boxShadow: "0 0 30px rgba(0,221,136,0.1)",
                }}
              />
              <span className="font-mono text-[8px] uppercase tracking-widest text-white/40 -mt-10 pt-3">
                Eternal<br/>Circulation
              </span>
            </div>
          </div>
        </div>

        {/* Notes */}
        <div className="mt-16 grid grid-cols-1 md:grid-cols-3 gap-6 max-w-3xl w-full">
          {[
            { label: "No Beginning", desc: "The cycle has no canonical start. Any node can begin an inspection." },
            { label: "No End", desc: "Learn feeds back into Inspect. The field renews eternally." },
            { label: "No Skipping", desc: "The operating system enforces the full sequence. Evidence gates each transition." },
          ].map((n) => (
            <div
              key={n.label}
              className="p-4 rounded-xl text-center"
              style={{ border: "1px solid rgba(255,255,255,0.08)", background: "rgba(255,255,255,0.02)" }}
            >
              <p className="font-cinzel text-sm font-bold text-white/60 mb-2">{n.label}</p>
              <p className="font-mono text-[10px] text-white/40 leading-relaxed">{n.desc}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
