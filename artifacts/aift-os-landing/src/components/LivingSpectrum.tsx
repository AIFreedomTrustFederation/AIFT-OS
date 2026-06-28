import { motion } from "framer-motion";

const STATES = [
  { status: "Unknown", color: "bg-stone-600", border: "border-stone-500", text: "text-stone-400", desc: "No evidence found" },
  { status: "Planned", color: "bg-emerald-900", border: "border-emerald-800", text: "text-emerald-500/70", desc: "Intention declared" },
  { status: "Detected", color: "bg-amber-500", border: "border-amber-400", text: "text-amber-400", desc: "Evidence found, not yet verified", glow: "shadow-[0_0_15px_rgba(245,158,11,0.5)]" },
  { status: "Ready", color: "bg-emerald-500", border: "border-emerald-400", text: "text-emerald-400", desc: "Verified and capable", glow: "shadow-[0_0_15px_rgba(16,185,129,0.5)]" },
  { status: "Active", color: "bg-cyan-500", border: "border-cyan-400", text: "text-cyan-400", desc: "Currently operating", glow: "shadow-[0_0_20px_rgba(6,182,212,0.8)]" },
  { status: "Blocked", color: "bg-orange-500", border: "border-orange-400", text: "text-orange-400", desc: "Verified but unable to proceed", glow: "shadow-[0_0_15px_rgba(249,115,22,0.5)]" },
  { status: "Deprecated", color: "bg-violet-500", border: "border-violet-400", text: "text-violet-400", desc: "Still running, marked for removal", glow: "shadow-[0_0_15px_rgba(139,92,246,0.5)]" },
  { status: "Removed", color: "bg-black", border: "border-stone-800", text: "text-stone-600", desc: "No longer present" },
];

export function LivingSpectrum() {
  return (
    <section className="py-32 px-6 bg-black/50 border-y border-border relative overflow-hidden">
      {/* Ambient background glow */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-full h-[500px] bg-gradient-to-r from-emerald-500/5 via-cyan-500/5 to-violet-500/5 blur-[100px] pointer-events-none" />

      <div className="max-w-6xl mx-auto relative z-10">
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 1 }}
          className="text-center mb-20"
        >
          <h2 className="text-4xl md:text-5xl font-bold mb-6 font-sans">The Living Spectrum</h2>
          <p className="text-xl text-muted-foreground font-mono max-w-3xl mx-auto">
            The operating system communicates state through color. Every visible color is backed by evidence. 
            <span className="text-foreground block mt-4 font-sans italic">Nothing glows without verification.</span>
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {STATES.map((state, idx) => (
            <motion.div
              key={state.status}
              initial={{ opacity: 0, scale: 0.95 }}
              whileInView={{ opacity: 1, scale: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: idx * 0.1 }}
              className="group p-6 rounded-xl border border-border bg-card/40 backdrop-blur-sm hover:border-border/80 transition-colors"
            >
              <div className="flex items-center gap-4 mb-4">
                <div className="relative flex items-center justify-center">
                  {state.glow && (
                    <div className={`absolute inset-0 rounded-full ${state.color} opacity-40 blur-md ${state.glow}`} />
                  )}
                  <div className={`w-4 h-4 rounded-full border-2 ${state.color} ${state.border} relative z-10`} />
                </div>
                <h3 className={`font-mono font-bold text-lg tracking-wide ${state.text}`}>
                  {state.status}
                </h3>
              </div>
              <p className="text-sm text-muted-foreground font-sans">
                {state.desc}
              </p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
