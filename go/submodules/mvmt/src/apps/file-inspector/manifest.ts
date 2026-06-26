import { AppManifest } from '../registry.js';

// First-party reference app. Exercises the cookie-authenticated FS API
// (/api/fs/sources, /api/fs/list, /api/fs/stat) end-to-end. Inline HTML
// payload matches DASHBOARD_PAGE_HTML pattern so no build-system changes
// are required to ship the app. Future apps shipping non-trivial bundles
// can move to a real disk-served static directory; not necessary yet.
const FILE_INSPECTOR_HTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8" />
<title>File Inspector - mvmt</title>
<meta name="viewport" content="width=device-width, initial-scale=1" />
<style>
:root{--text:#1d1d1f;--muted:#6e6e73;--border:#d2d2d7;--bg:#f5f5f7;--card:#fff}
*{box-sizing:border-box}
body{margin:0;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;color:var(--text);background:var(--bg);padding:1.5rem}
.bar{display:flex;align-items:baseline;justify-content:space-between;margin-bottom:1rem;gap:1rem;flex-wrap:wrap}
h1{font-size:1.25rem;margin:0}
a.back{color:#0a84ff;text-decoration:none;font-size:.9rem}
a.back:hover{text-decoration:underline}
.crumbs{font-size:.85rem;color:var(--muted);margin-bottom:.75rem}
.crumbs a{color:#0a84ff;text-decoration:none}
.crumbs a:hover{text-decoration:underline}
.grid{display:grid;grid-template-columns:1fr 1fr;gap:1rem}
@media(max-width:700px){.grid{grid-template-columns:1fr}}
.card{background:var(--card);border:1px solid var(--border);border-radius:8px;padding:.75rem 1rem}
.card h2{margin:0 0 .5rem;font-size:.95rem;font-weight:600}
.list{list-style:none;margin:0;padding:0}
.list li{padding:.4rem .5rem;border-radius:5px;cursor:pointer;display:flex;justify-content:space-between;gap:.5rem;font-size:.85rem}
.list li:hover{background:#0a84ff;color:#fff}
.list li.dir{font-weight:500}
.list li .meta{color:var(--muted);font-size:.75rem;flex:none}
.list li:hover .meta{color:#fff}
.stat{font-size:.8rem;color:var(--text);white-space:pre-wrap;font-family:ui-monospace,monospace;background:#f0f0f3;padding:.6rem;border-radius:5px;margin:0;overflow-x:auto}
.stat-actions{margin-top:.6rem}
.stat-actions button,.load-more{appearance:none;border:1px solid var(--border);background:#fff;border-radius:6px;padding:.42rem .65rem;font:inherit;font-size:.82rem;cursor:pointer;color:#0a84ff}
.stat-actions button:hover,.load-more:hover{background:#f5f9ff}
.load-more{width:100%;margin-top:.45rem;color:var(--text)}
.muted{color:var(--muted);font-size:.85rem}
.err{color:#c0392b;font-size:.85rem}
</style>
</head>
<body>
<div class="bar">
<h1>File Inspector</h1>
<a class="back" id="back">&larr; Dashboard</a>
</div>
<div class="crumbs" id="crumbs"></div>
<div class="grid">
<div class="card">
<h2 id="entries-title">Sources</h2>
<div id="entries"><div class="muted">Loading...</div></div>
</div>
<div class="card">
<h2>Selected</h2>
<div id="stat"><div class="muted">Click an entry to inspect.</div></div>
</div>
</div>
<script>
(function () {
  var ENTRY_BATCH_SIZE = 200;
  function appBase() {
    var p = location.pathname.replace(/\\/+$/, '');
    var marker = '/apps/file-inspector';
    var i = p.indexOf(marker);
    return i >= 0 ? p.slice(0, i) : '';
  }
  function api(path) {
    return fetch(appBase() + path, { credentials: 'same-origin', headers: { Accept: 'application/json' } })
      .then(function (r) { return r.ok ? r.json() : r.json().then(function (b) { throw new Error(b.error || ('HTTP ' + r.status)); }, function () { throw new Error('HTTP ' + r.status); }); });
  }
  var state = { path: '/', entries: [], visibleEntryCount: ENTRY_BATCH_SIZE };
  function $(id) { return document.getElementById(id); }
  function setError(node, msg) { node.innerHTML = ''; var d = document.createElement('div'); d.className = 'err'; d.textContent = msg; node.appendChild(d); }
  function fmtSize(n) { if (!n) return ''; if (n < 1024) return n + ' B'; if (n < 1024*1024) return (n/1024).toFixed(1) + ' KB'; return (n/1024/1024).toFixed(1) + ' MB'; }
  function renderStat(stat, stale) {
    var node = $('stat');
    var pre = document.createElement('pre'); pre.className = 'stat';
    pre.textContent = JSON.stringify(stat, null, 2);
    node.innerHTML = '';
    node.appendChild(pre);
    if (stale && stat.path) {
      var actions = document.createElement('div');
      actions.className = 'stat-actions';
      var btn = document.createElement('button');
      btn.type = 'button';
      btn.textContent = 'Refresh stat';
      btn.addEventListener('click', function () { inspect(stat.path); });
      actions.appendChild(btn);
      node.appendChild(actions);
    }
  }
  function renderCrumbs() {
    var c = $('crumbs'); c.innerHTML = '';
    var segs = state.path === '/' ? [] : state.path.split('/').filter(Boolean);
    var rootLink = document.createElement('a'); rootLink.href = '#'; rootLink.textContent = 'sources';
    rootLink.addEventListener('click', function (e) { e.preventDefault(); navigate('/'); });
    c.appendChild(rootLink);
    var acc = '';
    for (var i = 0; i < segs.length; i += 1) {
      c.appendChild(document.createTextNode(' / '));
      acc += '/' + segs[i];
      var captured = acc;
      var a = document.createElement('a'); a.href = '#'; a.textContent = segs[i];
      a.addEventListener('click', function (cap) { return function (e) { e.preventDefault(); navigate(cap); }; }(captured));
      c.appendChild(a);
    }
  }
  function renderEntries(entries) {
    var node = $('entries'); node.innerHTML = '';
    var visible = entries.slice(0, state.visibleEntryCount);
    if (!visible.length) {
      var d = document.createElement('div'); d.className = 'muted'; d.textContent = 'Empty.';
      node.appendChild(d); return;
    }
    var ul = document.createElement('ul'); ul.className = 'list';
    visible.forEach(function (entry) {
      var li = document.createElement('li');
      if (entry.type === 'directory') li.className = 'dir';
      var name = document.createElement('span'); name.textContent = entry.name || entry.path;
      var meta = document.createElement('span'); meta.className = 'meta';
      meta.textContent = entry.type === 'directory' ? 'dir' : fmtSize(entry.size);
      li.appendChild(name); li.appendChild(meta);
      li.addEventListener('click', function () {
        if (entry.unavailable) {
          setError($('stat'), entry.path + ' is unavailable on disk.');
          return;
        }
        if (entry.type === 'directory') navigate(entry.path);
        else renderStat(entry, true);
      });
      ul.appendChild(li);
    });
    node.appendChild(ul);
    if (visible.length < entries.length) {
      var more = document.createElement('button');
      more.type = 'button';
      more.className = 'load-more';
      more.textContent = 'Show more entries (' + visible.length + ' of ' + entries.length + ')';
      more.addEventListener('click', function () {
        state.visibleEntryCount += ENTRY_BATCH_SIZE;
        renderEntries(state.entries);
      });
      node.appendChild(more);
    }
  }
  function inspect(p) {
    var node = $('stat');
    node.innerHTML = '<div class="muted">Loading...</div>';
    api('/api/fs/stat?path=' + encodeURIComponent(p))
      .then(function (s) { renderStat(s, false); })
      .catch(function (e) { setError(node, e.message); });
  }
  function navigate(p) {
    state.path = p;
    state.visibleEntryCount = ENTRY_BATCH_SIZE;
    renderCrumbs();
    $('entries-title').textContent = p === '/' ? 'Sources' : p;
    $('entries').innerHTML = '<div class="muted">Loading...</div>';
    $('stat').innerHTML = '<div class="muted">Click an entry to inspect.</div>';
    var req = p === '/' ? api('/api/fs/sources').then(function (b) { return b.sources || []; })
                        : api('/api/fs/list?path=' + encodeURIComponent(p)).then(function (b) { return b.entries || []; });
    req.then(function (entries) {
      state.entries = entries || [];
      renderEntries(state.entries);
    }).catch(function (e) { setError($('entries'), e.message); });
  }
  $('back').setAttribute('href', appBase() + '/dashboard');
  navigate('/');
})();
</script>
</body>
</html>
`;

export const fileInspectorApp: AppManifest = {
  id: 'file-inspector',
  label: 'File Inspector',
  description: 'Browse mounted sources, list directories, and inspect file metadata through the cookie-authenticated FS API.',
  html: FILE_INSPECTOR_HTML,
};
