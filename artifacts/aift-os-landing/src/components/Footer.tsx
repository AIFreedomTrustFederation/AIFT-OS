import { motion } from "framer-motion";
import { Github } from "lucide-react";

export function Footer() {
  return (
    <footer className="relative py-32 px-6 overflow-hidden border-t border-border bg-card/20">
      <div className="absolute inset-0 flex items-center justify-center opacity-10 pointer-events-none">
        <div className="w-[800px] h-[800px] rounded-full bg-primary blur-[150px]" />
      </div>

      <div className="max-w-4xl mx-auto relative z-10 text-center">
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 1 }}
        >
          <h2 className="text-3xl md:text-5xl font-bold mb-8 font-sans text-glow-amber">Sovereign Growth</h2>
          
          <div className="prose prose-invert prose-lg mx-auto font-serif text-muted-foreground leading-relaxed mb-16">
            <p>
              The long-term objective is not to create another software platform. The objective is to cultivate a truthful digital ecosystem where every sovereign installation contributes to a larger federation without surrendering ownership of its own knowledge.
            </p>
            <p>
              Repositories become branches. Capabilities become fruit. Documentation becomes memory. Events become circulation. Artificial intelligence becomes assistance rather than authority.
            </p>
          </div>

          <a 
            href="https://github.com/AIFreedomTrustFederation"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-3 px-8 py-4 rounded-full bg-primary text-primary-foreground font-mono font-bold hover:bg-primary/90 transition-all hover:scale-105 active:scale-95 shadow-[0_0_30px_rgba(245,166,35,0.3)] hover:shadow-[0_0_40px_rgba(245,166,35,0.5)]"
          >
            <Github className="w-5 h-5" />
            Inspect the Source
          </a>
        </motion.div>
        
        <div className="mt-32 text-xs text-muted-foreground/40 font-mono flex flex-col items-center gap-2">
          <div className="w-px h-12 bg-gradient-to-b from-transparent to-muted-foreground/20" />
          <span>AIFT-OS • TRUTH BEFORE AUTHORITY</span>
        </div>
      </div>
    </footer>
  );
}
