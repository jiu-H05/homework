import os, hashlib, json, posixpath
root = r"C:\Users\Lenovo\Desktop\PJ\lms-modern"
include_ext = {".go",".mod",".sum",".js",".html",".css",".rs",".json",".toml",".md",".bat",".txt"}
exclude_parts = {"tools","target","node_modules","gen","binaries","icons",".preview"}
exclude_ext = {".exe",".zip",".log",".db",".ico",".png",".icns",".dll",".lnk",".pdb",".lib",".exp",".obj"}
files = {}
for dp, dn, fn in os.walk(root):
    dn[:] = [d for d in dn if d not in exclude_parts]
    for f in fn:
        full = os.path.join(dp, f)
        ext = os.path.splitext(f)[1].lower()
        if ext in exclude_ext: continue
        if f == ".gitignore": pass
        elif ext not in include_ext: continue
        rel = os.path.relpath(full, root).replace(os.sep, "/")
        if any(p in rel.split("/") for p in exclude_parts): continue
        with open(full,"rb") as fh: data = fh.read()
        blob = hashlib.sha1(b"blob %d\0" % len(data) + data).hexdigest()
        files[rel] = blob
out = r"C:\Users\Lenovo\Desktop\PJ\lms-modern\tools\local_inventory.json"
os.makedirs(os.path.dirname(out), exist_ok=True)
with open(out,"w",encoding="utf-8") as fh: json.dump(files, fh, ensure_ascii=False, indent=0)
print("TOTAL", len(files))
for k in sorted(files): print(k)