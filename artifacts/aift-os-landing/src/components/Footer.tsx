import { motion } from "framer-motion";
import { Github } from "lucide-react";

// Sacred geometry symbols as simple SVG paths
function TorusIcon({ size = 60 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 60 60" fill="none">
      <ellipse
        cx="30"
        cy="30"
        rx="26"
        ry="26"
        stroke="rgba(255,255,255,0.15)"
        strokeWidth="1"
      />
      <ellipse
        cx="30"
        cy="30"
        rx="18"
        ry="18"
        stroke="rgba(255,255,255,0.10)"
        strokeWidth="0.8"
      />
      <ellipse
        cx="30"
        cy="30"
        rx="8"
        ry="8"
        stroke="rgba(255,255,255,0.08)"
        strokeWidth="0.6"
      />
      <ellipse
        cx="30"
        cy="30"
        rx="26"
        ry="10"
        stroke="rgba(0,221,136,0.25)"
        strokeWidth="1"
      />
      <ellipse
        cx="30"
        cy="30"
        rx="18"
        ry="7"
        stroke="rgba(0,221,136,0.15)"
        strokeWidth="0.7"
      />
    </svg>
  );
}

function FlowerOfLifeIcon({ size = 60 }: { size?: number }) {
  const r = 10;
  const cx = 30,
    cy = 30;
  const angles = Array.from({ length: 6 }, (_, i) => (i * Math.PI) / 3);
  return (
    <svg width={size} height={size} viewBox="0 0 60 60" fill="none">
      <circle
        cx={cx}
        cy={cy}
        r={r}
        stroke="rgba(255,200,50,0.2)"
        strokeWidth="0.7"
      />
      {angles.map((a, i) => (
        <circle
          key={i}
          cx={cx + Math.cos(a) * r}
          cy={cy + Math.sin(a) * r}
          r={r}
          stroke="rgba(255,200,50,0.15)"
          strokeWidth="0.7"
        />
      ))}
      <circle
        cx={cx}
        cy={cy}
        r={r * 2}
        stroke="rgba(255,200,50,0.08)"
        strokeWidth="0.5"
      />
    </svg>
  );
}

function SpiralIcon({ size = 60 }: { size?: number }) {
  // Golden spiral approximation
  const path =
    "M30,30 Q35,25 38,30 Q41,37 35,42 Q27,47 20,40 Q11,30 18,20 Q27,8 40,15 Q52,23 50,38";
  return (
    <svg width={size} height={size} viewBox="0 0 60 60" fill="none">
      <path
        d={path}
        stroke="rgba(153,34,255,0.3)"
        strokeWidth="1"
        fill="none"
      />
      <circle cx="30" cy="30" r="2" fill="rgba(153,34,255,0.4)" />
    </svg>
  );
}

const SACRED_GEOMETRY = [
  { label: "Torus", sublabel: "(Circulation)", Icon: TorusIcon },
  { label: "Flower of Life", sublabel: "(Unity)", Icon: FlowerOfLifeIcon },
  { label: "Golden Spiral", sublabel: "(Harmonic Growth)", Icon: SpiralIcon },
];

export function Footer() {
  return (
    <footer className="relative overflow-hidden pt-24 pb-16 px-6">
      <div className="absolute top-0 w-full h-px spectrum-divider opacity-40" />

      {/* Background ambient glow */}
      <div className="absolute inset-0 pointer-events-none">
        <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-[800px] h-[400px] rounded-full bg-yellow-500/3 blur-[150px]" />
      </div>

      <div className="max-w-5xl mx-auto relative z-10">
        {/* Sacred geometry row */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 1 }}
          className="flex justify-center gap-10 md:gap-20 mb-20"
        >
          {SACRED_GEOMETRY.map(({ label, sublabel, Icon }) => (
            <div key={label} className="flex flex-col items-center gap-2">
              <Icon size={56} />
              <span className="font-mono text-[9px] text-white/30 uppercase tracking-widest">
                {label}
              </span>
              <span className="font-mono text-[8px] text-white/15">
                {sublabel}
              </span>
            </div>
          ))}
        </motion.div>

        {/* Main text */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 1, delay: 0.2 }}
          className="text-center"
        >
          <h2
            className="font-cinzel text-3xl md:text-5xl font-bold mb-8 text-glow-gold"
            style={{ color: "#e8c040" }}
          >
            One Manifold. Infinite Projections.
          </h2>

          <div className="max-w-3xl mx-auto mb-6">
            <p className="text-white/60 font-mono text-sm leading-relaxed mb-4">
              Spacetime, light, consciousness, information, and causality are
              different expressions of one continuous, self-folding topology.
            </p>
            <p className="text-white/50 font-mono text-sm leading-relaxed mb-4">
              The long-term objective is not another software platform. The
              objective is a truthful digital ecosystem where every sovereign
              installation contributes to a larger federation without
              surrendering ownership of its own knowledge.
            </p>
            <p className="text-white/45 font-mono text-sm leading-relaxed">
              Repositories become branches. Capabilities become fruit.
              Documentation becomes memory. Events become circulation.
              Artificial intelligence becomes assistance rather than authority.
            </p>
          </div>

          {/* Core principles grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3 max-w-2xl mx-auto mb-14 text-left">
            {[
              "The Observer is not separate from the field.",
              "The present moment (green) is the only stable intersection of all flows.",
              "All information is conserved and recirculated.",
              "The holographic principle: every part contains the whole.",
              "Geometry is the language of consciousness.",
              "Truth before authority. Evidence before claims.",
            ].map((p) => (
              <div
                key={p}
                className="flex items-start gap-2 p-3 rounded-lg"
                style={{
                  background: "rgba(255,255,255,0.02)",
                  border: "1px solid rgba(255,255,255,0.07)",
                }}
              >
                <span className="text-yellow-400/60 mt-0.5 shrink-0">·</span>
                <p className="font-mono text-[10px] text-white/50 leading-relaxed">
                  {p}
                </p>
              </div>
            ))}
          </div>

          {/* CTA */}
          <a
            href="https://github.com/AIFreedomTrustFederation"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-3 px-8 py-4 rounded-full font-mono font-bold text-sm transition-all duration-300 hover:scale-105 active:scale-95"
            style={{
              background:
                "linear-gradient(135deg, rgba(255,200,50,0.15), rgba(255,180,20,0.08))",
              border: "1px solid rgba(255,200,50,0.35)",
              color: "#e8c040",
              boxShadow: "0 0 30px rgba(255,200,50,0.1)",
            }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLElement).style.boxShadow =
                "0 0 50px rgba(255,200,50,0.2)";
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLElement).style.boxShadow =
                "0 0 30px rgba(255,200,50,0.1)";
            }}
          >
            <Github className="w-4 h-4" />
            Inspect the Source
          </a>
        </motion.div>

        {/* Bottom signature */}
        <div className="mt-20 flex flex-col items-center gap-3">
          <div className="spectrum-divider w-64" />
          <p className="font-mono text-[9px] uppercase tracking-[0.4em] text-white/35">
            AIFT-OS · Truth Before Authority · Every Point Contains The Whole
          </p>
        </div>
      </div>
    </footer>
  );
}
