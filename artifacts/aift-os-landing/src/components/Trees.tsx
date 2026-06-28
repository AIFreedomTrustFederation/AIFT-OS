import { motion } from "framer-motion";
import { useEffect, useRef } from "react";

// Procedural Tree Generator using Canvas
export function TreeVisualizer({ 
  type = "life", 
  width = 400, 
  height = 600 
}: { 
  type: "life" | "knowledge",
  width?: number,
  height?: number
}) {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    // High DPI display support
    const dpr = window.devicePixelRatio || 1;
    canvas.width = width * dpr;
    canvas.height = height * dpr;
    ctx.scale(dpr, dpr);
    canvas.style.width = `${width}px`;
    canvas.style.height = `${height}px`;

    let animationFrameId: number;
    let time = 0;

    // Tree settings based on type
    const isLife = type === "life";
    const primaryColor = isLife ? "rgba(20, 184, 166, " : "rgba(245, 166, 35, "; // Teal for life, Amber for knowledge
    const branchWidthMult = isLife ? 0.8 : 0.6; // Life has thicker branches (throughput), Knowledge has thinner deeper branches
    const maxDepth = isLife ? 6 : 8; // Knowledge goes deeper

    const drawBranch = (
      startX: number, 
      startY: number, 
      len: number, 
      angle: number, 
      depth: number,
      t: number
    ) => {
      if (depth === 0) return;

      // Calculate end point
      const endX = startX + len * Math.cos(angle);
      const endY = startY + len * Math.sin(angle);

      // Draw branch
      ctx.beginPath();
      ctx.moveTo(startX, startY);
      ctx.lineTo(endX, endY);
      
      const lineWidth = depth * branchWidthMult;
      ctx.lineWidth = lineWidth;
      
      // Dynamic opacity/pulse based on type
      let alpha = 0.4 + (depth / maxDepth) * 0.4;
      if (isLife) {
        // Life pulses along the branch
        alpha += Math.sin(t * 0.05 + depth) * 0.2;
      } else {
        // Knowledge settles, slowly gets brighter
        alpha = Math.min(0.8, 0.2 + t * 0.001);
      }
      
      ctx.strokeStyle = `${primaryColor}${alpha})`;
      ctx.stroke();

      // Draw nodes/leaves at ends
      if (depth <= 2) {
        ctx.beginPath();
        ctx.arc(endX, endY, isLife ? 2 : 1.5, 0, Math.PI * 2);
        ctx.fillStyle = `${primaryColor}${isLife ? 0.8 : 0.6})`;
        ctx.fill();
        
        // Occasional glowing nodes
        if (Math.random() > 0.95) {
           ctx.shadowColor = `${primaryColor}1)`;
           ctx.shadowBlur = 10;
           ctx.fill();
           ctx.shadowBlur = 0; // reset
        }
      }

      // Next branches — synchronous recursion, no setTimeout
      const sway = Math.sin(t * 0.01 + depth) * 0.05;
      drawBranch(endX, endY, len * 0.7, angle - 0.5 + sway, depth - 1, t);
      drawBranch(endX, endY, len * 0.7, angle + 0.5 + sway, depth - 1, t);
      // Deterministic middle branch for knowledge (use depth parity, no Math.random in render loop)
      if (!isLife && depth % 2 === 0) {
        drawBranch(endX, endY, len * 0.6, angle + sway * 2, depth - 1, t);
      }
    };

    const render = () => {
      ctx.clearRect(0, 0, width, height);
      
      // Start from bottom middle
      drawBranch(width / 2, height - 20, height * 0.25, -Math.PI / 2, maxDepth, time);
      
      time++;
      
      // Only animate continuously if it's the life tree, otherwise static after growth
      if (isLife) {
        animationFrameId = requestAnimationFrame(render);
      }
    };

    // Initial render
    render();

    return () => {
      if (animationFrameId) cancelAnimationFrame(animationFrameId);
    };
  }, [type, width, height]);

  return (
    <div className="relative flex flex-col items-center">
      <h3 className="font-mono text-xl mb-4 font-bold text-center h-16">
        {type === "life" ? (
          <span className="text-teal-400">Tree of Life<br/><span className="text-xs text-muted-foreground font-normal">"What is alive and running?"</span></span>
        ) : (
          <span className="text-amber-400">Tree of Knowledge<br/><span className="text-xs text-muted-foreground font-normal">"What has been learned?"</span></span>
        )}
      </h3>
      <div className="relative border border-border/50 rounded-xl bg-card/10 overflow-hidden backdrop-blur-sm">
        <canvas ref={canvasRef} className="block" />
        {/* Fake root connecting them conceptually if placed side by side */}
        <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-32 h-1 bg-gradient-to-r from-transparent via-primary/50 to-transparent blur-[2px]" />
      </div>
    </div>
  );
}

export function TreeSection() {
  return (
    <section className="py-32 px-6 max-w-7xl mx-auto">
      <div className="flex flex-col md:flex-row justify-center gap-12 md:gap-24 items-end">
        <motion.div
          initial={{ opacity: 0, x: -40 }}
          whileInView={{ opacity: 1, x: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 1 }}
        >
          <TreeVisualizer type="life" />
        </motion.div>
        
        <motion.div
          initial={{ opacity: 0, x: 40 }}
          whileInView={{ opacity: 1, x: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 1, delay: 0.2 }}
        >
          <TreeVisualizer type="knowledge" />
        </motion.div>
      </div>
      <div className="mt-16 text-center max-w-3xl mx-auto">
        <p className="text-muted-foreground font-mono text-sm leading-relaxed">
          Both trees grow from the same root. The Tree of Life shows throughput—living repositories, active services, events. The Tree of Knowledge shows depth—books, documentation, architecture, history.
        </p>
      </div>
    </section>
  );
}
