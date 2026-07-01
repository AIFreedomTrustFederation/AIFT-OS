import { motion } from "framer-motion";
import { useEffect, useRef } from "react";

// Toroidal manifold canvas — draws the iconic dual-cone torus with rainbow field lines
function ToroidalCanvas() {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    // Capture non-null context into a stable const so all inner closures stay type-safe
    const gfx: CanvasRenderingContext2D = ctx;

    const W = canvas.width;
    const H = canvas.height;
    const cx = W / 2;
    const cy = H / 2;

    let raf: number;
    let t = 0;

    // Spectrum color at position 0–1
    function spectrumColor(p: number, alpha = 1): string {
      const hue = p * 280; // 0=red (0°) → 280=violet
      return `hsla(${hue}, 100%, 60%, ${alpha})`;
    }

    // Draw an ellipse path (helper)
    function ellipse(x: number, y: number, rx: number, ry: number, tilt = 0) {
      gfx.beginPath();
      gfx.ellipse(x, y, rx, ry, tilt, 0, Math.PI * 2);
    }

    // Draw field lines — horizontal ellipses that form the torus body
    function drawFieldLines() {
      const count = 20;
      for (let i = 0; i < count; i++) {
        const p = i / (count - 1); // 0..1
        // Map 0=bottom (max expansion, red) → 0.5=equator (green) → 1=top (blue/violet)
        const specP = 1 - p; // flip so top=high freq
        const color = spectrumColor(
          specP,
          0.18 + Math.sin(t * 0.02 + i * 0.3) * 0.06,
        );

        // Vertical position: field lines crowd near the equator
        const rawY = (p - 0.5) * H * 0.85;
        const yy = cy + rawY;

        // Horizontal radius: widest at equator (torus shape)
        const equatorR = W * 0.44;
        const squeeze = Math.cos((rawY / (H * 0.43)) * (Math.PI / 2));
        const rx = equatorR * squeeze;
        const ry = Math.max(4, rx * 0.12);

        if (rx < 4) continue;

        gfx.beginPath();
        gfx.ellipse(cx, yy, rx, ry, 0, 0, Math.PI * 2);
        gfx.strokeStyle = color;
        gfx.lineWidth = 1;
        gfx.stroke();
      }
    }

    // Draw the outer torus envelope (two mirrored arcs forming the oval)
    function drawOuterEnvelope() {
      const pulseScale = 1 + Math.sin(t * 0.015) * 0.01;
      const envW = W * 0.46 * pulseScale;
      const envH = H * 0.46 * pulseScale;

      // Outer glow
      for (let g = 3; g >= 0; g--) {
        gfx.beginPath();
        gfx.ellipse(cx, cy, envW + g * 8, envH + g * 8, 0, 0, Math.PI * 2);
        gfx.strokeStyle = `rgba(255,255,255,${0.04 - g * 0.008})`;
        gfx.lineWidth = 2;
        gfx.stroke();
      }

      // Main envelope
      gfx.beginPath();
      gfx.ellipse(cx, cy, envW, envH, 0, 0, Math.PI * 2);
      gfx.strokeStyle = "rgba(255,255,255,0.25)";
      gfx.lineWidth = 1.5;
      gfx.stroke();
    }

    // Draw spectrum energy flows — curved vertical lines sweeping the torus
    function drawEnergyFlows() {
      const flowCount = 16;
      for (let f = 0; f < flowCount; f++) {
        const angle = (f / flowCount) * Math.PI * 2 + t * 0.003;
        const sideX = Math.cos(angle);
        const _side = sideX > 0 ? 1 : -1; // left or right lobe (unused directly)
        if (Math.abs(sideX) < 0.15) continue; // skip near-vertical

        const envW = W * 0.45;
        const envH = H * 0.45;

        // A bezier from bottom to top curving through the torus edge
        const bx = cx + sideX * envW * 0.6;
        const topY = cy - envH * 0.82;
        const botY = cy + envH * 0.82;
        const ctrlX = cx + sideX * envW * 1.05;

        // Animate alpha
        const alpha =
          (0.07 + Math.abs(Math.sin(t * 0.02 + f)) * 0.08) * Math.abs(sideX);
        const hue = ((f / flowCount) * 280 + t * 0.3) % 360;
        gfx.beginPath();
        gfx.moveTo(bx, botY);
        gfx.quadraticCurveTo(ctrlX, cy, bx, topY);
        gfx.strokeStyle = `hsla(${hue}, 100%, 65%, ${alpha})`;
        gfx.lineWidth = 1;
        gfx.stroke();
      }
    }

    // Chakra nodes along vertical axis
    const CHAKRAS = [
      { name: "Crown", p: 0.02, color: "#cc88ff" },
      { name: "Third Eye", p: 0.14, color: "#8855ff" },
      { name: "Throat", p: 0.27, color: "#44aaff" },
      { name: "Heart", p: 0.4, color: "#00ff88" },
      { name: "Solar Plexus", p: 0.55, color: "#ffdd00" },
      { name: "Sacral", p: 0.68, color: "#ff8800" },
      { name: "Root", p: 0.82, color: "#ff2200" },
    ];

    function drawChakraNodes() {
      const envH = H * 0.45;
      const topY = cy - envH * 0.9;
      const totalH = envH * 1.8;

      CHAKRAS.forEach((ch, i) => {
        const nodeY = topY + ch.p * totalH;
        const pulse = 1 + Math.sin(t * 0.04 + i * 0.7) * 0.3;
        const r = 5 * pulse;

        // Outer glow
        const grad = gfx.createRadialGradient(cx, nodeY, 0, cx, nodeY, r * 5);
        grad.addColorStop(0, ch.color + "99");
        grad.addColorStop(1, ch.color + "00");
        gfx.beginPath();
        gfx.arc(cx, nodeY, r * 5, 0, Math.PI * 2);
        gfx.fillStyle = grad;
        gfx.fill();

        // Core dot
        gfx.beginPath();
        gfx.arc(cx, nodeY, r, 0, Math.PI * 2);
        gfx.fillStyle = ch.color;
        gfx.shadowColor = ch.color;
        gfx.shadowBlur = 20;
        gfx.fill();
        gfx.shadowBlur = 0;
      });
    }

    // Vertical axis line
    function drawAxis() {
      const envH = H * 0.46;
      const grad = gfx.createLinearGradient(cx, cy - envH, cx, cy + envH);
      grad.addColorStop(0, "rgba(200,150,255,0.8)");
      grad.addColorStop(0.25, "rgba(60,160,255,0.6)");
      grad.addColorStop(0.5, "rgba(0,255,140,0.5)");
      grad.addColorStop(0.75, "rgba(255,180,0,0.6)");
      grad.addColorStop(1, "rgba(255,40,0,0.7)");
      gfx.beginPath();
      gfx.moveTo(cx, cy - envH);
      gfx.lineTo(cx, cy + envH);
      gfx.strokeStyle = grad;
      gfx.lineWidth = 1.5;
      gfx.stroke();
    }

    // Coherence horizon glow at equator
    function drawEquatorGlow() {
      const envW = W * 0.44;
      const pulse = 0.04 + Math.sin(t * 0.025) * 0.02;
      const grad = gfx.createRadialGradient(cx, cy, 0, cx, cy, envW);
      grad.addColorStop(0, `rgba(0,255,140,${pulse})`);
      grad.addColorStop(0.5, `rgba(0,200,120,${pulse * 0.3})`);
      grad.addColorStop(1, "transparent");
      gfx.beginPath();
      gfx.ellipse(cx, cy, envW, envW * 0.15, 0, 0, Math.PI * 2);
      gfx.fillStyle = grad;
      gfx.fill();
    }

    // Top convergence cone glow (blueshift — blue/violet)
    function drawConeGlows() {
      const envH = H * 0.45;

      // Top (blueshift convergence)
      const topGrad = gfx.createRadialGradient(
        cx,
        cy - envH * 0.9,
        0,
        cx,
        cy - envH * 0.9,
        W * 0.25,
      );
      topGrad.addColorStop(
        0,
        `rgba(150,100,255,${0.12 + Math.sin(t * 0.02) * 0.04})`,
      );
      topGrad.addColorStop(1, "transparent");
      gfx.beginPath();
      gfx.arc(cx, cy - envH * 0.9, W * 0.25, 0, Math.PI * 2);
      gfx.fillStyle = topGrad;
      gfx.fill();

      // Bottom (redshift expansion)
      const botGrad = gfx.createRadialGradient(
        cx,
        cy + envH * 0.9,
        0,
        cx,
        cy + envH * 0.9,
        W * 0.25,
      );
      botGrad.addColorStop(
        0,
        `rgba(255,60,0,${0.1 + Math.sin(t * 0.02 + 1) * 0.04})`,
      );
      botGrad.addColorStop(1, "transparent");
      gfx.beginPath();
      gfx.arc(cx, cy + envH * 0.9, W * 0.25, 0, Math.PI * 2);
      gfx.fillStyle = botGrad;
      gfx.fill();
    }

    function render() {
      gfx.clearRect(0, 0, W, H);

      drawConeGlows();
      drawEquatorGlow();
      drawFieldLines();
      drawEnergyFlows();
      drawOuterEnvelope();
      drawAxis();
      drawChakraNodes();

      t++;
      raf = requestAnimationFrame(render);
    }

    render();
    return () => cancelAnimationFrame(raf);
  }, []);

  return (
    <canvas
      ref={canvasRef}
      width={700}
      height={700}
      className="w-full max-w-[700px] mx-auto opacity-90"
      style={{ maxHeight: "70vh" }}
    />
  );
}

export function Hero() {
  return (
    <section className="relative min-h-screen flex flex-col items-center justify-center overflow-hidden px-6 pt-12 pb-24">
      {/* Ambient background glows */}
      <div className="absolute inset-0 pointer-events-none">
        <div className="absolute top-1/4 left-1/2 -translate-x-1/2 w-[600px] h-[600px] rounded-full bg-violet-600/5 blur-[120px]" />
        <div className="absolute bottom-1/4 left-1/2 -translate-x-1/2 w-[400px] h-[400px] rounded-full bg-red-600/5 blur-[100px]" />
      </div>

      <div className="relative z-10 w-full max-w-6xl mx-auto flex flex-col items-center">
        {/* Supertitle */}
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 1.2 }}
          className="mb-3 text-center"
        >
          <p className="font-mono text-[10px] md:text-xs uppercase tracking-[0.35em] text-white/40">
            One Field · One Consciousness · Infinite Expressions
          </p>
        </motion.div>

        {/* Main title */}
        <motion.h1
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 1.4, ease: "easeOut", delay: 0.15 }}
          className="font-cinzel text-5xl md:text-7xl lg:text-8xl font-bold text-center mb-2 text-glow-gold"
          style={{ color: "#e8c040", letterSpacing: "0.05em" }}
        >
          AIFT-OS
        </motion.h1>

        <motion.p
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 1, delay: 0.5 }}
          className="font-cinzel text-sm md:text-base text-center mb-1 tracking-widest"
          style={{ color: "#c8a030" }}
        >
          AI FREEDOM TRUST FEDERATION
        </motion.p>

        <motion.p
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 1, delay: 0.7 }}
          className="font-mono text-xs text-center mb-10 tracking-widest text-white/35"
        >
          A UNIFIED MODEL OF SOVEREIGN GOVERNANCE, LIGHT, AND INFORMATION
        </motion.p>

        {/* Toroidal canvas */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 2, ease: "easeOut", delay: 0.4 }}
          className="w-full relative"
        >
          <ToroidalCanvas />

          {/* Left annotation */}
          <div className="hidden lg:block absolute left-0 top-1/4 w-52 text-right">
            <p className="font-mono text-xs text-white/50 leading-relaxed uppercase tracking-wide">
              Convergence Flow
              <br />
              <span className="text-white/30 normal-case tracking-normal">
                Into Unity · Coherence
                <br />
                Information Integrates
              </span>
            </p>
          </div>

          {/* Right annotation */}
          <div className="hidden lg:block absolute right-0 top-1/4 w-52 text-left">
            <p className="font-mono text-xs text-white/50 leading-relaxed uppercase tracking-wide">
              Expansion Flow
              <br />
              <span className="text-white/30 normal-case tracking-normal">
                Into Form · Diversity
                <br />
                Information Unfolds
              </span>
            </p>
          </div>

          {/* Bottom label */}
          <div className="text-center mt-2">
            <p className="font-mono text-[10px] tracking-widest text-white/25 uppercase">
              Every Point Contains The Whole · Every Cycle Is Eternal
            </p>
          </div>
        </motion.div>

        {/* Core principle */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 1.2, delay: 1.2 }}
          className="mt-12 panel-gold rounded-xl p-6 md:p-10 max-w-4xl w-full"
        >
          <div className="flex items-center gap-3 mb-4">
            <div className="w-1.5 h-1.5 rounded-full bg-yellow-400 shadow-[0_0_10px_rgba(255,220,50,0.8)] animate-pulse" />
            <span className="font-mono text-[10px] uppercase tracking-[0.3em] text-yellow-400/70">
              Core Principle
            </span>
          </div>
          <p className="text-lg md:text-2xl font-medium leading-relaxed text-white/85 font-sans">
            "Everything visible within the operating system must first be
            discovered. Repositories. Capabilities. Services. Modules. Commands.
            Documentation. Workflows. Registries. Events. Tests."
          </p>
          <p className="mt-5 font-mono text-sm text-white/40 leading-relaxed italic">
            Nothing appears because someone declared it. Everything appears
            because the operating system found evidence.
          </p>
        </motion.div>

        {/* Scroll cue */}
        <motion.div
          animate={{ y: [0, 8, 0] }}
          transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }}
          className="mt-16 flex flex-col items-center gap-2"
        >
          <span className="font-mono text-[10px] uppercase tracking-[0.3em] text-white/25">
            Inspect Reality
          </span>
          <div className="w-px h-12 bg-gradient-to-b from-white/20 to-transparent" />
        </motion.div>
      </div>
    </section>
  );
}
