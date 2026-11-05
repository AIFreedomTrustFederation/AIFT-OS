package uxihttp

import (
    "os"
    "strings"
    "testing"
)

func TestOperatorConsoleKeyboardAccessibility(t *testing.T) {
    html, err := os.ReadFile("index.html")
    if err != nil { t.Fatal(err) }
    for _, required := range []string{
        `href="#main">Skip to main content</a>`,
        `<main class="main" id="main" tabindex="-1">`,
        `:focus-visible`,
        `@media(prefers-reduced-motion:reduce)`,
        `animation-duration:.01ms!important`,
        `transition-duration:.01ms!important`,
    } {
        if !strings.Contains(string(html), required) {
            t.Errorf("console missing accessibility affordance %q", required)
        }
    }
}
