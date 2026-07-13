#!/usr/bin/env bash
set -euo pipefail

mkdir -p tools/cli-self-verifier reports

cat > tools/cli-self-verifier/main.go <<'GO'
package main

import (
"fmt"
"go/ast"
"go/parser"
"go/token"
"os"
"regexp"
"sort"
"strings"
)

type Report struct {
Commands        []string
MissingHandlers []string
Handlers        []string
UnusedHandlers  []string
HelpCommands    []string
MissingHelp      []string
ExtraHelp        []string
}

func main() {
path := "cmd/aift/main.go"
srcBytes, err := os.ReadFile(path)
if err != nil {
ic(err)
}
src := string(srcBytes)

fset := token.NewFileSet()
file, err := parser.ParseFile(fset, path, srcBytes, 0)
if err != nil {
ic(err)
}

commands := extractSwitchCommands(file)
handlers := extractHandlers(file)
help := extractHelpCommands(src)

report := Report{
ds:     commands,
dlers:     handlers,
ds: help,
}

for _, cmd := range commands {
gs.HasPrefix(cmd, "-") || cmd == "help" || cmd == "version" {
tinue
"run" + toCamel(cmd)
== "plan" {
"runPlanner"
== "service-contracts" {
"runServiceContracts"
== "event-bus" {
"runEventBus"
== "kernel-registry" {
"runKernelRegistry"
dler(src, expected) && !contains(handlers, expected) {
gHandlers = append(report.MissingHandlers, expected)
h := range handlers {
gs.Contains(src, h+"(cfg") && !strings.Contains(src, h+"(cfg,") {
usedHandlers = append(report.UnusedHandlers, h)
cmd := range commands {
tainsHelp(help, cmd) && cmd != "-h" && cmd != "--help" {
gHelp = append(report.MissingHelp, cmd)
h := range help {
strings.Fields(h)
(root) > 0 && !contains(commands, root[0]) {
append(report.ExtraHelp, h)
gs(report.MissingHandlers)
sort.Strings(report.UnusedHandlers)
sort.Strings(report.MissingHelp)
sort.Strings(report.ExtraHelp)

out := render(report)
if err := os.WriteFile("reports/phase7-cli-self-verifier-report.md", []byte(out), 0644); err != nil {
ic(err)
}

fmt.Print(out)

if len(report.MissingHandlers) > 0 {
c extractSwitchCommands(file *ast.File) []string {
seen := map[string]bool{}
var out []string
ast.Inspect(file, func(n ast.Node) bool {
:= n.(*ast.CaseClause)
{
 true
expr := range cc.List {
:= expr.(*ast.BasicLit)
{
tinue
strings.Trim(lit.Value, `"`)
!= "" && !seen[v] {
[v] = true
append(out, v)
 true
})
sort.Strings(out)
return out
}

func extractHandlers(file *ast.File) []string {
var out []string
for _, decl := range file.Decls {
, ok := decl.(*ast.FuncDecl)
&& strings.HasPrefix(fn.Name.Name, "run") {
append(out, fn.Name.Name)
gs(out)
return out
}

func extractHelpCommands(src string) []string {
re := regexp.MustCompile(`fmt\.Println\("  ([^"]+)"\)`)
matches := re.FindAllStringSubmatch(src, -1)
var out []string
for _, m := range matches {
append(out, m[1])
}
sort.Strings(out)
return out
}

func callsHandler(src, name string) bool {
return strings.Contains(src, name+"(cfg")
}

func toCamel(s string) string {
parts := strings.FieldsFunc(s, func(r rune) bool { return r == '-' || r == '_' })
for i, p := range parts {
== "" {
tinue
strings.ToUpper(p[:1]) + p[1:]
}
return strings.Join(parts, "")
}

func contains(xs []string, x string) bool {
for _, v := range xs {
== x {
 true
 false
}

func containsHelp(help []string, cmd string) bool {
for _, h := range help {
gs.Fields(h)[0] == cmd {
 true
 false
}

func render(r Report) string {
var b strings.Builder
b.WriteString("# Phase 7 CLI Self-Verifier Report\n\n")
b.WriteString("## Commands\n")
for _, x := range r.Commands { b.WriteString("- " + x + "\n") }
b.WriteString("\n## Missing Handlers\n")
if len(r.MissingHandlers)==0 { b.WriteString("- none\n") }
for _, x := range r.MissingHandlers { b.WriteString("- " + x + "\n") }
b.WriteString("\n## Defined Handlers\n")
for _, x := range r.Handlers { b.WriteString("- " + x + "\n") }
b.WriteString("\n## Missing Help Entries\n")
if len(r.MissingHelp)==0 { b.WriteString("- none\n") }
for _, x := range r.MissingHelp { b.WriteString("- " + x + "\n") }
b.WriteString("\n## Extra Help Entries\n")
if len(r.ExtraHelp)==0 { b.WriteString("- none\n") }
for _, x := range r.ExtraHelp { b.WriteString("- " + x + "\n") }
return b.String()
}
GO

go run ./tools/cli-self-verifier || true

go test ./tools/cli-self-verifier

git add tools/cli-self-verifier/main.go reports/phase7-cli-self-verifier-report.md
git commit -m "test: add CLI self verifier" || true
git push --force-with-lease origin "$(git branch --show-current)"
