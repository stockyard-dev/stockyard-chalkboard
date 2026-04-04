package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Chalkboard</title>
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet">
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c94444;--blue:#5b8dd9;--mono:'JetBrains Mono',monospace}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--mono);line-height:1.5}
.hdr{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}.hdr h1{font-size:.9rem;letter-spacing:2px}.hdr h1 span{color:var(--rust)}
.main{padding:1.5rem;max-width:960px;margin:0 auto}
.stats{display:grid;grid-template-columns:repeat(3,1fr);gap:.5rem;margin-bottom:1rem}
.st{background:var(--bg2);border:1px solid var(--bg3);padding:.6rem;text-align:center}
.st-v{font-size:1.2rem;font-weight:700}.st-l{font-size:.5rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-top:.15rem}
.toolbar{display:flex;gap:.5rem;margin-bottom:1rem;align-items:center;flex-wrap:wrap}
.search{flex:1;min-width:180px;padding:.4rem .6rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.search:focus{outline:none;border-color:var(--leather)}
.filter-sel{padding:.4rem .5rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.65rem}
.lesson{background:var(--bg2);border:1px solid var(--bg3);padding:.8rem 1rem;margin-bottom:.5rem;transition:border-color .2s}
.lesson:hover{border-color:var(--leather)}
.lesson-top{display:flex;justify-content:space-between;align-items:flex-start;gap:.5rem}
.lesson-title{font-size:.85rem;font-weight:700}
.lesson-sub{font-size:.7rem;color:var(--cd);margin-top:.1rem}
.lesson-content{font-size:.65rem;color:var(--cm);margin-top:.3rem;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden}
.lesson-meta{font-size:.55rem;color:var(--cm);margin-top:.3rem;display:flex;gap:.5rem;flex-wrap:wrap;align-items:center}
.lesson-actions{display:flex;gap:.3rem;flex-shrink:0}
.badge{font-size:.5rem;padding:.12rem .35rem;text-transform:uppercase;letter-spacing:1px;border:1px solid}
.badge.draft{border-color:var(--gold);color:var(--gold)}.badge.ready{border-color:var(--green);color:var(--green)}.badge.taught{border-color:var(--blue);color:var(--blue)}
.subj-badge{font-size:.5rem;padding:.1rem .3rem;background:var(--bg3);color:var(--cd)}
.tag{font-size:.45rem;padding:.1rem .25rem;background:var(--bg3);color:var(--cm)}
.btn{font-size:.6rem;padding:.25rem .5rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd);transition:all .2s}
.btn:hover{border-color:var(--leather);color:var(--cream)}.btn-p{background:var(--rust);border-color:var(--rust);color:#fff}
.btn-sm{font-size:.55rem;padding:.2rem .4rem}
.modal-bg{display:none;position:fixed;inset:0;background:rgba(0,0,0,.65);z-index:100;align-items:center;justify-content:center}.modal-bg.open{display:flex}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:500px;max-width:92vw;max-height:90vh;overflow-y:auto}
.modal h2{font-size:.8rem;margin-bottom:1rem;color:var(--rust);letter-spacing:1px}
.fr{margin-bottom:.6rem}.fr label{display:block;font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-bottom:.2rem}
.fr input,.fr select,.fr textarea{width:100%;padding:.4rem .5rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.fr input:focus,.fr select:focus,.fr textarea:focus{outline:none;border-color:var(--leather)}
.row2{display:grid;grid-template-columns:1fr 1fr;gap:.5rem}
.row3{display:grid;grid-template-columns:1fr 1fr 1fr;gap:.5rem}
.acts{display:flex;gap:.4rem;justify-content:flex-end;margin-top:1rem}
.empty{text-align:center;padding:3rem;color:var(--cm);font-style:italic;font-size:.75rem}
</style></head><body>
<div class="hdr"><h1><span>&#9670;</span> CHALKBOARD</h1><button class="btn btn-p" onclick="openForm()">+ New Lesson</button></div>
<div class="main">
<div class="stats" id="stats"></div>
<div class="toolbar">
<input class="search" id="search" placeholder="Search lessons..." oninput="render()">
<select class="filter-sel" id="status-filter" onchange="render()"><option value="">All Status</option><option value="draft">Draft</option><option value="ready">Ready</option><option value="taught">Taught</option></select>
</div>
<div id="list"></div>
</div>
<div class="modal-bg" id="mbg" onclick="if(event.target===this)closeModal()"><div class="modal" id="mdl"></div></div>
<script>
var A='/api',items=[],editId=null;
async function load(){var r=await fetch(A+'/lessons').then(function(r){return r.json()});items=r.lessons||[];renderStats();render();}
function renderStats(){var total=items.length;var subjects={};items.forEach(function(l){if(l.subject)subjects[l.subject]=true});
var totalMin=items.reduce(function(s,l){return s+(l.duration||0)},0);
document.getElementById('stats').innerHTML='<div class="st"><div class="st-v">'+total+'</div><div class="st-l">Lessons</div></div><div class="st"><div class="st-v">'+Object.keys(subjects).length+'</div><div class="st-l">Subjects</div></div><div class="st"><div class="st-v">'+totalMin+'</div><div class="st-l">Total Min</div></div>';}
function render(){var q=(document.getElementById('search').value||'').toLowerCase();var sf=document.getElementById('status-filter').value;var f=items;
if(sf)f=f.filter(function(l){return l.status===sf});
if(q)f=f.filter(function(l){return(l.title||'').toLowerCase().includes(q)||(l.subject||'').toLowerCase().includes(q)||(l.content||'').toLowerCase().includes(q)});
if(!f.length){document.getElementById('list').innerHTML='<div class="empty">No lessons planned.</div>';return;}
var h='';f.forEach(function(l){
h+='<div class="lesson"><div class="lesson-top"><div style="flex:1">';
h+='<div class="lesson-title">'+esc(l.title)+'</div>';
var sub=[];if(l.subject)sub.push(l.subject);if(l.grade)sub.push('Grade '+l.grade);
if(sub.length)h+='<div class="lesson-sub">'+esc(sub.join(' &#183; '))+'</div>';
h+='</div><div class="lesson-actions">';
h+='<button class="btn btn-sm" onclick="openEdit(''+l.id+'')">Edit</button>';
h+='<button class="btn btn-sm" onclick="del(''+l.id+'')" style="color:var(--red)">&#10005;</button>';
h+='</div></div>';
if(l.content)h+='<div class="lesson-content">'+esc(l.content)+'</div>';
h+='<div class="lesson-meta">';
if(l.status)h+='<span class="badge '+l.status+'">'+l.status+'</span>';
if(l.subject)h+='<span class="subj-badge">'+esc(l.subject)+'</span>';
if(l.duration)h+='<span>'+l.duration+' min</span>';
if(l.tags){l.tags.split(',').forEach(function(t){t=t.trim();if(t)h+='<span class="tag">#'+esc(t)+'</span>';});}
h+='</div></div>';});
document.getElementById('list').innerHTML=h;}
async function del(id){if(!confirm('Delete?'))return;await fetch(A+'/lessons/'+id,{method:'DELETE'});load();}
function formHTML(lesson){var i=lesson||{title:'',subject:'',content:'',grade:'',duration:0,status:'draft',tags:''};var isEdit=!!lesson;
var h='<h2>'+(isEdit?'EDIT':'NEW')+' LESSON</h2>';
h+='<div class="fr"><label>Title *</label><input id="f-title" value="'+esc(i.title)+'"></div>';
h+='<div class="row3"><div class="fr"><label>Subject</label><input id="f-subj" value="'+esc(i.subject)+'" placeholder="Math"></div>';
h+='<div class="fr"><label>Grade</label><input id="f-grade" value="'+esc(i.grade)+'"></div>';
h+='<div class="fr"><label>Duration (min)</label><input id="f-dur" type="number" value="'+(i.duration||'')+'"></div></div>';
h+='<div class="fr"><label>Content</label><textarea id="f-content" rows="5">'+esc(i.content)+'</textarea></div>';
h+='<div class="row2"><div class="fr"><label>Status</label><select id="f-status">';
['draft','ready','taught'].forEach(function(s){h+='<option value="'+s+'"'+(i.status===s?' selected':'')+'>'+s.charAt(0).toUpperCase()+s.slice(1)+'</option>';});
h+='</select></div><div class="fr"><label>Tags</label><input id="f-tags" value="'+esc(i.tags)+'" placeholder="comma separated"></div></div>';
h+='<div class="acts"><button class="btn" onclick="closeModal()">Cancel</button><button class="btn btn-p" onclick="submit()">'+(isEdit?'Save':'Create')+'</button></div>';
return h;}
function openForm(){editId=null;document.getElementById('mdl').innerHTML=formHTML();document.getElementById('mbg').classList.add('open');}
function openEdit(id){var l=null;for(var j=0;j<items.length;j++){if(items[j].id===id){l=items[j];break;}}if(!l)return;editId=id;document.getElementById('mdl').innerHTML=formHTML(l);document.getElementById('mbg').classList.add('open');}
function closeModal(){document.getElementById('mbg').classList.remove('open');editId=null;}
async function submit(){var title=document.getElementById('f-title').value.trim();if(!title){alert('Title required');return;}
var body={title:title,subject:document.getElementById('f-subj').value.trim(),content:document.getElementById('f-content').value.trim(),grade:document.getElementById('f-grade').value.trim(),duration:parseInt(document.getElementById('f-dur').value)||0,status:document.getElementById('f-status').value,tags:document.getElementById('f-tags').value.trim()};
if(editId){await fetch(A+'/lessons/'+editId,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});}
else{await fetch(A+'/lessons',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});}
closeModal();load();}
function esc(s){if(!s)return'';var d=document.createElement('div');d.textContent=s;return d.innerHTML;}
document.addEventListener('keydown',function(e){if(e.key==='Escape')closeModal();});
load();
</script></body></html>`
