package uxihttp

import _ "embed"

//go:embed index.html
var indexHTML string

//go:embed styles.css
var stylesCSS string

//go:embed app.js
var appJS string
