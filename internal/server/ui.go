package server

import "net/http"

const uiHTML = `<!DOCTYPE html><html lang="en"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Chalkboard — Stockyard</title>
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:wght@400;700&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
<style>:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#c45d2c;--rust-light:#e8753a;--rust-dark:#8b3d1a;--leather:#a0845c;--cream:#f0e6d3;--cream-dim:#bfb5a3;--cream-muted:#7a7060;--green:#5ba86e;--font-serif:'Libre Baskerville',Georgia,serif;--font-mono:'JetBrains Mono',monospace}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--font-serif);min-height:100vh}a{color:var(--rust-light);text-decoration:none}
.hdr{background:var(--bg2);border-bottom:2px solid var(--rust-dark);padding:.9rem 1.8rem;display:flex;align-items:center;justify-content:space-between}.hdr-left{display:flex;align-items:center;gap:1rem}.hdr-brand{font-family:var(--font-mono);font-size:.75rem;color:var(--leather);letter-spacing:3px;text-transform:uppercase}.hdr-title{font-family:var(--font-mono);font-size:1.1rem;color:var(--cream)}.badge{font-family:var(--font-mono);font-size:.6rem;padding:.2rem .6rem;border:1px solid var(--green);color:var(--green);letter-spacing:1px;text-transform:uppercase}
.main{max-width:1000px;margin:0 auto;padding:2rem 1.5rem}.cards{display:grid;grid-template-columns:repeat(auto-fit,minmax(140px,1fr));gap:1rem;margin-bottom:2rem}.card{background:var(--bg2);border:1px solid var(--bg3);padding:1rem 1.2rem}.card-val{font-family:var(--font-mono);font-size:1.6rem;font-weight:700;display:block}.card-lbl{font-family:var(--font-mono);font-size:.58rem;letter-spacing:2px;text-transform:uppercase;color:var(--leather);margin-top:.2rem}
.section{margin-bottom:2rem}.section-title{font-family:var(--font-mono);font-size:.68rem;letter-spacing:3px;text-transform:uppercase;color:var(--rust-light);margin-bottom:.8rem;padding-bottom:.5rem;border-bottom:1px solid var(--bg3)}table{width:100%;border-collapse:collapse;font-family:var(--font-mono);font-size:.75rem}th{background:var(--bg3);padding:.4rem .8rem;text-align:left;color:#c4a87a;font-weight:400;font-size:.62rem;letter-spacing:1px;text-transform:uppercase}td{padding:.4rem .8rem;border-bottom:1px solid var(--bg3);color:var(--cream-dim)}tr:hover td{background:var(--bg2)}.empty{color:var(--cream-muted);text-align:center;padding:2rem;font-style:italic}
.btn{font-family:var(--font-mono);font-size:.7rem;padding:.3rem .8rem;border:1px solid var(--leather);background:transparent;color:var(--cream);cursor:pointer}.btn:hover{border-color:var(--rust-light);color:var(--rust-light)}.btn-rust{border-color:var(--rust);color:var(--rust-light)}.btn-rust:hover{background:var(--rust);color:var(--cream)}.btn-sm{font-size:.62rem;padding:.2rem .5rem}
.lbl{font-family:var(--font-mono);font-size:.62rem;letter-spacing:1px;text-transform:uppercase;color:var(--leather)}input,textarea{font-family:var(--font-mono);font-size:.78rem;background:var(--bg3);border:1px solid var(--bg3);color:var(--cream);padding:.4rem .7rem;outline:none;width:100%}input:focus,textarea:focus{border-color:var(--leather)}.row{display:flex;gap:.8rem;align-items:flex-end;flex-wrap:wrap;margin-bottom:1rem}.field{display:flex;flex-direction:column;gap:.3rem}
.tabs{display:flex;gap:0;margin-bottom:1.5rem;border-bottom:1px solid var(--bg3)}.tab{font-family:var(--font-mono);font-size:.72rem;padding:.6rem 1.2rem;color:var(--cream-muted);cursor:pointer;border-bottom:2px solid transparent;letter-spacing:1px;text-transform:uppercase}.tab:hover{color:var(--cream-dim)}.tab.active{color:var(--rust-light);border-bottom-color:var(--rust-light)}.tab-content{display:none}.tab-content.active{display:block}
</style></head><body>
<div class="hdr"><div class="hdr-left">
<svg viewBox="0 0 64 64" width="22" height="22" fill="none"><rect x="8" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="28" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="48" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="8" y="27" width="48" height="7" rx="2.5" fill="#c4a87a"/></svg>
<span class="hdr-brand">Stockyard</span><span class="hdr-title">Chalkboard</span></div>
<div style="display:flex;gap:.8rem;align-items:center"><span class="badge">Free</span><a href="/wiki/" class="lbl" style="color:var(--leather)">Wiki</a></div></div>
<div class="main">
<div class="cards">
  <div class="card"><span class="card-val" id="s-pages">—</span><span class="card-lbl">Pages</span></div>
  <div class="card"><span class="card-val" id="s-versions">—</span><span class="card-lbl">Versions</span></div>
</div>
<div class="tabs">
  <div class="tab active" onclick="switchTab('pages')">Pages</div>
  <div class="tab" onclick="switchTab('create')">Create</div>
</div>
<div id="tab-pages" class="tab-content active">
  <div class="section"><div class="section-title">Pages</div>
  <table><thead><tr><th>Title</th><th>Slug</th><th>Updated</th><th></th></tr></thead>
  <tbody id="pages-body"></tbody></table></div>
</div>
<div id="tab-create" class="tab-content">
  <div class="section"><div class="section-title">New Page</div>
  <div class="row"><div class="field" style="flex:2"><span class="lbl">Title</span><input id="c-title" placeholder="Getting Started"></div><div class="field" style="flex:1"><span class="lbl">Slug</span><input id="c-slug" placeholder="auto"></div></div>
  <div class="field" style="margin-bottom:.8rem"><span class="lbl">Content</span><textarea id="c-content" rows="10" placeholder="Write your page..."></textarea></div>
  <button class="btn btn-rust" onclick="createPage()">Create</button>
  <div id="c-result" style="margin-top:.5rem"></div></div>
</div>
</div>
<script>
function switchTab(n){document.querySelectorAll('.tab').forEach(t=>t.classList.toggle('active',t.textContent.toLowerCase()===n));document.querySelectorAll('.tab-content').forEach(t=>t.classList.toggle('active',t.id==='tab-'+n));}
async function refresh(){
  try{const s=await(await fetch('/api/status')).json();document.getElementById('s-pages').textContent=s.pages||0;document.getElementById('s-versions').textContent=s.versions||0;}catch(e){}
  try{const d=await(await fetch('/api/pages')).json();const ps=d.pages||[];const tb=document.getElementById('pages-body');
  if(!ps.length){tb.innerHTML='<tr><td colspan="4" class="empty">No pages yet.</td></tr>';return;}
  tb.innerHTML=ps.map(p=>'<tr><td style="color:var(--cream);font-weight:600"><a href="/wiki/'+esc(p.slug)+'">'+esc(p.title)+'</a></td><td style="font-size:.65rem">/wiki/'+esc(p.slug)+'</td><td style="font-size:.62rem;color:var(--cream-muted)">'+timeAgo(p.updated_at)+'</td><td><button class="btn btn-sm" onclick="del(\''+p.id+'\')">Del</button></td></tr>').join('');}catch(e){}
}
async function createPage(){const title=document.getElementById('c-title').value.trim();if(!title)return;const r=await fetch('/api/pages',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({title,slug:document.getElementById('c-slug').value.trim(),content:document.getElementById('c-content').value})});const d=await r.json();if(r.ok){document.getElementById('c-result').innerHTML='<span style="color:var(--green)">Created → <a href="'+d.url+'">'+d.url+'</a></span>';document.getElementById('c-title').value='';document.getElementById('c-slug').value='';document.getElementById('c-content').value='';refresh();}else{document.getElementById('c-result').innerHTML='<span style="color:#c0392b">'+esc(d.error)+'</span>';}}
async function del(id){if(!confirm('Delete?'))return;await fetch('/api/pages/'+id,{method:'DELETE'});refresh();}
function esc(s){const d=document.createElement('div');d.textContent=s||'';return d.innerHTML;}
function timeAgo(s){if(!s)return'—';const d=new Date(s);const diff=Date.now()-d.getTime();if(diff<60000)return'now';if(diff<3600000)return Math.floor(diff/60000)+'m';if(diff<86400000)return Math.floor(diff/3600000)+'h';return Math.floor(diff/86400000)+'d';}
refresh();setInterval(refresh,8000);
</script></body></html>`

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(uiHTML))
}
