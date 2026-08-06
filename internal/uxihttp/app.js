const state = { session: null, sessions: [], repos: [], adapters: [], sources: [] };
const $ = selector => document.querySelector(selector);

function element(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text !== undefined) node.textContent = text;
  return node;
}

function clear(node) { while (node.firstChild) node.removeChild(node.firstChild); }

async function api(path, options = {}) {
  const response = await fetch(path, { headers: { "Content-Type": "application/json" }, ...options });
  const data = await response.json();
  if (!response.ok) throw new Error(data.error || response.statusText);
  return data;
}

async function load() {
  const [system, sessions, repos, adapters, sources] = await Promise.all([
    api("/v1/system"),
    api("/v1/sessions"),
    api("/v1/repositories"),
    api("/v1/adapters"),
    api("/v1/sources")
  ]);
  state.sessions = sessions.sessions || [];
  state.repos = repos.repositories || [];
  state.adapters = adapters.adapters || [];
  state.sources = sources.sources || [];
  $("#pulse").textContent = `active · ${system.repository_count} repos · ${state.adapters.length} read-only adapters`;
  const systemNode = $("#system");
  clear(systemNode);
  systemNode.append(element("strong", "", system.status), document.createElement("br"), document.createTextNode(system.truth_contract), document.createElement("br"), document.createElement("br"), element("span", "badge", system.aift_root));
  renderSessions();
  renderRepos();
  renderAdapterOptions();
  renderTargetOptions();
  if (state.sessions[0]) await openSession(state.sessions[0].id);
  else renderTurns();
}

function renderSessions() {
  const root = $("#sessions");
  clear(root);
  for (const session of state.sessions) {
    const button = element("button", `nav-button ${state.session && state.session.id === session.id ? "active" : ""}`, session.title);
    button.append(element("small", "", new Date(session.updated_at).toLocaleString()));
    button.addEventListener("click", () => openSession(session.id));
    root.append(button);
  }
}

function renderAdapterOptions() {
  const select = $("#actionKind");
  clear(select);
  for (const adapter of state.adapters) {
    const option = element("option", "", adapter.kind);
    option.value = adapter.kind;
    option.title = adapter.description;
    select.append(option);
  }
}

function renderTargetOptions() {
  const list = $("#actionTargets");
  clear(list);
  for (const repo of state.repos) {
    const option = element("option");
    option.value = repo.name;
    option.label = `${repo.role} · repository`;
    list.append(option);
  }
  for (const source of state.sources) {
    const option = element("option");
    option.value = source.id;
    option.label = `${source.repository} · ${source.description || source.kind} · ${source.status}`;
    list.append(option);
  }
}

function renderRepos() {
  const root = $("#repos");
  clear(root);
  for (const repo of state.repos) {
    const button = element("button", "nav-button", repo.name);
    button.append(element("small", "", `${repo.role} · ${repo.status}`));
    button.addEventListener("click", () => {
      $("#message").value = `/inspect ${repo.name}`;
      showRepo(repo.name);
      closeDrawers();
      $("#message").focus();
    });
    root.append(button);
  }
}

function showRepo(name) {
  const repo = state.repos.find(item => item.name === name);
  const root = $("#evidence");
  clear(root);
  if (!repo || !(repo.evidence || []).length) {
    root.textContent = "No evidence records.";
    return;
  }
  for (const evidence of repo.evidence) root.append(evidenceCard(evidence));
}

function evidenceCard(evidence) {
  const card = element("div", "evidence-card");
  card.append(element("strong", "", evidence.summary), document.createElement("br"), document.createTextNode(evidence.source));
  if (evidence.detail) card.append(document.createElement("br"), document.createTextNode(evidence.detail));
  return card;
}

async function createSession() {
  const session = await api("/v1/sessions", { method: "POST", body: JSON.stringify({ title: "MoBox workspace" }) });
  state.sessions.unshift(session);
  await openSession(session.id);
}

async function openSession(id) {
  state.session = await api(`/v1/sessions/${encodeURIComponent(id)}`);
  const index = state.sessions.findIndex(item => item.id === state.session.id);
  if (index >= 0) state.sessions[index] = state.session;
  renderSessions();
  renderTurns();
  renderGovernance();
  closeDrawers();
}

function renderTurns() {
  const root = $("#stream");
  clear(root);
  if (!state.session || !(state.session.turns || []).length) {
    const empty = element("div", "empty");
    empty.append(element("h1", "", "Begin with evidence."), element("p", "", "Try /inspect AIFT-OS, /forge, or ask the local model about the federation."));
    root.append(empty);
    return;
  }
  for (const turn of state.session.turns) {
    const card = element("article", `turn ${turn.role === "user" ? "user" : "assistant"}`);
    const mode = turn.metadata && turn.metadata.mode ? ` · ${turn.metadata.mode}` : "";
    card.append(element("div", "turn-role", `${turn.role}${mode}`), document.createTextNode(turn.content));
    for (const evidence of turn.evidence || []) card.append(evidenceCard(evidence));
    root.append(card);
  }
  root.scrollTop = root.scrollHeight;
}

function renderGovernance() {
  const root = $("#governance");
  clear(root);
  if (!state.session) {
    root.textContent = "Read-only adapters create supervised jobs and evidence artifacts.";
    return;
  }
  for (const plan of state.session.plans || []) {
    const card = element("div", "governance-card");
    card.append(element("strong", "", `Plan · ${plan.status}`), document.createElement("br"), document.createTextNode(plan.objective));
    for (const step of plan.steps || []) card.append(document.createElement("br"), element("small", "", `${step.title} · ${step.status}`));
    root.append(card);
  }
  for (const action of state.session.actions || []) {
    const card = element("div", "governance-card");
    card.append(element("strong", "", `Action · ${action.status}`), document.createElement("br"), document.createTextNode(`${action.kind} → ${action.target}`), document.createElement("br"), element("small", "", `risk ${action.risk}`));
    const adapter = state.adapters.find(item => item.kind === action.kind);
    if (action.status === "proposed" || action.status === "awaiting_approval") {
      if (action.approval_required) {
        const approve = element("button", "decision-button approve", "Approve record");
        approve.addEventListener("click", () => decideAction(action.id, "approved"));
        const reject = element("button", "decision-button reject", "Reject");
        reject.addEventListener("click", () => decideAction(action.id, "rejected"));
        card.append(document.createElement("br"), approve, reject);
      } else if (adapter && !adapter.mutating && action.status === "proposed") {
        const run = element("button", "decision-button approve", "Run read-only");
        run.addEventListener("click", () => invokeAction(action.id));
        card.append(document.createElement("br"), run);
      }
    } else if (action.status === "approved" && adapter && !adapter.mutating) {
      const run = element("button", "decision-button approve", "Run approved read-only");
      run.addEventListener("click", () => invokeAction(action.id));
      card.append(document.createElement("br"), run);
    }
    root.append(card);
  }
  for (const job of state.session.jobs || []) {
    const card = element("div", "governance-card");
    card.append(element("strong", "", `Job · ${job.status}`), document.createElement("br"), document.createTextNode(`${job.adapter_kind} · ${job.id}`));
    if (job.error) card.append(document.createElement("br"), element("small", "", job.error));
    if (job.result_artifact_id) {
      const view = element("button", "decision-button", "View evidence artifact");
      view.addEventListener("click", () => viewArtifact(job.result_artifact_id));
      card.append(document.createElement("br"), view);
    }
    root.append(card);
  }
  if (!root.firstChild) root.textContent = "Read-only adapters create supervised jobs and evidence artifacts.";
}

async function invokeAction(actionID) {
  const response = await api(`/v1/sessions/${encodeURIComponent(state.session.id)}/actions/${encodeURIComponent(actionID)}/invoke`, { method: "POST" });
  state.session = response.session || response;
  renderSessions();
  renderGovernance();
}

async function viewArtifact(artifactID) {
  const artifact = await api(`/v1/artifacts/${encodeURIComponent(artifactID)}`);
  const root = $("#evidence");
  clear(root);
  const card = element("div", "evidence-card");
  card.append(element("strong", "", `Artifact ${artifactID}`), document.createElement("br"), document.createTextNode(JSON.stringify(artifact, null, 2)));
  root.append(card);
  document.body.classList.add("right-open");
}

async function decideAction(actionID, decision) {
  state.session = await api(`/v1/sessions/${encodeURIComponent(state.session.id)}/actions/${encodeURIComponent(actionID)}/decision`, {
    method: "POST",
    body: JSON.stringify({ decision, actor: "human-local-operator" })
  });
  renderSessions();
  renderGovernance();
}

$("#newSession").addEventListener("click", createSession);
$("#composer").addEventListener("submit", async event => {
  event.preventDefault();
  const input = $("#message");
  const content = input.value.trim();
  if (!content) return;
  if (!state.session) await createSession();
  input.value = "";
  input.disabled = true;
  try {
    state.session = await api(`/v1/sessions/${encodeURIComponent(state.session.id)}/messages`, { method: "POST", body: JSON.stringify({ content }) });
    const index = state.sessions.findIndex(item => item.id === state.session.id);
    if (index >= 0) state.sessions[index] = state.session;
    renderSessions();
    renderTurns();
    renderGovernance();
  } catch (error) {
    window.alert(error.message);
  } finally {
    input.disabled = false;
    input.focus();
  }
});

$("#planForm").addEventListener("submit", async event => {
  event.preventDefault();
  if (!state.session) await createSession();
  const objective = $("#planObjective").value.trim();
  const steps = $("#planSteps").value.split("\n").map(value => value.trim()).filter(Boolean).map(title => ({ title }));
  state.session = await api(`/v1/sessions/${encodeURIComponent(state.session.id)}/plans`, { method: "POST", body: JSON.stringify({ objective, steps }) });
  event.target.reset();
  renderGovernance();
});

$("#actionForm").addEventListener("submit", async event => {
  event.preventDefault();
  if (!state.session) await createSession();
  const payload = {
    kind: $("#actionKind").value.trim(),
    target: $("#actionTarget").value.trim(),
    risk: $("#actionRisk").value,
    approval_required: $("#actionApproval").checked
  };
  state.session = await api(`/v1/sessions/${encodeURIComponent(state.session.id)}/actions`, { method: "POST", body: JSON.stringify(payload) });
  event.target.reset();
  $("#actionApproval").checked = true;
  renderGovernance();
});

function closeDrawers() { document.body.classList.remove("left-open", "right-open"); }
$("#leftToggle").addEventListener("click", () => document.body.classList.toggle("left-open"));
$("#rightToggle").addEventListener("click", () => document.body.classList.toggle("right-open"));
$("#overlay").addEventListener("click", closeDrawers);

load().catch(error => {
  $("#pulse").textContent = "blocked";
  $("#system").textContent = error.message;
});
