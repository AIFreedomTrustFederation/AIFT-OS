import { motion } from "framer-motion";

// Eight OS states mapped to the visible spectrum + edges
const STATES = [
  {
    status: "Unknown",
    hue: null,
    color: "#444444",
    rgb: "68,68,68",
    border: "rgba(80,80,80,0.4)",
    desc: "No evidence found. The field has not yet differentiated.",
    phase: "Void · Pre-Differentiation",
  },
  {
    status: "Planned",
    hue: 0,
    color: "#ff2200",
    rgb: "255,34,0",
    border: "rgba(255,34,0,0.4)",
    desc: "Intention declared. Root-level commitment anchored in the registry.",
    phase: "Red · Foundation · Survival",
  },
  {
    status: "Detected",
    hue: 30,
    color: "#ff8800",
    rgb: "255,136,0",
    border: "rgba(255,136,0,0.4)",
    desc: "Evidence found. Not yet verified. The field is differentiating.",
    phase: "Orange · Creation · Energy",
  },
  {
    status: "Ready",
    hue: 55,
    color: "#ffdd00",
    rgb: "255,221,0",
    border: "rgba(255,221,0,0.4)",
    desc: "Verified and capable. Solar plexus active — will, power, readiness.",
    phase: "Yellow · Intellect · Discernment",
  },
  {
    status: "Active",
    hue: 145,
    color: "#00dd55",
    rgb: "0,221,85",
    border: "rgba(0,221,85,0.5)",
    desc: "Currently operating. The observer is present. Equilibrium holds.",
    phase: "Green · Equilibrium · The Observer",
  },
  {
    status: "Blocked",
    hue: 185,
    color: "#00ccff",
    rgb: "0,204,255",
    border: "rgba(0,204,255,0.4)",
    desc: "Verified but unable to proceed. Coherence reached, waiting for flow.",
    phase: "Cyan · Harmony · Communication",
  },
  {
    status: "Deprecated",
    hue: 260,
    color: "#9922ff",
    rgb: "153,34,255",
    border: "rgba(153,34,255,0.4)",
    desc: "Still running, marked for removal. Converging toward the coherence horizon.",
    phase: "Violet · Transcendence · High Order",
  },
  {
    status: "Removed",
    hue: null,
    color: "#ffffff",
    rgb: "255,255,255",
    border: "rgba(255,255,255,0.3)",
    desc: "No longer present. All spectra merged. Complete phase coherence — the Toroidal Seam.",
    phase: "White · All Spectra Merge · Complete Coherence",
  },
];

export function LivingSpectrum() {
  return (
    <section className="py-24 px-6 relative overflow-hidden">
      {/* Spectrum gradient bg stripe */}
      <div
        className="absolute inset-x-0 top-1/2 -translate-y-1/2 h-[300px] pointer-events-none"
        style={{
          background:
            "linear-gradient(90deg, rgba(255,34,0,0.03), rgba(255,136,0,0.03), rgba(255,221,0,0.03), rgba(0,221,85,0.03), rgba(0,204,255,0.03), rgba(51,85,255,0.03), rgba(153,34,255,0.03))",
          filter: "blur(40px)",
        }}
      />

      <div className="max-w-6xl mx-auto relative z-10">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 1 }}
          className="text-center mb-4"
        >
          <p className="font-mono text-[10px] uppercase tracking-[0.3em] text-white/30 mb-4">
            Color / Phase Map · Clockwise Around the Torus
          </p>
          <h2
            className="font-cinzel text-4xl md:text-5xl font-bold mb-4 text-glow-gold"
            style={{ color: "#e8c040" }}
          >
            The Living Spectrum
          </h2>
          <p className="font-mono text-sm text-white/40 max-w-2xl mx-auto">
            Every color is a phase of the same field. The operating system
            communicates state through coherence, not labels.{" "}
            <span className="text-white/60 italic">
              Nothing glows without verification.
            </span>
          </p>
        </motion.div>

        {/* Horizontal spectrum bar */}
        <div className="my-10 mx-auto max-w-3xl">
          <div
            className="h-1.5 rounded-full w-full"
            style={{
              background:
                "linear-gradient(90deg, #444444 0%, #ff2200 12%, #ff8800 25%, #ffdd00 38%, #00dd55 50%, #00ccff 63%, #9922ff 78%, #ffffff 100%)",
              boxShadow: "0 0 20px rgba(255,255,255,0.1)",
            }}
          />
          <div className="flex justify-between mt-2">
            {[
              "Unknown",
              "Planned",
              "Detected",
              "Ready",
              "Active",
              "Blocked",
              "Deprecated",
              "Removed",
            ].map((s) => (
              <span
                key={s}
                className="font-mono text-[8px] text-white/20 uppercase"
              >
                {s}
              </span>
            ))}
          </div>
        </div>

        {/* State cards */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {STATES.map((state, idx) => (
            <motion.div
              key={state.status}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: idx * 0.08 }}
              className="group relative rounded-xl p-5 transition-all duration-500 overflow-hidden cursor-default"
              style={{
                border: state.border,
                background: `rgba(${state.rgb},0.04)`,
              }}
            >
              {/* Left glow bar */}
              <div
                className="absolute left-0 top-0 bottom-0 w-0.5 transition-all duration-500 group-hover:w-1"
                style={{
                  background: state.color,
                  boxShadow: `0 0 8px ${state.color}`,
                }}
              />

              {/* Status indicator */}
              <div className="flex items-center gap-3 mb-3">
                <div className="relative flex items-center justify-center">
                  <div
                    className="w-4 h-4 rounded-full"
                    style={{
                      background: state.color,
                      boxShadow:
                        state.status === "Removed"
                          ? `0 0 20px rgba(255,255,255,0.5), 0 0 40px rgba(255,255,255,0.2)`
                          : state.status === "Unknown"
                            ? "none"
                            : `0 0 12px ${state.color}, 0 0 24px ${state.color}66`,
                    }}
                  />
                </div>
                <h3
                  className="font-cinzel font-bold text-base tracking-wide"
                  style={{ color: state.color }}
                >
                  {state.status}
                </h3>
              </div>

              <p
                className="font-mono text-[9px] uppercase tracking-widest mb-2"
                style={{ color: `rgba(${state.rgb},0.5)` }}
              >
                {state.phase}
              </p>

              <p className="text-xs text-white/40 leading-relaxed">
                {state.desc}
              </p>
            </motion.div>
          ))}
        </div>

        {/* Bottom note */}
        <div className="mt-10 text-center">
          <p className="font-mono text-[10px] text-white/20 uppercase tracking-widest">
            All colors are present at every point · They differ only in dominant
            phase
          </p>
        </div>
      </div>
    </section>
  );
}
