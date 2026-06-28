import { motion } from "framer-motion";

export function Hero() {
  return (
    <section className="relative min-h-screen flex flex-col items-center justify-center overflow-hidden px-6 pt-20">
      <div className="absolute inset-0 flex items-center justify-center opacity-20 pointer-events-none">
        <motion.div
          animate={{
            scale: [1, 1.1, 1],
            opacity: [0.3, 0.5, 0.3],
          }}
          transition={{ duration: 10, repeat: Infinity, ease: "easeInOut" }}
          className="w-[800px] h-[800px] rounded-full bg-primary/20 blur-[120px]"
        />
        <motion.div
          animate={{
            scale: [1, 1.2, 1],
            opacity: [0.2, 0.4, 0.2],
          }}
          transition={{ duration: 15, repeat: Infinity, ease: "easeInOut", delay: 2 }}
          className="absolute w-[600px] h-[600px] rounded-full bg-secondary/20 blur-[100px]"
        />
      </div>

      <div className="relative z-10 max-w-5xl mx-auto text-center flex flex-col items-center">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 1, ease: "easeOut" }}
          className="mb-6 flex items-center gap-3"
        >
          <div className="h-[1px] w-12 bg-primary/50" />
          <span className="font-mono text-primary uppercase tracking-widest text-sm">
            AI Freedom Trust Federation
          </span>
          <div className="h-[1px] w-12 bg-primary/50" />
        </motion.div>

        <motion.h1
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 1.2, ease: "easeOut", delay: 0.2 }}
          className="text-6xl md:text-8xl font-bold tracking-tighter mb-8 text-glow-amber text-foreground"
        >
          AIFT<span className="text-primary">-</span>OS
        </motion.h1>

        <motion.p
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 1, ease: "easeOut", delay: 0.5 }}
          className="text-xl md:text-2xl text-muted-foreground font-mono max-w-3xl leading-relaxed mb-16"
        >
          The living, federated operating system governed by evidence, not authority.
        </motion.p>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 1.5, ease: "easeOut", delay: 0.8 }}
          className="relative group p-1"
        >
          <div className="absolute inset-0 bg-gradient-to-b from-primary/20 to-transparent rounded-lg blur-md opacity-50 group-hover:opacity-100 transition-opacity duration-700" />
          <div className="relative border border-primary/20 bg-card/50 backdrop-blur-sm p-8 md:p-12 rounded-lg text-left max-w-4xl mx-auto border-t-primary/50">
            <h2 className="font-mono text-sm text-primary mb-4 uppercase tracking-widest flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-primary animate-pulse" />
              Core Principle
            </h2>
            <p className="text-xl md:text-3xl font-medium leading-snug text-foreground/90 font-sans">
              "Everything visible within the operating system must first be discovered. Repositories. Capabilities. Services. Modules. Commands. Documentation. Workflows. Registries. Events. Tests."
            </p>
            <p className="mt-6 text-lg md:text-xl text-muted-foreground font-mono">
              Nothing appears because someone declared it. Everything appears because the operating system found evidence.
            </p>
          </div>
        </motion.div>

        <motion.div
          animate={{ y: [0, 10, 0] }}
          transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }}
          className="mt-24 text-muted-foreground/50 flex flex-col items-center gap-2"
        >
          <span className="font-mono text-xs uppercase tracking-widest">Inspect Reality</span>
          <div className="w-[1px] h-16 bg-gradient-to-b from-muted-foreground/50 to-transparent" />
        </motion.div>
      </div>
    </section>
  );
}
