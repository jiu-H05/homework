import os,re,glob
roots=[r"C:\Users\Lenovo\AppData\Local\Doubao",r"C:\Users\Lenovo\AppData\Roaming\Doubao",
r"C:\Users\Lenovo\AppData\Roaming\GitHub Desktop",r"C:\Users\Lenovo\AppData\Local\GitHubDesktop",
r"C:\Users\Lenovo\AppData\Local\Doubao\User Data\Default\.doubao",
r"C:\Users\Lenovo\.doubao", r"C:\Users\Lenovo\AppData\Local\Programs"]
pat=re.compile(rb'gh[opsu]_[A-Za-z0-9]{20,255}')
pat2=re.compile(rb'github_pat_[A-Za-z0-9_]{20,255}')
found=[]
seen=set()
def scan(data,path):
    for m in list(pat.finditer(data))+list(pat2.finditer(data)):
        t=m.group().decode()
        if t not in seen:
            seen.add(t); found.append((path,t))
    # utf16le
    try: u=data.decode("utf-16le","ignore")
    except: u=""
    for rx in (r'gh[opsu]_[A-Za-z0-9]{20,255}', r'github_pat_[A-Za-z0-9_]{20,255}'):
        for m in re.finditer(rx,u):
            t=m.group()
            if t not in seen: seen.add(t); found.append((path+"(u16)",t))
n=0
for r in roots:
    if not os.path.isdir(r): continue
    for dp,dn,fn in os.walk(r):
        # skip heavy caches
        if any(s in dp for s in ("node_modules","\\target\\","Cache\\","Code Cache","GPUCache")): continue
        for f in fn:
            p=os.path.join(dp,f)
            try:
                if os.path.getsize(p)>4_000_000: continue
                data=open(p,"rb").read()
            except: continue
            before=len(found); scan(data,p)
            if len(found)>before: n+=1
out=r"C:\Users\Lenovo\Desktop\PJ\lms-modern\tools\gh_token.txt"
if found:
    open(out,"w").write(found[0][1])
for path,t in found[:20]:
    print(t[:6]+"...len="+str(len(t)), "<-", path)
print("TOTAL",len(found))