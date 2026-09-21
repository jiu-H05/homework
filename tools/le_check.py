import os
root=r"C:\Users\Lenovo\Desktop\PJ\lms-modern"
changed=["backend/internal/config/config.go","backend/internal/httpapi/respond.go","backend/internal/httpapi/router.go","backend/internal/models/models.go","backend/internal/service/auth.go","backend/internal/service/flow_test.go","backend/internal/service/policy.go","backend/internal/service/policy_test.go","backend/internal/service/readers.go","backend/internal/service/service.go","backend/internal/store/books.go","backend/tests/integration/api_test.go","frontend/.gitignore","frontend/src-tauri/src/main.rs","frontend/src/js/app.js","frontend/src/js/views/dashboard.js","frontend/src/styles.css","frontend/src/vendor/http.global.js"]
import json
local=json.load(open(os.path.join(root,"tools","local_inventory.json"),encoding="utf-8"))
def le(rel):
    d=open(os.path.join(root,*rel.split("/")),"rb").read()
    crlf=d.count(b"\r\n"); lf=d.count(b"\n")-crlf
    return ("CRLF" if crlf and lf==0 else "LF" if lf and crlf==0 else "MIXED", crlf, lf)
for rel in changed:
    print(le(rel)[0], rel)
print("--- sample unchanged ---")
for rel in ["backend/go.mod","frontend/src/js/api.js","backend/internal/store/db.go","frontend/src-tauri/src/lib.rs"]:
    print(le(rel)[0], rel)