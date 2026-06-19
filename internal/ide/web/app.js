/* DataDream Studio */

function monacoBasePath() {
  return 'vendor/monaco/min/vs';
}

async function loadMonaco() {
  const base = monacoBasePath();
  return new Promise((resolve, reject) => {
    if (typeof require === 'undefined') {
      reject(new Error('Monaco loader missing — run scripts/fetch-monaco before building the IDE'));
      return;
    }
    require.config({ paths: { vs: base } });
    require(['vs/editor/editor.main'], monaco => resolve(monaco), reject);
  });
}

function wailsApp() {
  return window.go?.main?.App ?? null;
}

function createAPI() {
  const app = wailsApp();
  if (app) {
    return {
      isDesktop: true,
      version: () => app.Version(),
      tree: () => app.Tree(''),
      search: q => app.Search(q),
      read: path => app.Read(path),
      write: (path, content) => app.Write(path, content).then(() => ({ ok: true })),
      newFile: (path, template) => app.NewFile(path, template || ''),
      check: (path, content) => app.Check(path, content),
      run: (path, content) => Promise.resolve(app.Run(path, content)),
      build: (path, content, release) => Promise.resolve(app.Build(path, content, !!release)),
      doctor: () => app.Doctor(),
      openProject: () => app.OpenProject(),
    };
  }
  return {
    isDesktop: false,
    version: () => fetch('/api/version').then(r => r.json()),
    tree: () => fetch('/api/tree').then(r => r.json()),
    search: q => fetch(`/api/search?q=${encodeURIComponent(q)}`).then(r => r.json()),
    read: path => fetch(`/api/read?path=${encodeURIComponent(path)}`).then(async r => {
      if (!r.ok) throw new Error((await r.json()).error);
      return r.json();
    }),
    write: (path, content) => fetch('/api/write', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path, content }),
    }).then(async r => { if (!r.ok) throw new Error((await r.json()).error); return r.json(); }),
    newFile: (path, template) => fetch('/api/new', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path, template }),
    }).then(async r => { if (!r.ok) throw new Error((await r.json()).error); return r.json(); }),
    check: (path, content) => fetch('/api/check', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path, content }),
    }).then(r => r.json()),
    run: (path, content) => fetch('/api/run', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path, content }),
    }).then(r => r.json()),
    build: (path, content, release) => fetch('/api/build', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path, content, release }),
    }).then(r => r.json()),
    doctor: () => fetch('/api/doctor').then(r => r.json()),
    openProject: null,
  };
}

let API = createAPI();

async function waitForDesktopBackend() {
  if (wailsApp()) return true;
  for (let i = 0; i < 80; i++) {
    await new Promise(r => setTimeout(r, 25));
    if (wailsApp()) return true;
  }
  return false;
}

function refreshAPI() {
  if (wailsApp()) {
    API = createAPI();
    document.documentElement.classList.add('desktop-app');
  }
}

function applyProjectInfo(info) {
  document.getElementById('status-version').textContent = `v${info.version}`;
  document.getElementById('status-workspace').textContent = info.root;
  document.getElementById('status-workspace').title = info.root;
  document.getElementById('status-branch').textContent = info.name;
}

const state = {
  tabs: new Map(),
  activePath: null,
  editor: null,
  monaco: null,
  checkTimer: null,
  treeData: null,
  pendingClose: null,
  paletteMode: 'commands',
  paletteIndex: 0,
  quickOpenIndex: 0,
  sdkReady: null,
};

const COMMANDS = [
  { id: 'open', label: 'Quick Open File', keys: 'Ctrl+P', run: () => openQuickOpen() },
  { id: 'save', label: 'Save File', keys: 'Ctrl+S', run: () => saveFile() },
  { id: 'check', label: 'Check Syntax', keys: 'Ctrl+Shift+C', run: () => checkFile() },
  { id: 'run', label: 'Run Program', keys: 'Ctrl+Enter', run: () => runFile() },
  { id: 'build', label: 'Build Binary', keys: 'Ctrl+B', run: () => buildFile() },
  { id: 'new', label: 'New File', keys: '', run: () => showNewFileModal() },
  { id: 'openProject', label: 'Open Project Folder…', keys: '', run: () => openProjectFolder() },
  { id: 'close', label: 'Close Tab', keys: 'Ctrl+W', run: () => state.activePath && requestCloseTab(state.activePath) },
  { id: 'panel', label: 'Toggle Panel', keys: 'Ctrl+J', run: () => togglePanel() },
  { id: 'explorer', label: 'Show Explorer', keys: '', run: () => setActivityView('explorer') },
  { id: 'sdk', label: 'Show SDK Status', keys: '', run: () => setActivityView('sdk') },
  { id: 'shortcuts', label: 'Keyboard Shortcuts', keys: '?', run: () => showShortcuts() },
  { id: 'clear', label: 'Clear Output', keys: '', run: () => clearOutput() },
  { id: 'clicker', label: 'Open: clicker.dd', keys: '', run: () => openFile('examples/beginner/clicker.dd') },
  { id: 'hello', label: 'Open: hello_friendly.dd', keys: '', run: () => openFile('examples/raylib/hello_friendly.dd') },
];

const SHORTCUTS = [
  ['Quick Open', 'Ctrl+P'], ['Command Palette', 'Ctrl+Shift+P'], ['Save', 'Ctrl+S'],
  ['Run', 'Ctrl+Enter'], ['Check', 'Ctrl+Shift+C'], ['Build', 'Ctrl+B'],
  ['Close Tab', 'Ctrl+W'], ['Toggle Panel', 'Ctrl+J'], ['Shortcuts', '?'],
];

// ── Language & theme ──
function defineDataDreamLanguage(monaco) {
  monaco.languages.register({ id: 'datadream' });

  const drawAPI = ['text','rect','circle','line','sprite','fps','grid','texture'];
  const inputAPI = ['pressed','down','released','mouse','mousePressed','mouseReleased'];
  const colorNames = ['white','black','red','green','blue','gold','gray','darkgray','lightgray'];

  monaco.languages.setMonarchTokensProvider('datadream', {
    defaultToken: '',
    keywords: [
      'if','else','while','loop','for','in','break','continue','return','match','defer','try',
      'let','const','fn','struct','entity','enum','app','scene','system','spawn','destroy','self',
      'on','start','update','draw','window','ui','state','asset','preload','async','await',
      'data','shader','module','link','include','import','export','use','using','as','extern',
      'and','or','not','c','sync','rpc','arena','pool','network',
    ],
    types: ['bool','int','float','double','string','char','byte','void','cstring','voidptr','usize','isize',
      'i8','i16','i32','i64','u8','u16','u32','u64','f32','f64','ptr'],
    constants: ['true','false','null','none'],
    namespaces: ['draw','input','keys','screen','colors','random','time','math','audio','assets','collision','quit','vec2','vec3','distance','ui'],
    operators: ['..','..=','=>','==','!=','<=','>=','&&','||','+','-','*','/','%','=','<','>','!','&','|','?','.',':','@'],
    symbols: /[=><!~?:&|+\-*\/\^%.]+/,
    tokenizer: {
      root: [
        [/\/\/.*$/, 'comment'],
        [/\/\*/, 'comment', '@comment'],
        [/"/, 'string', '@string'],
        [/#([0-9A-Fa-f]{3,8})\b/, 'number.hex'],
        [/\b\d+\.\d+([fF])?\b/, 'number.float'],
        [/\b\d+\b/, 'number'],
        [/[{}()\[\]]/, '@brackets'],
        [/[;,.]/, 'delimiter'],
        [/\b([a-zA-Z_]\w*)\b/, { cases: {
          '@keywords': 'keyword', '@types': 'type', '@constants': 'constant',
          '@namespaces': 'namespace', '@default': 'identifier',
        }}],
        [/@symbols/, { cases: { '@operators': 'operator', '@default': '' }}],
      ],
      comment: [[/[^\/*]+/, 'comment'], [/\*\//, 'comment', '@pop'], [/[\/*]/, 'comment']],
      string: [[/[^\\"{]+/, 'string'], [/\\./, 'string.escape'], [/\{/, 'string.interpolated', '@interp'], [/"/, 'string', '@pop']],
      interp: [[/[^}]+/, 'identifier'], [/\}/, 'string.interpolated', '@pop']],
    },
  });

  monaco.languages.setLanguageConfiguration('datadream', {
    comments: { lineComment: '//', blockComment: ['/*', '*/'] },
    brackets: [['{','}'],['[',']'],['(',')']],
    autoClosingPairs: [
      { open: '{', close: '}' }, { open: '[', close: ']' }, { open: '(', close: ')' },
      { open: '"', close: '"' }, { open: "'", close: "'" },
    ],
    indentationRules: {
      increaseIndentPattern: /^\s*(fn|if|else|while|loop|for|match|app|scene|system|entity|draw|start|update|window|extern|struct|enum|\{)\b/,
      decreaseIndentPattern: /^\s*(else|\}|\))/,
    },
  });

  const snippets = [
    { label: 'app', doc: 'Friendly game app scaffold', text: 'app "${1:MyGame}";\n\nwindow {\n    size: 800, 600;\n    title: "${2:Game}";\n}\n\nstart {\n    $0\n}\n\nupdate {\n    if input.pressed("escape") { quit(); }\n}\n\ndraw {\n    clear(colors.black);\n}' },
    { label: 'fn main', doc: 'Main entry point', text: 'fn main() {\n    $0\n}' },
    { label: 'draw.text', doc: 'Draw text on screen', text: 'draw.text("${1:Hello}", {\n    position: vec2(${2:100}, ${3:100}),\n    size: ${4:24},\n    color: colors.white\n});' },
    { label: 'draw.circle', doc: 'Draw a circle', text: 'draw.circle({\n    position: vec2(${1:400}, ${2:300}),\n    radius: ${3:40},\n    color: colors.white\n});' },
    { label: 'entity', doc: 'ECS entity definition', text: 'entity ${1:Name} {\n    ${2:field}: ${3:type};\n\n    update {\n        $0\n    }\n}' },
    { label: 'match', doc: 'Pattern match', text: 'match ${1:expr} {\n    ${2:pattern} => {\n        $0\n    }\n}' },
    { label: 'defer', doc: 'Defer cleanup', text: 'defer ${1:CloseWindow}();' },
  ];

  monaco.languages.registerCompletionItemProvider('datadream', {
    triggerCharacters: ['.', '"'],
    provideCompletionItems(model, position) {
      const word = model.getWordUntilPosition(position);
      const range = { startLineNumber: position.lineNumber, endLineNumber: position.lineNumber, startColumn: word.startColumn, endColumn: word.endColumn };
      const line = model.getLineContent(position.lineNumber).slice(0, position.column - 1);
      const suggestions = [];

      for (const s of snippets) {
        suggestions.push({
          label: s.label, kind: monaco.languages.CompletionItemKind.Snippet,
          documentation: s.doc, insertText: s.text, range,
          insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
        });
      }

      if (line.endsWith('draw.')) {
        for (const m of drawAPI) suggestions.push({ label: m, kind: monaco.languages.CompletionItemKind.Method, insertText: m, range });
      }
      if (line.endsWith('input.')) {
        for (const m of inputAPI) suggestions.push({ label: m, kind: monaco.languages.CompletionItemKind.Method, insertText: m, range });
      }
      if (line.endsWith('colors.')) {
        for (const c of colorNames) suggestions.push({ label: c, kind: monaco.languages.CompletionItemKind.Color, insertText: c, range });
      }

      return { suggestions };
    },
  });

  monaco.languages.registerHoverProvider('datadream', {
    provideHover(model, pos) {
      const word = model.getWordAtPosition(pos);
      if (!word) return null;
      const docs = {
        app: 'Declare a friendly game application. Auto-enables raylib runtime.',
        draw: 'Friendly drawing namespace — draw.text, draw.circle, draw.rect, …',
        input: 'Input namespace — input.pressed(), input.mouse(), …',
        entity: 'Define an ECS entity with fields and lifecycle blocks.',
        defer: 'Run cleanup when leaving the current scope.',
        match: 'Pattern match on values, including struct destructuring.',
        loop: 'Infinite loop — use break to exit.',
      };
      if (docs[word.word]) {
        return { range: new monaco.Range(pos.lineNumber, word.startColumn, pos.lineNumber, word.endColumn), contents: [{ value: `**${word.word}**\n\n${docs[word.word]}` }] };
      }
      return null;
    },
  });
}

function defineTheme(monaco) {
  monaco.editor.defineTheme('datadream-dark', {
    base: 'vs-dark', inherit: true,
    rules: [
      { token: 'comment', foreground: '5C6B7A', fontStyle: 'italic' },
      { token: 'keyword', foreground: 'E88388' },
      { token: 'type', foreground: '6EC8D8' },
      { token: 'constant', foreground: 'E88388' },
      { token: 'namespace', foreground: '6EC8D8' },
      { token: 'string', foreground: '9FD89F' },
      { token: 'number.hex', foreground: '6EC8D8' },
      { token: 'operator', foreground: '6EC8D8' },
      { token: 'identifier', foreground: 'E8ECEF' },
    ],
    colors: {
      'editor.background': '#0B0E14',
      'editor.foreground': '#E8ECEF',
      'editor.lineHighlightBackground': '#141B24',
      'editor.selectionBackground': '#2A7C8844',
      'editorCursor.foreground': '#3DB8C9',
      'editorLineNumber.foreground': '#5C6B7A',
      'editorLineNumber.activeForeground': '#8B9AAB',
      'editorGutter.background': '#0B0E14',
      'editorWidget.background': '#141B24',
      'editorWidget.border': '#2A3544',
      'minimap.background': '#0B0E14',
    },
  });
}

// ── Editor ──
function initEditor(monaco) {
  state.monaco = monaco;
  defineDataDreamLanguage(monaco);
  defineTheme(monaco);

  state.editor = monaco.editor.create(document.getElementById('monaco-editor'), {
    language: 'datadream', theme: 'datadream-dark',
    fontFamily: "'JetBrains Mono', Consolas, monospace",
    fontSize: 13, lineHeight: 20, minimap: { enabled: false },
    scrollBeyondLastLine: false, automaticLayout: true, padding: { top: 4, bottom: 4 },
    renderLineHighlight: 'all', cursorBlinking: 'smooth', smoothScrolling: true,
    bracketPairColorization: { enabled: true }, guides: { bracketPairs: true, indentation: true },
    tabSize: 4, insertSpaces: true, folding: true, renderWhitespace: 'selection',
  });

  document.getElementById('editor-loading').classList.add('hidden');

  state.editor.onDidChangeCursorPosition(e => {
    document.getElementById('status-position').textContent = `Ln ${e.position.lineNumber}, Col ${e.position.column}`;
  });

  state.editor.onDidChangeModelContent(() => {
    if (!state.activePath) return;
    const tab = state.tabs.get(state.activePath);
    tab.dirty = true;
    tab.content = state.editor.getValue();
    renderTabs();
    updateDirtyStatus();
    scheduleCheck();
  });

  state.editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => saveFile());
  state.editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => runFile());
  state.editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyMod.Shift | monaco.KeyCode.KeyC, () => checkFile());
  state.editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyB, () => buildFile());
  state.editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyW, () => state.activePath && requestCloseTab(state.activePath));
  state.editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyJ, () => togglePanel());
  state.editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyP, () => openQuickOpen());
  state.editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyMod.Shift | monaco.KeyCode.KeyP, () => openCommandPalette());
}

// ── Tabs ──
function openTab(path, content) {
  if (!state.tabs.has(path)) state.tabs.set(path, { content, dirty: false });
  state.activePath = path;
  renderTabs();
  showEditor();
  updateBreadcrumb(path);

  const tab = state.tabs.get(path);
  if (!tab.model) {
    tab.model = state.monaco.editor.createModel(content, 'datadream', state.monaco.Uri.file(path));
  }
  state.editor.setModel(tab.model);
  document.getElementById('status-file').textContent = path;
  updateDirtyStatus();
  highlightTreeItem(path);
  checkFile();
}

function requestCloseTab(path) {
  const tab = state.tabs.get(path);
  if (tab?.dirty) {
    state.pendingClose = path;
    document.getElementById('unsaved-msg').textContent = `Save changes to "${basename(path)}" before closing?`;
    showOverlay('overlay-unsaved');
    return;
  }
  closeTab(path);
}

function closeTab(path) {
  const tab = state.tabs.get(path);
  if (tab?.model) tab.model.dispose();
  state.tabs.delete(path);

  if (state.activePath === path) {
    const keys = [...state.tabs.keys()];
    if (keys.length) openTab(keys[keys.length - 1], state.tabs.get(keys[keys.length - 1]).content);
    else {
      state.activePath = null;
      state.editor?.setModel(null);
      hideEditor();
      renderTabs();
      updateBreadcrumb(null);
      document.getElementById('status-file').textContent = 'No file open';
    }
  } else renderTabs();
}

function renderTabs() {
  const bar = document.getElementById('tab-bar');
  const empty = document.getElementById('tab-empty');
  bar.querySelectorAll('.tab').forEach(t => t.remove());

  if (!state.tabs.size) { empty.style.display = 'flex'; return; }
  empty.style.display = 'none';

  for (const [path, tab] of state.tabs) {
    const el = document.createElement('div');
    el.className = 'tab' + (path === state.activePath ? ' active' : '');
    el.innerHTML = `<span class="${tab.dirty ? 'tab-dirty' : ''}">${basename(path)}</span><button class="tab-close">×</button>`;
    el.addEventListener('click', () => { if (path !== state.activePath) openTab(path, tab.content); });
    el.querySelector('.tab-close').addEventListener('click', e => { e.stopPropagation(); requestCloseTab(path); });
    bar.appendChild(el);
  }
}

function basename(p) { return p.split(/[/\\]/).pop(); }
function showEditor() {
  document.getElementById('editor-welcome').classList.add('hidden');
  document.getElementById('monaco-editor').classList.add('visible');
}
function hideEditor() {
  document.getElementById('editor-welcome').classList.remove('hidden');
  document.getElementById('monaco-editor').classList.remove('visible');
}
function updateDirtyStatus() {
  const el = document.getElementById('status-dirty');
  if (!state.activePath) { el.textContent = ''; return; }
  el.textContent = state.tabs.get(state.activePath)?.dirty ? '● Unsaved' : '';
}

function updateBreadcrumb(path) {
  const bar = document.getElementById('breadcrumb-bar');
  if (!path) { bar.innerHTML = '<span class="breadcrumb-item">No file open</span>'; return; }
  const parts = path.split('/');
  bar.innerHTML = parts.map((p, i) =>
    `<span class="breadcrumb-item">${escapeHtml(p)}</span>${i < parts.length - 1 ? '<span class="breadcrumb-sep">›</span>' : ''}`
  ).join('');
}

async function openProjectFolder() {
  if (!API.openProject) {
    showToast('Use the desktop app to open project folders', 'error');
    return;
  }
  try {
    setStatus('Opening project…', 'busy');
    const info = await API.openProject();
    if (!info?.root) return;
    applyProjectInfo(info);
    state.tabs.clear();
    state.activePath = null;
    state.editor?.setValue('');
    document.getElementById('tab-bar').innerHTML = '<div class="tab-bar-empty" id="tab-empty">Open a .dd file — Ctrl+P to quick open</div>';
    document.getElementById('editor-welcome')?.classList.remove('hidden');
    document.getElementById('monaco-editor')?.classList.remove('visible');
    updateBreadcrumb(null);
    await loadTree();
    const d = await API.doctor();
    updateSDKStatus(d);
    setStatus('Ready', 'ready');
    showToast('Project: ' + info.root, 'success');
  } catch (err) {
    setStatus('Ready', 'error');
    showToast(String(err), 'error');
  }
}

// ── File tree ──
async function loadTree() {
  try {
    state.treeData = await API.tree();
    renderTree(state.treeData, document.getElementById('file-tree'), 0);
    if (state.activePath) highlightTreeItem(state.activePath);
  } catch (err) { showToast('Failed to load file tree: ' + err.message, 'error'); }
}

function renderTree(node, container, depth) {
  if (depth === 0) container.innerHTML = '';
  const filter = document.getElementById('tree-filter')?.value?.toLowerCase() || '';

  if (node.isDir) {
    if (depth === 0) { for (const c of node.children || []) renderTree(c, container, depth); return; }

    const wrap = document.createElement('div');
    const row = document.createElement('div');
    row.className = 'tree-item tree-dir';
    row.style.paddingLeft = `${8 + depth * 10}px`;
    row.dataset.path = node.path;
    row.innerHTML = `<span class="tree-chevron open">▶</span><svg class="tree-icon tree-icon-folder" viewBox="0 0 16 16"><path fill="currentColor" d="M1.5 2.75A1.25 1.25 0 0 1 2.75 1.5h3.5a.5.5 0 0 1 .4.2l1.3 1.74 1.3-1.74a.5.5 0 0 1 .4-.2h3.5A1.25 1.25 0 0 1 14.5 2.75v10.5A1.25 1.25 0 0 1 13.25 14.5H2.75A1.25 1.25 0 0 1 1.5 13.25V2.75z"/></svg><span>${node.name}</span>`;

    const kids = document.createElement('div');
    let open = true;
    row.addEventListener('click', e => {
      if (e.target.closest('.tree-chevron') || e.currentTarget === row) {
        open = !open; kids.style.display = open ? 'block' : 'none';
        row.querySelector('.tree-chevron').classList.toggle('open', open);
      }
    });
    for (const c of node.children || []) renderTree(c, kids, depth + 1);
    wrap.appendChild(row); wrap.appendChild(kids); container.appendChild(wrap);
    if (filter && !matchesFilter(node, filter)) wrap.classList.add('hidden-by-filter');
  } else {
    const row = document.createElement('div');
    row.className = 'tree-item tree-file';
    row.style.paddingLeft = `${8 + depth * 10}px`;
    row.dataset.path = node.path;
    row.innerHTML = `<span class="tree-chevron hidden">▶</span><svg class="tree-icon tree-icon-file" viewBox="0 0 16 16"><path fill="currentColor" d="M4 1.75A1.75 1.75 0 0 1 5.75 0h4.5A1.75 1.75 0 0 1 12 1.75V5H4V1.75zM4 6h8v8.25A1.75 1.75 0 0 1 10.25 16h-4.5A1.75 1.75 0 0 1 4 14.25V6z"/></svg><span>${node.name}</span>`;
    row.addEventListener('click', () => openFile(node.path));
    if (filter && !node.path.toLowerCase().includes(filter) && !node.name.toLowerCase().includes(filter)) {
      row.classList.add('hidden-by-filter');
    }
    container.appendChild(row);
  }
}

function matchesFilter(node, filter) {
  if (!node.isDir) return node.path.toLowerCase().includes(filter);
  return (node.children || []).some(c => matchesFilter(c, filter));
}

function filterTree() {
  if (state.treeData) renderTree(state.treeData, document.getElementById('file-tree'), 0);
}

function highlightTreeItem(path) {
  document.querySelectorAll('.tree-item').forEach(el => el.classList.toggle('active', el.dataset.path === path));
}

async function openFile(path) {
  setStatus('Opening…');
  try {
    if (state.tabs.has(path)) openTab(path, state.tabs.get(path).content);
    else { const data = await API.read(path); openTab(path, data.content); }
    setStatus('Ready');
  } catch (err) { showToast('Cannot open: ' + err.message, 'error'); setStatus('Error'); }
}

// ── Actions ──
async function saveFile() {
  if (!state.activePath) return;
  setStatus('Saving…');
  try {
    const content = state.editor.getValue();
    await API.write(state.activePath, content);
    const tab = state.tabs.get(state.activePath);
    tab.content = content; tab.dirty = false;
    renderTabs(); updateDirtyStatus();
    showToast('Saved', 'success'); setStatus('Ready');
  } catch (err) { showToast('Save failed: ' + err.message, 'error'); setStatus('Error'); }
}

function scheduleCheck() {
  clearTimeout(state.checkTimer);
  state.checkTimer = setTimeout(checkFile, 700);
}

async function checkFile() {
  if (!state.activePath) return;
  try {
    const result = await API.check(state.activePath, state.editor.getValue());
    showDiagnostics(result.diagnostics || []);
  } catch (_) {}
}

async function runFile() {
  if (!state.activePath) return;
  if (state.tabs.get(state.activePath)?.dirty) await saveFile();
  setStatus('Running…'); setPanel('output');
  appendOutput('▶ Running ' + state.activePath + '…\n');
  try {
    const result = await API.run(state.activePath, state.editor.getValue());
    appendOutput((result.stdout || '') + (result.stderr || ''));
    showToast(result.ok ? 'Run completed' : 'Run failed', result.ok ? 'success' : 'error');
    setStatus(result.ok ? 'Ready' : 'Error');
  } catch (err) { appendOutput(err.message); showToast('Run failed', 'error'); setStatus('Error'); }
}

async function buildFile() {
  if (!state.activePath) return;
  if (state.tabs.get(state.activePath)?.dirty) await saveFile();
  const release = isReleaseBuild();
  setStatus('Building…'); setPanel('output');
  appendOutput(`▶ Building ${state.activePath}${release ? ' (release)' : ''}…\n`);
  try {
    const result = await API.build(state.activePath, state.editor.getValue(), release);
    appendOutput((result.stdout || '') + (result.stderr || ''));
    showToast(result.ok ? 'Build succeeded' : 'Build failed', result.ok ? 'success' : 'error');
    setStatus(result.ok ? 'Ready' : 'Error');
  } catch (err) { appendOutput(err.message); showToast('Build failed', 'error'); setStatus('Error'); }
}

// ── Diagnostics ──
function showDiagnostics(diags) {
  const list = document.getElementById('problems-list');
  const badge = document.getElementById('problem-count');
  if (!diags.length) {
    list.innerHTML = '<div class="problems-empty">No problems detected</div>';
    badge.textContent = '0'; badge.className = 'badge';
    clearEditorMarkers(); return;
  }
  const errors = diags.filter(d => !d.warning);
  badge.textContent = String(diags.length);
  badge.className = 'badge' + (errors.length ? ' has-errors' : ' has-warnings');
  list.innerHTML = '';
  for (const d of diags) {
    const item = document.createElement('div');
    item.className = 'problem-item';
    item.innerHTML = `<svg class="problem-icon ${d.warning ? 'warning' : 'error'}" viewBox="0 0 16 16"><path fill="currentColor" d="${d.warning ? 'M8 1.5a.75.75 0 0 1 .67.418l6.5 12A.75.75 0 0 1 14.5 15h-13a.75.75 0 0 1-.67-1.082l6.5-12A.75.75 0 0 1 8 1.5zM8 5.75a.75.75 0 0 0-.75.75v3.5a.75.75 0 0 0 1.5 0v-3.5A.75.75 0 0 0 8 5.75zm0 8a1 1 0 1 0 0-2 1 1 0 0 0 0 2z' : 'M2.343 2.343a8 8 0 1 1 11.314 11.314A8 8 0 0 1 2.343 2.343zM8 4a.75.75 0 0 0-.75.75v3.5a.75.75 0 0 0 1.5 0v-3.5A.75.75 0 0 0 8 4zm0 8a1 1 0 1 0 0-2 1 1 0 0 0 0 2z'}"/></svg><div><div class="problem-msg">${escapeHtml(d.message)}</div><div class="problem-loc">${d.stage} · line ${d.line}, col ${d.col}</div>${d.hint ? `<div class="problem-hint">${escapeHtml(d.hint)}</div>` : ''}</div>`;
    item.addEventListener('click', () => { state.editor.revealLineInCenter(d.line); state.editor.setPosition({ lineNumber: d.line, column: d.col }); state.editor.focus(); });
    list.appendChild(item);
  }
  setEditorMarkers(diags);
  if (errors.length) {
    document.getElementById('bottom-panel').classList.remove('collapsed');
    setPanel('problems');
  }
}

function setEditorMarkers(diags) {
  if (!state.editor) return;
  const model = state.editor.getModel();
  if (!model) return;
  state.monaco.editor.setModelMarkers(model, 'datadream', diags.map(d => ({
    severity: d.warning ? state.monaco.MarkerSeverity.Warning : state.monaco.MarkerSeverity.Error,
    startLineNumber: d.line, startColumn: d.col, endLineNumber: d.line, endColumn: d.col + 1,
    message: d.message + (d.hint ? '\n' + d.hint : ''),
  })));
}

function clearEditorMarkers() {
  const model = state.editor?.getModel();
  if (model) state.monaco?.editor.setModelMarkers(model, 'datadream', []);
}

// ── Output ──
function appendOutput(text) {
  const log = document.getElementById('output-log');
  log.textContent += text;
  log.scrollTop = log.scrollHeight;
}
function clearOutput() { document.getElementById('output-log').textContent = ''; }

// ── Panel ──
function setPanel(name) {
  document.querySelectorAll('.panel-tab').forEach(t => t.classList.toggle('active', t.dataset.panel === name));
  document.querySelectorAll('.panel-pane').forEach(p => p.classList.toggle('active', p.id === `panel-${name}`));
}
function togglePanel() { document.getElementById('bottom-panel').classList.toggle('collapsed'); }

// ── Activity views ──
function setActivityView(view) {
  document.querySelectorAll('.activity-btn[data-view]').forEach(b => b.classList.toggle('active', b.dataset.view === view));
  document.querySelectorAll('.sidebar-view').forEach(v => v.classList.toggle('active', v.id === `view-${view}`));
  if (view === 'search') document.getElementById('quick-open-input').focus();
  if (view === 'sdk') loadSDK();
}

// ── SDK panel ──
async function loadSDK() {
  const panel = document.getElementById('sdk-panel');
  try {
    const d = await API.doctor();
    state.sdkReady = d.ready;
    updateSDKStatus(d);
    panel.innerHTML = `
      <div class="sdk-banner ${d.ready ? 'ready' : 'not-ready'}">${d.ready ? '✓ SDK ready — build and run enabled' : '⚠ SDK incomplete — some features may fail'}</div>
      ${sdkRow('Platform', d.platform, true)}
      ${sdkRow('Clang', d.clang, d.clangOk)}
      ${sdkRow('Toolchain match', d.clangFlavorOk ? 'OK' : 'Mismatch', d.clangFlavorOk)}
      ${sdkRow('raylib headers', d.raylibInclude || 'missing', d.raylibIncludeOk)}
      ${sdkRow('raylib library', d.raylibLib || 'missing', d.raylibLibOk)}
      ${sdkRow('raylib version', d.raylibVersion, true)}
    `;
  } catch (err) { panel.innerHTML = `<div class="sdk-loading">Failed: ${escapeHtml(err.message)}</div>`; }
}

function sdkRow(label, value, ok) {
  return `<div class="sdk-row"><span class="sdk-label">${label}</span><span class="sdk-value ${ok ? 'sdk-ok' : 'sdk-fail'}">${escapeHtml(String(value))} ${ok ? '✓' : '✗'}</span></div>`;
}

function updateSDKStatus(d) {
  const el = document.getElementById('status-sdk');
  el.textContent = d?.ready ? 'SDK ✓' : 'SDK ✗';
  el.className = 'status-item ' + (d?.ready ? 'ok' : 'fail');
  el.title = d?.ready ? 'Toolchain ready' : 'Run datadream doctor for details';
}

// ── Command palette & quick open ──
function openCommandPalette() {
  state.paletteMode = 'commands';
  state.paletteIndex = 0;
  const input = document.getElementById('palette-input');
  input.value = '';
  renderPaletteList('');
  showOverlay('overlay-palette');
  input.focus();
}

function openQuickOpen() {
  state.quickOpenIndex = 0;
  const input = document.getElementById('modal-quickopen');
  input.value = '';
  renderQuickOpen('');
  showOverlay('overlay-quickopen');
  input.focus();
}

function renderPaletteList(q) {
  const list = document.getElementById('palette-list');
  const filtered = COMMANDS.filter(c => c.label.toLowerCase().includes(q.toLowerCase()));
  if (!filtered.length) { list.innerHTML = '<div class="palette-empty">No matching commands</div>'; return; }
  list.innerHTML = filtered.map((c, i) =>
    `<div class="palette-item${i === state.paletteIndex ? ' selected' : ''}" data-id="${c.id}"><span class="label">${escapeHtml(c.label)}</span>${c.keys ? `<span class="hint">${c.keys}</span>` : ''}</div>`
  ).join('');
  list.querySelectorAll('.palette-item').forEach(el => {
    el.addEventListener('click', () => { hideOverlay('overlay-palette'); COMMANDS.find(c => c.id === el.dataset.id)?.run(); });
  });
}

async function renderQuickOpen(q) {
  const list = document.getElementById('quickopen-list');
  list.innerHTML = '<div class="palette-empty">Searching…</div>';
  try {
    const { files } = await API.search(q);
    if (!files.length) { list.innerHTML = '<div class="palette-empty">No files found</div>'; return; }
    list.innerHTML = files.map((f, i) =>
      `<div class="search-item${i === state.quickOpenIndex ? ' selected' : ''}" data-path="${escapeHtml(f.path)}"><span class="name">${escapeHtml(f.name)}</span><span class="path">${escapeHtml(f.path)}</span></div>`
    ).join('');
    list.querySelectorAll('.search-item').forEach(el => {
      el.addEventListener('click', () => { hideOverlay('overlay-quickopen'); openFile(el.dataset.path); });
    });
  } catch (_) { list.innerHTML = '<div class="palette-empty">Search failed</div>'; }
}

// ── Modals ──
function showOverlay(id) { document.getElementById(id).classList.remove('hidden'); }
function hideOverlay(id) { document.getElementById(id).classList.add('hidden'); }

function showNewFileModal() {
  document.getElementById('new-file-path').value = state.activePath
    ? dirname(state.activePath) + '/game.dd' : 'my-game/game.dd';
  showOverlay('overlay-newfile');
  document.getElementById('new-file-path').focus();
}

function dirname(p) { const parts = p.split('/'); parts.pop(); return parts.join('/') || ''; }

async function createNewFile() {
  const path = document.getElementById('new-file-path').value.trim();
  if (!path) return;
  try {
    const data = await API.newFile(path);
    hideOverlay('overlay-newfile');
    await loadTree();
    openTab(data.path, data.content);
    showToast('Created ' + path, 'success');
  } catch (err) { showToast(err.message, 'error'); }
}

function showShortcuts() {
  const grid = document.getElementById('shortcuts-grid');
  grid.innerHTML = SHORTCUTS.map(([label, keys]) =>
    `<div class="shortcut-row"><span>${label}</span><span class="shortcut-keys">${keys.split('+').map(k => `<kbd>${k}</kbd>`).join('+')}</span></div>`
  ).join('');
  showOverlay('overlay-shortcuts');
}

// ── Resizers ──
function initResizers() {
  initDragResizer('sidebar-resizer', (dx) => {
    const sidebar = document.getElementById('sidebar');
    const w = Math.max(200, Math.min(480, sidebar.offsetWidth + dx));
    document.documentElement.style.setProperty('--sidebar-w', w + 'px');
  }, 'col');

  initDragResizer('panel-resizer', (dy) => {
    const panel = document.getElementById('bottom-panel');
    if (panel.classList.contains('collapsed')) return;
    const h = Math.max(100, Math.min(480, panel.offsetHeight - dy));
    document.documentElement.style.setProperty('--panel-h', h + 'px');
  }, 'row');
}

function initDragResizer(id, onDrag, axis) {
  const el = document.getElementById(id);
  el.addEventListener('mousedown', e => {
    e.preventDefault();
    let last = axis === 'col' ? e.clientX : e.clientY;
    el.classList.add('dragging');
    const move = ev => {
      const current = axis === 'col' ? ev.clientX : ev.clientY;
      onDrag(current - last);
      last = current;
    };
    const up = () => { el.classList.remove('dragging'); document.removeEventListener('mousemove', move); document.removeEventListener('mouseup', up); };
    document.addEventListener('mousemove', move);
    document.addEventListener('mouseup', up);
  });
}

// ── Menubar ──
function closeAllMenus() {
  document.querySelectorAll('.menu-root.open').forEach(m => m.classList.remove('open'));
}

function initMenubar() {
  document.querySelectorAll('.menu-root').forEach(root => {
    const trigger = root.querySelector('.menu-trigger');
    trigger.addEventListener('click', e => {
      e.stopPropagation();
      const wasOpen = root.classList.contains('open');
      closeAllMenus();
      if (!wasOpen) root.classList.add('open');
    });
  });

  document.querySelectorAll('.menu-entry[data-cmd]').forEach(entry => {
    entry.addEventListener('click', () => {
      closeAllMenus();
      runMenuCommand(entry.dataset.cmd);
    });
  });

  document.addEventListener('click', closeAllMenus);
  document.getElementById('menubar')?.addEventListener('click', e => e.stopPropagation());
}

function runMenuCommand(cmd) {
  const map = {
    new: showNewFileModal,
    openProject: openProjectFolder,
    open: openQuickOpen,
    save: saveFile,
    close: () => state.activePath && requestCloseTab(state.activePath),
    palette: openCommandPalette,
    check: checkFile,
    run: runFile,
    build: buildFile,
    explorer: () => setActivityView('explorer'),
    search: openQuickOpen,
    sdk: () => setActivityView('sdk'),
    panel: togglePanel,
    problems: () => { document.getElementById('bottom-panel').classList.remove('collapsed'); setPanel('problems'); },
    output: () => { document.getElementById('bottom-panel').classList.remove('collapsed'); setPanel('output'); },
    shortcuts: showShortcuts,
  };
  map[cmd]?.();
}

function isReleaseBuild() {
  return document.getElementById('chk-release-menu')?.checked ?? false;
}

// ── Helpers ──
function setStatus(msg) {
  const el = document.getElementById('status-ready');
  el.textContent = msg;
  el.className = 'status-item';
  if (/ing…|Running|Building|Saving|Opening/i.test(msg)) el.classList.add('busy');
  else if (/Error|failed/i.test(msg)) el.classList.add('error');
}

function showToast(msg, type = '') {
  document.querySelector('.toast')?.remove();
  const t = document.createElement('div');
  t.className = 'toast' + (type ? ' ' + type : '');
  t.textContent = msg;
  document.body.appendChild(t);
  setTimeout(() => t.remove(), 2800);
}

function escapeHtml(s) {
  return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

// ── Init ──
async function init() {
  document.querySelectorAll('.panel-tab').forEach(t => t.addEventListener('click', () => setPanel(t.dataset.panel)));
  document.getElementById('btn-save').addEventListener('click', saveFile);
  document.getElementById('btn-check').addEventListener('click', checkFile);
  document.getElementById('btn-run').addEventListener('click', runFile);
  document.getElementById('btn-build').addEventListener('click', buildFile);
  document.getElementById('btn-refresh-tree').addEventListener('click', loadTree);
  document.getElementById('btn-new-file').addEventListener('click', showNewFileModal);
  document.getElementById('btn-command').addEventListener('click', openCommandPalette);
  document.getElementById('btn-clear-output').addEventListener('click', clearOutput);
  document.getElementById('btn-toggle-panel').addEventListener('click', togglePanel);
  document.getElementById('btn-shortcuts').addEventListener('click', showShortcuts);
  document.getElementById('btn-new-create').addEventListener('click', createNewFile);
  document.getElementById('btn-new-cancel').addEventListener('click', () => hideOverlay('overlay-newfile'));
  document.getElementById('btn-shortcuts-close').addEventListener('click', () => hideOverlay('overlay-shortcuts'));
  document.getElementById('tree-filter').addEventListener('input', filterTree);

  document.querySelectorAll('.activity-btn[data-view]').forEach(b =>
    b.addEventListener('click', () => setActivityView(b.dataset.view)));
  document.querySelectorAll('.quick-link').forEach(b =>
    b.addEventListener('click', () => openFile(b.dataset.path)));
  document.getElementById('status-sdk').addEventListener('click', () => setActivityView('sdk'));
  document.getElementById('btn-run-header')?.addEventListener('click', runFile);

  initMenubar();

  // Unsaved dialog
  document.getElementById('btn-unsaved-save').addEventListener('click', async () => {
    hideOverlay('overlay-unsaved');
    await saveFile();
    if (state.pendingClose) { closeTab(state.pendingClose); state.pendingClose = null; }
  });
  document.getElementById('btn-unsaved-discard').addEventListener('click', () => {
    hideOverlay('overlay-unsaved');
    if (state.pendingClose) { closeTab(state.pendingClose); state.pendingClose = null; }
  });
  document.getElementById('btn-unsaved-cancel').addEventListener('click', () => { hideOverlay('overlay-unsaved'); state.pendingClose = null; });

  // Palette keyboard
  document.getElementById('palette-input').addEventListener('input', e => { state.paletteIndex = 0; renderPaletteList(e.target.value); });
  document.getElementById('palette-input').addEventListener('keydown', e => {
    const items = [...document.querySelectorAll('#palette-list .palette-item')];
    if (e.key === 'Escape') hideOverlay('overlay-palette');
    if (e.key === 'ArrowDown') { e.preventDefault(); state.paletteIndex = Math.min(state.paletteIndex + 1, items.length - 1); renderPaletteList(e.target.value); }
    if (e.key === 'ArrowUp') { e.preventDefault(); state.paletteIndex = Math.max(state.paletteIndex - 1, 0); renderPaletteList(e.target.value); }
    if (e.key === 'Enter' && items[state.paletteIndex]) { hideOverlay('overlay-palette'); COMMANDS.find(c => c.id === items[state.paletteIndex].dataset.id)?.run(); }
  });

  // Quick open
  let searchTimer;
  document.getElementById('modal-quickopen').addEventListener('input', e => {
    clearTimeout(searchTimer);
    searchTimer = setTimeout(() => renderQuickOpen(e.target.value), 150);
  });
  document.getElementById('modal-quickopen').addEventListener('keydown', e => {
    if (e.key === 'Escape') hideOverlay('overlay-quickopen');
    if (e.key === 'Enter') document.querySelector('#quickopen-list .search-item.selected, #quickopen-list .search-item')?.click();
  });
  document.getElementById('quick-open-input').addEventListener('input', e => renderQuickOpen(e.target.value));

  // Global shortcuts
  document.addEventListener('keydown', e => {
    if (e.key === '?' && !e.ctrlKey && !e.metaKey && document.activeElement.tagName !== 'INPUT') { e.preventDefault(); showShortcuts(); }
    if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === 'P') { e.preventDefault(); openCommandPalette(); }
    if ((e.ctrlKey || e.metaKey) && e.key === 'p' && !e.shiftKey) { e.preventDefault(); openQuickOpen(); }
    if (e.key === 'Escape') document.querySelectorAll('.overlay:not(.hidden)').forEach(o => o.classList.add('hidden'));
  });

  // Close overlays on backdrop click
  document.querySelectorAll('.overlay').forEach(o => o.addEventListener('click', e => { if (e.target === o) o.classList.add('hidden'); }));

  initResizers();

  await waitForDesktopBackend();
  refreshAPI();

  try {
    const info = await API.version();
    applyProjectInfo(info);
  } catch (_) {}

  try {
    const d = await API.doctor();
    updateSDKStatus(d);
  } catch (_) {}

  await loadTree();

  try {
    const monaco = await loadMonaco();
    initEditor(monaco);
    if (API.isDesktop && !state.activePath) {
      openFile('examples/beginner/clicker.dd').catch(() => {});
    }
  } catch (err) {
    showToast('Editor failed to load: ' + err.message, 'error');
  }
}

init();
