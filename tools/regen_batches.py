import os, json, glob
root=r"C:\Users\Lenovo\Desktop\PJ\lms-modern"
for f in glob.glob(os.path.join(root,"tools","sync_*.json")): os.remove(f)
todo=["backend/go.sum","frontend/package-lock.json",
"backend/internal/config/config.go","backend/internal/httpapi/respond.go","backend/internal/httpapi/router.go","backend/internal/models/models.go","backend/internal/service/auth.go","backend/internal/service/flow_test.go","backend/internal/service/policy.go","backend/internal/service/policy_test.go","backend/internal/service/readers.go","backend/internal/service/service.go","backend/internal/store/books.go","backend/tests/integration/api_test.go","frontend/.gitignore","frontend/src-tauri/src/main.rs","frontend/src/js/app.js","frontend/src/js/views/dashboard.js","frontend/src/styles.css","frontend/src/vendor/http.global.js"]
LIMIT=14000; batches=[]; cur=[]; cursz=0
for rel in todo:
    content=open(os.path.join(root,*rel.split("/")),"rb").read().decode("utf-8"); size=len(content.encode())
    if cur and cursz+size>LIMIT: batches.append(cur); cur=[]; cursz=0
    cur.append({"path":"lms-modern/"+rel,"content":content}); cursz+=size
if cur: batches.append(cur)
for i,b in enumerate(batches,1):
    fp=os.path.join(root,"tools","sync_%02d.json"%i)
    json.dump(b,open(fp,"w",encoding="utf-8"),ensure_ascii=False)
    print("sync_%02d"%i, [x["path"].split("lms-modern/")[1] for x in b], sum(len(x['content'].encode()) for x in b))