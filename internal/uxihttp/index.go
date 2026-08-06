package uxihttp

import (
	_ "embed"
)

//go:embed index.html
var baseIndexHTML string

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
	html := injectIndexFragment(baseIndexHTML,
		`<button id="treeTab" role="tab" aria-selected="false">Tree of Life</button></nav>`,
		`<button id="treeTab" role="tab" aria-selected="false">Tree Game</button><button id="worldTab" role="tab" aria-selected="false">World Game</button></nav>`,
	)
	return injectIndexFragment(html, "</body>", "  <script>\n"+gameLauncherJS+"\n  </script>\n</body>")
}

func injectIndexFragment(source, marker, replacement string) string {
	if !strings.Contains(source, marker) {
		panic("MoBox UXI composition marker not found: " + marker)
	}
	return strings.Replace(source, marker, replacement, 1)
}
