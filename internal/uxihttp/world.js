
    const federationWorld={data:null,deviceRaw:null,device:null,precision:"city",loading:false};
    const previousSetView=setView;
    setView=function(view){
      if(view!=="world"){
        $("#worldView").classList.remove("active");
        $("#worldTab").classList.remove("active");
        $("#worldTab").setAttribute("aria-selected","false");
        previousSetView(view);
        return;
      }
      $("#conversationView").classList.remove("active");
      $("#treeView").classList.remove("active");
      $("#worldView").classList.add("active");
      $("#conversationTab").classList.remove("active");
      $("#treeTab").classList.remove("active");
      $("#worldTab").classList.add("active");
      $("#conversationTab").setAttribute("aria-selected","false");
      $("#treeTab").setAttribute("aria-selected","false");
      $("#worldTab").setAttribute("aria-selected","true");
      $("#composerWrap").classList.add("hidden");
      $("#main").classList.add("tree-mode");
      closeDrawers();
      loadFederationWorld().catch(reportError);
    };

    function worldCoordinateDecimals(precision){
      if(precision==="country")return 0;
      if(precision==="region")return 1;
      if(precision==="exact")return 5;
      return 2;
    }
    function worldRound(value,precision){const factor=10**worldCoordinateDecimals(precision);return Math.round(value*factor)/factor}
    function worldProject(latitude,longitude){return{x:((longitude+180)/360)*1000,y:((90-latitude)/180)*500}}
    function worldText(value,max=24){const text=String(value||"");return text.length>max?text.slice(0,max-1)+"…":text}
    function temporaryWorldRepositories(){return federationWorld.data?(federationWorld.data.unmapped||[]).filter(item=>item.state==="missing"):[]}

    async function loadFederationWorld(force=false){
      if(federationWorld.loading)return;
      if(federationWorld.data&&!force){renderFederationWorld();return}
      federationWorld.loading=true;$("#worldStatus").textContent="Reading local location declarations…";
      try{federationWorld.data=await api("/v1/federation/world");renderFederationWorld()}finally{federationWorld.loading=false}
    }

    function renderFederationWorld(){
      if(!federationWorld.data)return;
      renderWorldGrid();renderWorldStats();renderWorldMap();renderWorldLists();
      if(!federationWorld.device)$("#worldStatus").textContent=`${federationWorld.data.progress.mapped} declared nodes · GPS has not been requested.`;
    }

    function renderWorldGrid(){
      const grid=$("#worldGrid");clear(grid);
      for(let longitude=-150;longitude<=150;longitude+=30){const x=worldProject(0,longitude).x;grid.append(svgElement("line",{x1:x,y1:0,x2:x,y2:500,class:"world-grid"}))}
      for(let latitude=-60;latitude<=60;latitude+=30){const y=worldProject(latitude,0).y;grid.append(svgElement("line",{x1:0,y1:y,x2:1000,y2:y,class:latitude===0?"world-equator":"world-grid"}))}
    }

    function renderWorldStats(){
      const progress=federationWorld.data.progress;const temporary=federationWorld.device?temporaryWorldRepositories().length:0;
      const stats=[["World level",progress.level],["Mapping XP",progress.xp],["Declared nodes",`${progress.mapped}/${progress.repositories}`],["Temporary GPS",temporary],["Open quests",progress.open_quests]];
      const root=$("#worldStats");clear(root);
      for(const [label,value] of stats){const card=element("div","world-stat");card.append(element("small","",label),element("strong","",String(value)));root.append(card)}
    }

    function worldClusters(){
      const groups=new Map();
      for(const node of federationWorld.data.nodes||[]){const key=`${node.latitude}:${node.longitude}`;if(!groups.has(key))groups.set(key,[]);groups.get(key).push(node)}
      return [...groups.values()];
    }

    function clusterStatus(nodes){if(nodes.some(node=>node.status==="blocked"))return"blocked";if(nodes.every(node=>node.status==="ready"))return"ready";return"detected"}

    function renderWorldMap(){
      const markers=$("#worldMarkers"),connections=$("#worldConnections");clear(markers);clear(connections);
      const clusters=worldClusters();
      if(federationWorld.device){
        const origin=worldProject(federationWorld.device.latitude,federationWorld.device.longitude);
        for(const cluster of clusters){const target=worldProject(cluster[0].latitude,cluster[0].longitude);if(Math.abs(target.x-origin.x)<1&&Math.abs(target.y-origin.y)<1)continue;const curve=(origin.y+target.y)/2-35;connections.append(svgElement("path",{d:`M ${origin.x} ${origin.y} Q ${(origin.x+target.x)/2} ${curve} ${target.x} ${target.y}`,class:"world-route"}))}
      }
      for(const cluster of clusters)markers.append(worldClusterMarker(cluster));
      if(federationWorld.device)markers.append(worldDeviceMarker());
    }

    function worldClusterMarker(nodes){
      const location=worldProject(nodes[0].latitude,nodes[0].longitude);const status=clusterStatus(nodes);const group=svgElement("g",{class:`world-marker ${status}`,transform:`translate(${location.x} ${location.y})`,tabindex:"0",role:"button","aria-label":`${nodes.length} federation repositories at ${nodes[0].label}`});
      group.append(svgElement("circle",{r:nodes.length>1?12:9,fill:status==="ready"?"#1e4b2a":status==="blocked"?"#542323":"#4d3d17",stroke:"currentColor","stroke-width":2.4}));
      const count=svgElement("text",{y:3});count.textContent=nodes.length>1?String(nodes.length):"●";group.append(count);
      const label=svgElement("text",{y:24});label.textContent=worldText(nodes[0].label,20);group.append(label);
      const select=()=>selectWorldCluster(nodes);group.addEventListener("click",select);group.addEventListener("keydown",event=>{if(event.key==="Enter"||event.key===" "){event.preventDefault();select()}});return group;
    }

    function worldDeviceMarker(){
      const temporary=temporaryWorldRepositories();const location=worldProject(federationWorld.device.latitude,federationWorld.device.longitude);const group=svgElement("g",{class:"world-marker device",transform:`translate(${location.x} ${location.y})`,tabindex:"0",role:"button","aria-label":`This device and ${temporary.length} temporarily anchored repositories`});
      group.append(svgElement("circle",{r:19,class:"world-device-ring"}));group.append(svgElement("circle",{r:10,fill:"#0b4544",stroke:"#5de4d7","stroke-width":2.8}));
      const dot=svgElement("text",{y:3});dot.textContent=temporary.length?String(temporary.length):"◎";group.append(dot);
      const label=svgElement("text",{y:25});label.textContent=temporary.length?`You + ${temporary.length} local`:"Your device";group.append(label);
      const select=()=>selectDeviceLocation();group.addEventListener("click",select);group.addEventListener("keydown",event=>{if(event.key==="Enter"||event.key===" "){event.preventDefault();select()}});return group;
    }

    function renderWorldLists(){
      const nodeList=$("#worldNodeList");clear(nodeList);
      if(federationWorld.device){const temporary=temporaryWorldRepositories();const button=element("button","world-row");button.type="button";button.append(element("strong","",`This phone · ${temporary.length} temporary repositories`),element("span","",`${federationWorld.device.latitude}, ${federationWorld.device.longitude} · browser memory only`));button.addEventListener("click",selectDeviceLocation);nodeList.append(button)}
      for(const node of federationWorld.data.nodes||[]){const button=element("button","world-row");button.type="button";button.append(element("strong","",node.repository),element("span","",`${node.label} · ${node.precision} · ${node.visibility}`));button.addEventListener("click",()=>selectWorldCluster([node]));nodeList.append(button)}
      for(const item of federationWorld.data.unmapped||[]){const button=element("button","world-row");button.type="button";button.append(element("strong","",`${item.repository} · ${item.state}`),element("span","",item.reason));button.addEventListener("click",()=>selectUnmappedWorldNode(item));nodeList.append(button)}
      if(!nodeList.firstChild)nodeList.append(element("div","world-empty","No geographic repository evidence has been declared."));

      const questList=$("#worldQuestList");clear(questList);const quests=[...(federationWorld.data.quests||[])].sort((a,b)=>(a.status==="open"?0:1)-(b.status==="open"?0:1)||a.repository.localeCompare(b.repository));
      for(const quest of quests){const button=element("button",`world-quest ${quest.status}`);button.type="button";button.append(element("strong","",quest.title),element("span","",quest.repository),element("span","",quest.status==="complete"?`✓ ${quest.evidence||"evidence observed"}`:`${quest.description} · ${quest.reward_xp} XP`));button.addEventListener("click",()=>{const item=(federationWorld.data.unmapped||[]).find(entry=>entry.repository_id===quest.repository_id);const node=(federationWorld.data.nodes||[]).find(entry=>entry.repository_id===quest.repository_id);if(node)selectWorldCluster([node]);else if(item)selectUnmappedWorldNode(item)});questList.append(button)}
    }

    function worldDetailCard(title,rows){const card=element("div","node-detail-card");card.append(element("strong","",title));const grid=element("div","node-kv");for(const [label,value] of rows)grid.append(element("span","",label),element("span","",String(value)));card.append(grid);return card}

    function selectWorldCluster(nodes){
      const root=$("#selectedNode");clear(root);const first=nodes[0];root.append(worldDetailCard(nodes.length>1?`${nodes.length} repositories · ${first.label}`:first.repository,[["Location",first.label],["Coordinates",`${first.latitude}, ${first.longitude}`],["Precision",first.precision],["Visibility",first.visibility],["Source",first.source],["Status",clusterStatus(nodes)]]));
      for(const node of nodes){const row=element("div","evidence-card");row.append(element("strong","",node.repository),document.createElement("br"),document.createTextNode(`${node.role} · ${node.status} · ${node.evidence}`));root.append(row)}
      const evidence=$("#evidence");clear(evidence);evidence.append(evidenceCard({summary:"Declared geographic evidence",source:first.evidence,detail:`${first.label} at ${first.precision} precision`}));document.body.classList.add("right-open");
    }

    function selectDeviceLocation(){
      if(!federationWorld.device)return;const temporary=temporaryWorldRepositories();const root=$("#selectedNode");clear(root);root.append(worldDetailCard("Current device GPS",[["Coordinates",`${federationWorld.device.latitude}, ${federationWorld.device.longitude}`],["Precision",federationWorld.precision],["Accuracy reported",`${Math.round(federationWorld.device.accuracy||0)} m`],["Storage","browser memory only"],["Temporary repositories",temporary.length]]));
      for(const item of temporary.slice(0,12)){const row=element("div","evidence-card");row.append(element("strong","",item.repository),document.createElement("br"),document.createTextNode("Temporarily anchored to this phone; no repository location declaration exists."));root.append(row)}document.body.classList.add("right-open");
    }

    function selectUnmappedWorldNode(item){const root=$("#selectedNode");clear(root);root.append(worldDetailCard(item.repository,[["Map state",item.state],["Reason",item.reason],["Role",item.role],["Repository status",item.status]]));document.body.classList.add("right-open")}

    function requestDeviceLocation(){
      if(!navigator.geolocation){reportError(new Error("This browser does not expose geolocation."));return}
      federationWorld.precision=$("#worldPrecision").value;$("#worldStatus").textContent="Waiting for GPS permission…";
      navigator.geolocation.getCurrentPosition(position=>{
        federationWorld.deviceRaw={latitude:position.coords.latitude,longitude:position.coords.longitude,accuracy:position.coords.accuracy};applyDevicePrecision();$("#worldStatus").textContent=`GPS active · ${federationWorld.precision} precision · memory only`;renderFederationWorld();selectDeviceLocation();
      },error=>{const names={1:"permission denied",2:"position unavailable",3:"request timed out"};$("#worldStatus").textContent=`GPS ${names[error.code]||"failed"}.`;reportError(new Error(`GPS ${names[error.code]||error.message}`))},{enableHighAccuracy:federationWorld.precision==="exact",timeout:15000,maximumAge:60000});
    }

    function applyDevicePrecision(){if(!federationWorld.deviceRaw)return;federationWorld.device={latitude:worldRound(federationWorld.deviceRaw.latitude,federationWorld.precision),longitude:worldRound(federationWorld.deviceRaw.longitude,federationWorld.precision),accuracy:federationWorld.deviceRaw.accuracy}}
    function forgetDeviceLocation(){federationWorld.deviceRaw=null;federationWorld.device=null;renderFederationWorld();$("#worldStatus").textContent="GPS forgotten. No device coordinates remain in this page."}

    $("#worldTab").addEventListener("click",()=>setView("world"));
    $("#worldLocate").addEventListener("click",requestDeviceLocation);
    $("#worldForget").addEventListener("click",forgetDeviceLocation);
    $("#worldRefresh").addEventListener("click",()=>loadFederationWorld(true).catch(reportError));
    $("#worldPrecision").addEventListener("change",event=>{federationWorld.precision=event.target.value;applyDevicePrecision();if(federationWorld.device){$("#worldStatus").textContent=`GPS active · ${federationWorld.precision} precision · memory only`;renderFederationWorld()}});
