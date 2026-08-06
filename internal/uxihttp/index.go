package uxihttp

import (
	_ "embed"
	"strings"
)

//go:embed index.html
var baseIndexHTML string

//go:embed world.css
var worldCSS string

//go:embed world.html
var worldHTML string

//go:embed world.js
var worldJS string

const gameLauncherJS = `
(function(){
  const routes={treeTab:"/tree",worldTab:"/world"};
  for(const [id,route] of Object.entries(routes)){
    const button=document.getElementById(id);
    if(!button)continue;
    button.addEventListener("click",function(event){
      event.preventDefault();
      event.stopImmediatePropagation();
      window.location.assign(route);
    },true);
  }

  const inspect=new URLSearchParams(window.location.search).get("inspect");
  const input=document.getElementById("message");
  if(inspect&&input){
    input.value="/inspect "+inspect;
    input.focus();
    window.history.replaceState({},"",window.location.pathname);
  }
})();`

var indexHTML = composeIndexHTML()

func composeIndexHTML() string {
	html := injectIndexFragment(baseIndexHTML, "</style>", worldCSS+"\n</style>")
	html = injectIndexFragment(html,
		`<button id="treeTab" role="tab" aria-selected="false">Tree of Life</button></nav>`,
		`<button id="treeTab" role="tab" aria-selected="false">Tree Game</button><button id="worldTab" role="tab" aria-selected="false">World Game</button></nav>`,
	)
	html = injectIndexFragment(html, `<div class="composer" id="composerWrap">`, worldHTML+`\n      <div class="composer" id="composerWrap">`)
	html = injectIndexFragment(html, "</body>", "  <script>\n"+worldJS+"\n"+gameLauncherJS+"\n  </script>\n</body>")
	return html
}

func injectIndexFragment(source, marker, replacement string) string {
	if !strings.Contains(source, marker) {
		panic("MoBox UXI composition marker not found: " + marker)
	}
	return strings.Replace(source, marker, replacement, 1)
}
