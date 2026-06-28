import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from '@/components/ui/toaster';
import { TooltipProvider } from '@/components/ui/tooltip';
import NotFound from '@/pages/not-found';
import { Route, Switch, Router as WouterRouter } from 'wouter';

import { Hero } from '@/components/Hero';
import { ArchitectureRings } from '@/components/ArchitectureRings';
import { LivingLayers } from '@/components/LivingLayers';
import { TreeSection } from '@/components/Trees';
import { LivingSpectrum } from '@/components/LivingSpectrum';
import { DiscoveryLifecycle } from '@/components/DiscoveryLifecycle';
import { Footer } from '@/components/Footer';

const queryClient = new QueryClient();

function Home() {
  return (
    <main className="min-h-screen w-full selection:bg-primary/30 selection:text-primary">
      <Hero />
      <ArchitectureRings />
      <LivingSpectrum />
      <LivingLayers />
      <TreeSection />
      <DiscoveryLifecycle />
      <Footer />
    </main>
  );
}

function Router() {
  return (
    <Switch>
      <Route path="/" component={Home} />
      <Route component={NotFound} />
    </Switch>
  );
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <TooltipProvider>
        <WouterRouter base={import.meta.env.BASE_URL.replace(/\/$/, '')}>
          <Router />
        </WouterRouter>
        <Toaster />
      </TooltipProvider>
    </QueryClientProvider>
  );
}

export default App;
