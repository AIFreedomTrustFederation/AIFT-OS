import { motion } from "framer-motion";
import { useEffect, useRef } from "react";

export function TreeVisualizer({
  type = "life",
  width = 420,
  height = 580,
}: {
  type: "life" | "knowledge";
  width?: number;
  height?: number;
}) {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    // Capture non-null context immediately — TypeScript cannot narrow across closure boundaries
    const gfx: CanvasRenderingContext2D = ctx;

    const dpr = window.devicePixelRatio || 1;
    canvas.width = width * dpr;
    canvas.height = height * dpr;
    gfx.scale(dpr, dpr);
    canvas.style.width = `${width}px`;
    canvas.style.height = `${height}px`;

    let raf: number;
    let time = 0;
    const isLife = type === "life";

    // Tree of Life: cyan→green spectrum (equilibrium / federation side)
    // Tree of Knowledge: amber→violet spectrum (knowledge / convergence side)
    function branchColor(depth: number, maxDepth: number, t: number): string {
      const p = depth / maxDepth; // 1=trunk, 0=tip
      if (isLife) {
        // Root is red-orange, tips are cyan-blue (expansion → equilibrium)
        const hue = 10 + (1 - p) * 170; // 10=red-orange → 180=cyan
        const alpha = 0.2 + p * 0.5 + Math.sin(t * 0.04 + depth) * 0.15;
        return `hsla(${hue}, 90%, 60%, ${Math.min(1, alpha)})`;
      } else {
        // Knowledge: amber at root, violet at tips (history → insight)
        const hue = 40 + (1 - p) * 250; // 40=amber → 280=violet
        const alpha = Math.min(0.85, 0.2 + p * 0.55);
        return `hsla(${hue}, 90%, 62%, ${alpha})`;
      }
    }

    function leafColor(depth: number, t: number): string {
      if (isLife) {
        const hue = 140 + Math.sin(t * 0.03 + depth) * 40;
        return `hsla(${hue}, 100%, 65%, 0.9)`;
      } else {
        const hue = 270 + depth * 5;
        return `hsla(${hue}, 90%, 70%, 0.8)`;
      }
    }

    const maxDepth = isLife ? 6 : 8;
    const branchWidthMult = isLife ? 0.9 : 0.65;

    function drawBranch(
      startX: number,
      startY: number,
      len: number,
      angle: number,
      depth: number,
      t: number,
    ) {
      if (depth === 0 || len < 1) return;

      const endX = startX + len * Math.cos(angle);
      const endY = startY + len * Math.sin(angle);

      gfx.beginPath();
      gfx.moveTo(startX, startY);
      gfx.lineTo(endX, endY);
      gfx.lineWidth = Math.max(0.3, depth * branchWidthMult);
      gfx.strokeStyle = branchColor(depth, maxDepth, t);
      gfx.shadowBlur = depth === maxDepth ? 0 : isLife ? 4 : 2;
      gfx.shadowColor = branchColor(depth, maxDepth, t);
      gfx.stroke();
      gfx.shadowBlur = 0;

      // Leaf / node at tips
      if (depth <= 2) {
        gfx.beginPath();
        gfx.arc(endX, endY, isLife ? 2.5 : 1.8, 0, Math.PI * 2);
        gfx.fillStyle = leafColor(depth, t);
        gfx.shadowColor = leafColor(depth, t);
        gfx.shadowBlur = 8;
        gfx.fill();
        gfx.shadowBlur = 0;
      }

      // Subtle glow node at junction
      if (depth === maxDepth) {
        const grad = gfx.createRadialGradient(
          startX,
          startY,
          0,
          startX,
          startY,
          12,
        );
        grad.addColorStop(
          0,
          isLife ? "rgba(0,220,180,0.15)" : "rgba(255,180,50,0.12)",
        );
        grad.addColorStop(1, "transparent");
        gfx.beginPath();
        gfx.arc(startX, startY, 12, 0, Math.PI * 2);
        gfx.fillStyle = grad;
        gfx.fill();
      }

      // Sway animation
      const sway = Math.sin(t * 0.008 + depth * 0.4) * (isLife ? 0.06 : 0.02);

      drawBranch(endX, endY, len * 0.7, angle - 0.48 + sway, depth - 1, t);
      drawBranch(endX, endY, len * 0.7, angle + 0.48 + sway, depth - 1, t);
      // Deterministic middle branch for knowledge depth
      if (!isLife && depth % 2 === 0) {
        drawBranch(endX, endY, len * 0.58, angle + sway * 2, depth - 1, t);
      }
    }

    function render() {
      gfx.clearRect(0, 0, width, height);

      // Background radial glow at base
      const baseGrad = gfx.createRadialGradient(
        width / 2,
        height - 10,
        0,
        width / 2,
        height - 10,
        80,
      );
      baseGrad.addColorStop(
        0,
        isLife ? "rgba(0,200,150,0.08)" : "rgba(255,160,40,0.08)",
      );
      baseGrad.addColorStop(1, "transparent");
      gfx.beginPath();
      gfx.arc(width / 2, height - 10, 80, 0, Math.PI * 2);
      gfx.fillStyle = baseGrad;
      gfx.fill();

      drawBranch(
        width / 2,
        height - 20,
        height * 0.26,
        -Math.PI / 2,
        maxDepth,
        time,
      );

      time++;
      if (isLife) {
        raf = requestAnimationFrame(render);
      }
    }

    render();
    return () => {
      if (raf) cancelAnimationFrame(raf);
    };
  }, [type, width, height]);

  const lifeColor = "#00ccaa";
  const knowledgeColor = "#cc88ff";

  return (
    <div className="flex flex-col items-center">
      <div className="mb-4 text-center">
        {type === "life" ? (
          <>
            <h3
              className="font-cinzel text-xl font-bold mb-1"
              style={{
                color: lifeColor,
                textShadow: `0 0 20px ${lifeColor}66`,
              }}
            >
              Tree of Life
            </h3>
            <p className="font-mono text-[10px] uppercase tracking-widest text-white/30">
              Expansion Phase · "What is alive and running?"
            </p>
          </>
        ) : (
          <>
            <h3
              className="font-cinzel text-xl font-bold mb-1"
              style={{
                color: knowledgeColor,
                textShadow: `0 0 20px ${knowledgeColor}66`,
              }}
            >
              Tree of Knowledge
            </h3>
            <p className="font-mono text-[10px] uppercase tracking-widest text-white/30">
              Convergence Phase · "What has been learned?"
            </p>
          </>
        )}
      </div>

      <div
        className="relative rounded-xl overflow-hidden"
        style={{
          border: `1px solid ${type === "life" ? lifeColor : knowledgeColor}22`,
          background: "rgba(0,0,0,0.5)",
          boxShadow: `0 0 40px ${type === "life" ? "rgba(0,200,160,0.06)" : "rgba(200,140,255,0.06)"}`,
        }}
      >
        <canvas ref={canvasRef} className="block" />
        {/* Root ground glow */}
        <div
          className="absolute bottom-0 left-1/2 -translate-x-1/2 w-48 h-1 blur-sm"
          style={{
            background: `linear-gradient(90deg, transparent, ${type === "life" ? lifeColor : knowledgeColor}60, transparent)`,
          }}
        />
      </div>
    </div>
  );
}

export function TreeSection() {
  return (
    <section className="py-24 px-6 relative overflow-hidden">
      <div className="absolute inset-0 pointer-events-none">
        <div className="absolute top-0 w-full h-px spectrum-divider opacity-20" />
      </div>

      <div className="max-w-6xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 1 }}
          className="text-center mb-16"
        >
          <p className="font-mono text-[10px] uppercase tracking-[0.3em] text-white/30 mb-4">
            Fibonacci Spiral Flow · Governing the Manifold
          </p>
          <h2
            className="font-cinzel text-4xl md:text-5xl font-bold mb-4 text-glow-gold"
            style={{ color: "#e8c040" }}
          >
            Two Trees, One Root
          </h2>
          <p className="font-mono text-sm text-white/40 max-w-2xl mx-auto">
            Expansion and convergence are two directions of one circulation. The
            Tree of Life is the outward flow — throughput, events, living
            services. The Tree of Knowledge is the inward flow — depth, memory,
            understanding.
          </p>
        </motion.div>

        <div className="flex flex-col md:flex-row justify-center gap-8 md:gap-16 items-end">
          <motion.div
            initial={{ opacity: 0, x: -40 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 1 }}
          >
            <TreeVisualizer type="life" />
          </motion.div>

          {/* Center connecting element */}
          <div className="hidden md:flex flex-col items-center justify-center pb-12 gap-3">
            <div className="w-px h-24 bg-gradient-to-b from-transparent via-white/20 to-transparent" />
            <div
              className="w-8 h-8 rounded-full border flex items-center justify-center text-white/40 font-mono text-[8px] uppercase"
              style={{
                borderColor: "rgba(255,255,255,0.1)",
                background: "rgba(255,255,255,0.02)",
              }}
            >
              ⊕
            </div>
            <div className="w-px h-24 bg-gradient-to-b from-transparent via-white/20 to-transparent" />
          </div>

          <motion.div
            initial={{ opacity: 0, x: 40 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 1, delay: 0.2 }}
          >
            <TreeVisualizer type="knowledge" />
          </motion.div>
        </div>

        {/* Core principles row */}
        <div className="mt-16 grid grid-cols-1 md:grid-cols-3 gap-4 max-w-3xl mx-auto">
          {[
            {
              label: "Logarithmic Growth",
              desc: "Fibonacci spiral (φ = 1.618) governs branch ratios in both trees.",
            },
            {
              label: "Holographic Encoding",
              desc: "Every node contains information about the whole federation.",
            },
            {
              label: "Phase Coherence",
              desc: "When both trees synchronize, the field enters coherence interference.",
            },
          ].map((p) => (
            <div
              key={p.label}
              className="p-4 rounded-xl text-center"
              style={{
                border: "1px solid rgba(255,255,255,0.06)",
                background: "rgba(255,255,255,0.01)",
              }}
            >
              <p className="font-mono text-[10px] uppercase tracking-widest text-white/35 mb-2">
                {p.label}
              </p>
              <p className="font-mono text-xs text-white/25 leading-relaxed">
                {p.desc}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
