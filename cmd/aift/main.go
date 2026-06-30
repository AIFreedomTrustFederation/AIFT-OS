package main

import (
	"fmt"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/builder"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/capability"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/compiler"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/fedbuild"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/lifecycle"
	"os"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/ai"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/api"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/capabilities"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/daemon"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/doctor"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/eventmesh"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/federation"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/gitx"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/graph"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/intelligence"
	_ "github.com/AIFreedomTrustFederation/AIFT-OS/internal/kernel"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/manifests"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/manual"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/planner"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/plugins"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/providers"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/registry"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/repo"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/reports"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/runtime"
	_ "github.com/AIFreedomTrustFederation/AIFT-OS/internal/scheduler"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/servicecontracts"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/services"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/sync"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/version"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workflow"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

func main() {
	cfg := config.Load()
	cmd := "help"
	args := []string{}

	if len(os.Args) > 1 {
		cmd = os.Args[1]
		args = os.Args[2:]
	}

	if looksLikeExecutablePath(cmd) {
		if len(args) > 0 {
			cmd = args[0]
			args = args[1:]
		} else {
			cmd = "help"
		}
	}

	var err error

	switch cmd {
	case "help", "-h", "--help":
		help()
	case "version":
		fmt.Printf("%s %s — %s\n", version.Name, version.Version, version.Role)
	case "doctor":
		err = doctor.Run(cfg)
	case "status":
		err = status(cfg)
	case "manifest":
		err = manifests.EnsureAll(cfg)
		if err == nil {
			fmt.Println("OK: manifests ensured")
		}
	case "registry":
		err = registry.Generate(cfg)
	case "dashboard":
		err = reports.Dashboard(cfg)
	case "deps":
		err = reports.Deps(cfg)
	case "plugins":
		err = plugins.List(cfg)
	case "providers":
		err = providers.List(cfg)
	case "events":
		err = events.Tail(cfg, 25)
	case "services":
		err = services.List(cfg)
	case "start":
		err = runtime.StartOnce(cfg)
	case "tick":
		err = runtime.Tick(cfg)
	case "serve":
		addr := ":8787"
		if len(args) > 0 {
			addr = args[0]
		}
		err = api.New(cfg, addr).Serve()
	case "daemon":
		addr := ":8787"
		if len(args) > 0 {
			addr = args[0]
		}
		err = daemon.Start(cfg, addr)
	case "sync":
		if len(args) == 0 || args[0] == "--safe" || args[0] == "safe" {
			err = sync.Safe(cfg)
		} else {
			err = fmt.Errorf("only sync --safe is implemented in Go kernel")
		}
	case "federation":
		err = runFederation(cfg, args)
	case "repo":
		err = runRepo(cfg, args)
	case "workflow":
		err = runWorkflow(cfg, args)
	case "capabilities":
		err = runCapabilities(cfg, args)
	case "intelligence":
		err = runIntelligence(cfg, args)
	case "manual":
		err = runManual(cfg, args)
	case "graph":
		err = graph.Query(cfg, args)
	case "mesh":
		err = runMesh(cfg, args)
	case "service-contracts":
		err = runServiceContracts(cfg, args)
	case "plan":
		err = runPlanner(cfg, args)
	case "modules":
		err = runModules(cfg, args)
	case "kernel-registry":
		err = runKernelRegistry(cfg, args)
	case "discovery":
		err = runDiscovery(cfg, args)
	case "event-bus":
		err = runEventBus(cfg, args)
	case "patch-engine":
		err = runPatchEngine(cfg, args)
	case "kernel":
		err = runKernelRuntime(cfg, args)
	case "runtime":
		err = runRuntime(cfg, args)
	case "operator":
		err = runOperator(cfg, args)
	case "scheduler":
		err = runScheduler(cfg, args)

	case "ai":
		if err := ai.Run(cfg); err != nil {
			panic(err)
		}
	case "compile":
		if err := compiler.Run(cfg); err != nil {
			panic(err)
		}
	case "compiler":
		if err := compiler.Run(cfg); err != nil {
			panic(err)
		}
	case "capability":
		if err := capability.Run(cfg); err != nil {
			panic(err)
		}
	case "federation-build":
		async := len(args) > 1 && args[1] == "--async"
		if err := fedbuild.Run(cfg, async); err != nil {
			panic(err)
		}
	case "build":
		if err := builder.Run(cfg); err != nil {
			panic(err)
		}
	case "lifecycle":
		if err := lifecycle.Run(cfg); err != nil {
			panic(err)
		}
	case "verify":
		err = verify(cfg)
	default:
		err = fmt.Errorf("unknown command: %s", cmd)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}

func looksLikeExecutablePath(s string) bool {
	return len(s) > 0 && (s[0] == '/' || s == "aiftd" || s == "./aiftd" || s == "bin/aiftd")
}

func help() {
	fmt.Println("AIFT-OS Federation Control Plane")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  help")
	fmt.Println("  version")
	fmt.Println("  doctor")
	fmt.Println("  status")
	fmt.Println("  manifest")
	fmt.Println("  registry")
	fmt.Println("  dashboard")
	fmt.Println("  deps")
	fmt.Println("  plugins")
	fmt.Println("  providers")
	fmt.Println("  events")
	fmt.Println("  services")
	fmt.Println("  start")
	fmt.Println("  tick")
	fmt.Println("  serve [:8787]")
	fmt.Println("  daemon [:8787]")
	fmt.Println("  sync --safe")
	fmt.Println("  federation scan|graph|verify")
	fmt.Println("  repo list|inspect|run")
	fmt.Println("  workflow list")
	fmt.Println("  intelligence scan|report|repo|roadmap")
	fmt.Println("  manual init-all|scan|report|repo")
	fmt.Println("  graph [summary|repo|type|status]")
	fmt.Println("  mesh init-all|scan|topics|subscribers|publish|replay|tail|report")
	fmt.Println("  service-contracts init-all|scan|list|repo|report")
	fmt.Println("  plan build|summary|repo|ready|blocked|report")
	fmt.Println("  modules init-all|scan|list|repo|report")
	fmt.Println("  kernel-registry scan|list|object|report")
	fmt.Println("  discovery scan|list|object|report")
	fmt.Println("  event-bus publish|list|replay|report")
	fmt.Println("  patch-engine inspect|plan|validate")
	fmt.Println("  kernel boot|status|report")
	fmt.Println("  runtime scan|status|ready|blocked|report")
	fmt.Println("  capabilities scan|list|info|report")
	fmt.Println("  operator check")
	fmt.Println("  scheduler plan|ready|blocked|report")
	fmt.Println("  ai")
	fmt.Println("  compile")
	fmt.Println("  compiler")
	fmt.Println("  build")
	fmt.Println("  capability")
	fmt.Println("  federation-build")
	fmt.Println("  lifecycle")
	fmt.Println("  verify")
}

func runFederation(cfg config.Config, args []string) error {
	if len(args) == 0 || args[0] == "scan" {
		return federation.Scan(cfg)
	}
	if args[0] == "graph" {
		return federation.Graph(cfg)
	}
	if args[0] == "verify" {
		return federation.Verify(cfg)
	}
	return fmt.Errorf("usage: aift federation scan|graph|verify")
}

func runRepo(cfg config.Config, args []string) error {
	if len(args) == 0 || args[0] == "list" {
		return repo.PrintList(cfg)
	}
	if args[0] == "inspect" {
		if len(args) < 2 {
			return fmt.Errorf("usage: aift repo inspect <name>")
		}
		return repo.PrintInspect(cfg, repo.NormalizeName(args[1]))
	}
	if args[0] == "run" {
		if len(args) < 3 {
			return fmt.Errorf("usage: aift repo run <name> <command> [args...]")
		}
		return repo.RunCommand(cfg, repo.NormalizeName(args[1]), args[2], args[3:])
	}
	return fmt.Errorf("usage: aift repo list|inspect|run")
}

func runWorkflow(cfg config.Config, args []string) error {
	if len(args) == 0 || args[0] == "list" {
		return workflow.List(cfg)
	}
	return fmt.Errorf("usage: aift workflow list")
}

func verify(cfg config.Config) error {
	if err := doctor.Run(cfg); err != nil {
		return err
	}
	if err := manifests.EnsureAll(cfg); err != nil {
		return err
	}
	if err := providers.WriteRegistry(cfg); err != nil {
		return err
	}
	if err := registry.Generate(cfg); err != nil {
		return err
	}
	if err := reports.Dashboard(cfg); err != nil {
		return err
	}
	if err := reports.Deps(cfg); err != nil {
		return err
	}
	if err := capabilities.Scan(cfg); err != nil {
		return err
	}
	if err := intelligence.Scan(cfg); err != nil {
		return err
	}
	if err := manual.Scan(cfg); err != nil {
		return err
	}
	if err := graph.Build(cfg); err != nil {
		return err
	}
	if err := eventmesh.Scan(cfg); err != nil {
		return err
	}
	if err := servicecontracts.Scan(cfg); err != nil {
		return err
	}
	if err := planner.Build(cfg); err != nil {
		return err
	}
	if err := events.Emit(cfg, "verify.complete", "verify", "federation verified", nil); err != nil {
		return err
	}
	fmt.Println("OK: federation verified")
	return nil
}

func status(cfg config.Config) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("%-32s %-12s %-8s %s\n", "REPOSITORY", "BRANCH", "STATE", "REMOTE")
	for _, repo := range repos {
		state := "clean"
		if gitx.Dirty(repo.Path) {
			state = "dirty"
		}
		fmt.Printf("%-32s %-12s %-8s %s\n", repo.Name, gitx.Branch(repo.Path), state, gitx.Remote(repo.Path))
	}

	return nil
}
