import { AppManifest } from '../registry.js';

// First real first-party app. It intentionally consumes only the shared
// cookie-authenticated FS API; there is no photos-specific backend.
const PHOTOS_HTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8" />
<title>Photos - mvmt</title>
<meta name="viewport" content="width=device-width, initial-scale=1" />
<style>
:root{--text:#1d1d1f;--muted:#6e6e73;--border:#d2d2d7;--bg:#f5f5f7;--card:#fff;--accent:#0a84ff}
*{box-sizing:border-box}
body{margin:0;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;color:var(--text);background:var(--bg)}
.shell{min-height:100vh;display:flex;flex-direction:column}
.topbar{height:58px;display:flex;align-items:center;justify-content:space-between;gap:1rem;padding:0 1.25rem;background:rgba(255,255,255,.86);border-bottom:1px solid var(--border);position:sticky;top:0;z-index:10;backdrop-filter:saturate(180%) blur(16px)}
.title{display:flex;align-items:center;gap:.7rem;min-width:0}
.title h1{font-size:1.05rem;margin:0;font-weight:650}
.crumbs{display:flex;align-items:center;gap:.35rem;color:var(--muted);font-size:.86rem;min-width:0;overflow:hidden;white-space:nowrap}
.crumbs a,.back{color:var(--accent);text-decoration:none}
.crumbs a:hover,.back:hover{text-decoration:underline}
.main{display:grid;grid-template-columns:220px minmax(0,1fr);gap:0;min-height:calc(100vh - 58px)}
.sidebar{border-right:1px solid var(--border);background:#fbfbfd;padding:1rem}
.sidebar h2{font-size:.75rem;color:var(--muted);text-transform:uppercase;letter-spacing:.08em;margin:.25rem .45rem .6rem}
.source-list{display:flex;flex-direction:column;gap:.25rem}
.source-btn{appearance:none;border:0;background:transparent;color:var(--text);text-align:left;border-radius:7px;padding:.48rem .55rem;font:inherit;font-size:.9rem;cursor:pointer;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.source-btn:hover{background:#ececf0}
.source-btn.active{background:#e6f1ff;color:#005dbb;font-weight:600}
.content{padding:1rem 1.2rem 2rem}
.status{color:var(--muted);font-size:.9rem;margin:.5rem 0 1rem}
.err{color:#c0392b}
.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(154px,1fr));gap:14px;align-items:start}
.tile{border:0;background:transparent;padding:0;text-align:left;color:inherit;cursor:pointer;min-width:0}
.album,.photo{background:var(--card);border:1px solid var(--border);border-radius:10px;overflow:hidden;box-shadow:0 1px 2px rgba(0,0,0,.04)}
.album{height:154px;display:flex;align-items:center;justify-content:center;color:var(--muted);font-weight:600}
.album::before{content:'';display:block;width:62px;height:46px;border:2px solid #9aa8ba;border-radius:5px;background:linear-gradient(#fff,#edf2f8);box-shadow:0 -7px 0 -2px #c8d4e4}
.photo{aspect-ratio:1/1;background:#ececf0}
.photo img{display:block;width:100%;height:100%;object-fit:cover}
.caption{font-size:.82rem;line-height:1.2;margin:.45rem .1rem 0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.load-more{height:154px;border:1px dashed var(--border);border-radius:10px;background:rgba(255,255,255,.75);color:var(--accent);font-weight:600;display:flex;align-items:center;justify-content:center}
.empty{border:1px dashed var(--border);border-radius:10px;padding:2rem;color:var(--muted);background:rgba(255,255,255,.72);text-align:center}
.viewer{position:fixed;inset:0;background:rgba(0,0,0,.86);display:flex;flex-direction:column;z-index:30}
.viewer.hidden{display:none}
.viewer-bar{height:56px;display:flex;align-items:center;justify-content:space-between;gap:1rem;padding:0 1rem;color:#fff}
.viewer-title{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:.95rem}
.viewer-actions{display:flex;align-items:center;gap:.6rem;flex:none}
.viewer-actions a,.viewer-actions button{border:1px solid rgba(255,255,255,.28);background:rgba(255,255,255,.1);color:#fff;border-radius:7px;padding:.42rem .7rem;text-decoration:none;font:inherit;cursor:pointer}
.viewer-stage{flex:1;min-height:0;display:flex;align-items:center;justify-content:center;padding:0 1rem 1rem}
.viewer-stage img{max-width:100%;max-height:100%;object-fit:contain;box-shadow:0 8px 40px rgba(0,0,0,.35)}
@media(max-width:720px){.main{grid-template-columns:1fr}.sidebar{border-right:0;border-bottom:1px solid var(--border)}.source-list{flex-direction:row;overflow-x:auto}.source-btn{flex:none}.grid{grid-template-columns:repeat(auto-fill,minmax(118px,1fr));gap:10px}.album{height:118px}.content{padding:.8rem}.topbar{align-items:flex-start;height:auto;min-height:58px;padding:.75rem .9rem;flex-direction:column}.back{align-self:flex-end}}
</style>
</head>
<body>
<div class="shell">
  <header class="topbar">
    <div class="title">
      <h1>Photos</h1>
      <nav class="crumbs" id="crumbs"></nav>
    </div>
    <a class="back" id="back">Dashboard</a>
  </header>
  <main class="main">
    <aside class="sidebar">
      <h2>Sources</h2>
      <div class="source-list" id="sources"><div class="status">Loading...</div></div>
    </aside>
    <section class="content">
      <div class="status" id="status">Loading...</div>
      <div class="grid" id="grid"></div>
    </section>
  </main>
</div>
<div class="viewer hidden" id="viewer" role="dialog" aria-modal="true">
  <div class="viewer-bar">
    <div class="viewer-title" id="viewer-title"></div>
    <div class="viewer-actions">
      <a id="viewer-download" download>Download</a>
      <button type="button" id="viewer-close">Close</button>
    </div>
  </div>
  <div class="viewer-stage" id="viewer-stage"></div>
</div>
<script>
(function () {
  var PHOTO_BATCH_SIZE = 48;
  var IMAGE_EXTENSIONS = { '.avif': true, '.gif': true, '.heic': true, '.heif': true, '.jpg': true, '.jpeg': true, '.png': true, '.webp': true };
  var state = { path: '/', sources: [], entries: [], visiblePhotoCount: PHOTO_BATCH_SIZE };
  function appBase() {
    var p = location.pathname.replace(/\\/+$/, '');
    var marker = '/apps/photos';
    var i = p.indexOf(marker);
    return i >= 0 ? p.slice(0, i) : '';
  }
  function api(path) {
    return fetch(appBase() + path, { credentials: 'same-origin', headers: { Accept: 'application/json' } })
      .then(function (r) { return r.ok ? r.json() : r.json().then(function (b) { throw new Error(b.error || ('HTTP ' + r.status)); }, function () { throw new Error('HTTP ' + r.status); }); });
  }
  function $(id) { return document.getElementById(id); }
  function fileUrl(path) { return appBase() + '/api/fs/file?path=' + encodeURIComponent(path); }
  function ext(path) {
    var name = (path || '').toLowerCase();
    var dot = name.lastIndexOf('.');
    return dot >= 0 ? name.slice(dot) : '';
  }
  function isImage(entry) { return entry.type === 'file' && !!IMAGE_EXTENSIONS[ext(entry.path || entry.name)]; }
  function setStatus(message, kind) {
    var node = $('status');
    node.textContent = message || '';
    node.className = kind === 'error' ? 'status err' : 'status';
  }
  function renderSources() {
    var node = $('sources');
    node.innerHTML = '';
    if (!state.sources.length) {
      var empty = document.createElement('div');
      empty.className = 'status';
      empty.textContent = 'No sources.';
      node.appendChild(empty);
      return;
    }
    state.sources.forEach(function (source) {
      var btn = document.createElement('button');
      btn.type = 'button';
      btn.className = source.path === state.path ? 'source-btn active' : 'source-btn';
      btn.textContent = source.name || source.path;
      btn.title = source.path;
      btn.addEventListener('click', function () { navigate(source.path); });
      node.appendChild(btn);
    });
  }
  function renderCrumbs() {
    var node = $('crumbs');
    node.innerHTML = '';
    var segs = state.path === '/' ? [] : state.path.split('/').filter(Boolean);
    if (segs.length === 0) {
      node.textContent = 'sources';
      return;
    }
    var root = document.createElement('a');
    root.href = '#';
    root.textContent = 'sources';
    root.addEventListener('click', function (event) { event.preventDefault(); navigate('/'); });
    node.appendChild(root);
    var acc = '';
    segs.forEach(function (seg) {
      node.appendChild(document.createTextNode(' / '));
      acc += '/' + seg;
      var target = acc;
      var link = document.createElement('a');
      link.href = '#';
      link.textContent = seg;
      link.addEventListener('click', function (captured) {
        return function (event) { event.preventDefault(); navigate(captured); };
      }(target));
      node.appendChild(link);
    });
  }
  function renderEntries(entries) {
    var grid = $('grid');
    grid.innerHTML = '';
    var directories = entries.filter(function (entry) { return entry.type === 'directory' && !entry.unavailable; });
    var photos = entries.filter(isImage);
    var visiblePhotos = photos.slice(0, state.visiblePhotoCount);
    var visible = directories.concat(visiblePhotos);
    if (!visible.length) {
      var empty = document.createElement('div');
      empty.className = 'empty';
      empty.textContent = state.path === '/' ? 'No photo sources yet.' : 'No photos in this folder.';
      grid.appendChild(empty);
      setStatus('');
      return;
    }
    setStatus(visiblePhotos.length + ' of ' + photos.length + ' photos, ' + directories.length + ' folders');
    visible.forEach(function (entry) {
      var button = document.createElement('button');
      button.type = 'button';
      button.className = 'tile';
      var frame = document.createElement('div');
      frame.className = entry.type === 'directory' ? 'album' : 'photo';
      if (entry.type === 'file') {
        var img = document.createElement('img');
        img.loading = 'lazy';
        img.decoding = 'async';
        img.alt = entry.name || entry.path;
        img.src = fileUrl(entry.path);
        frame.appendChild(img);
      }
      var caption = document.createElement('div');
      caption.className = 'caption';
      caption.textContent = entry.name || entry.path;
      button.appendChild(frame);
      button.appendChild(caption);
      button.addEventListener('click', function () {
        if (entry.type === 'directory') navigate(entry.path);
        else openViewer(entry);
      });
      grid.appendChild(button);
    });
    if (visiblePhotos.length < photos.length) {
      var more = document.createElement('button');
      more.type = 'button';
      more.className = 'tile load-more';
      more.textContent = 'Show more photos';
      more.addEventListener('click', function () {
        state.visiblePhotoCount += PHOTO_BATCH_SIZE;
        renderEntries(state.entries);
      });
      grid.appendChild(more);
    }
  }
  function openViewer(entry) {
    var viewer = $('viewer');
    var stage = $('viewer-stage');
    stage.innerHTML = '';
    var img = document.createElement('img');
    img.alt = entry.name || entry.path;
    img.src = fileUrl(entry.path);
    stage.appendChild(img);
    $('viewer-title').textContent = entry.name || entry.path;
    $('viewer-download').setAttribute('href', fileUrl(entry.path));
    $('viewer-download').setAttribute('download', entry.name || 'photo');
    viewer.classList.remove('hidden');
  }
  function closeViewer() {
    $('viewer').classList.add('hidden');
    $('viewer-stage').innerHTML = '';
  }
  function navigate(path) {
    state.path = path;
    state.visiblePhotoCount = PHOTO_BATCH_SIZE;
    renderSources();
    renderCrumbs();
    $('grid').innerHTML = '';
    setStatus('Loading...');
    var req = path === '/' ? api('/api/fs/sources').then(function (body) { state.sources = body.sources || []; return state.sources; })
                           : api('/api/fs/list?path=' + encodeURIComponent(path)).then(function (body) { return body.entries || []; });
    req.then(function (entries) {
      if (path === '/') renderSources();
      state.entries = entries || [];
      renderEntries(state.entries);
    }).catch(function (error) {
      setStatus(error.message || 'Unable to load photos.', 'error');
    });
  }
  $('back').setAttribute('href', appBase() + '/dashboard');
  $('viewer-close').addEventListener('click', closeViewer);
  $('viewer').addEventListener('click', function (event) { if (event.target === $('viewer')) closeViewer(); });
  document.addEventListener('keydown', function (event) { if (event.key === 'Escape') closeViewer(); });
  navigate('/');
})();
</script>
</body>
</html>
`;

export const photosApp: AppManifest = {
  id: 'photos',
  label: 'Photos',
  description: 'Browse image files from mounted sources using the shared filesystem API.',
  html: PHOTOS_HTML,
};
