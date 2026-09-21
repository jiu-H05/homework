import os, json
root = r"C:\Users\Lenovo\Desktop\PJ\lms-modern"
GH = {
"backend/go.mod":"781a2e10448a53cf14529b287fbc0dc7fe1a0950",
"backend/cmd/server/main.go":"9039651bbf1fe994ce98edf99c162ded86d94728",
"backend/internal/config/config.go":"3b6bb9e03ecfe697e3da95995b89a6358f55322b",
"backend/internal/httpapi/auth_handler.go":"0fae9909e9ac15d4eba232f1d165c40e58583de2",
"backend/internal/httpapi/book_handler.go":"6ba0c4d88817e81c8d3be41d93c35bd91ca9280c",
"backend/internal/httpapi/borrow_handler.go":"0cbe4b24ac742e47d3a398f7165acb8bf2cf3f07",
"backend/internal/httpapi/reader_handler.go":"7c290bcf5f0add3d834e5feb561578a6d0028904",
"backend/internal/httpapi/respond.go":"39e34af5784cc9cbba412e0dffd7dc999356c92a",
"backend/internal/httpapi/router.go":"4b2a9fa6e522042cd4e74f9defd97a7120d60e44",
"backend/internal/httpapi/stats_handler.go":"5549b0ce25a5bcabd0a5095c1971e962b2d63856",
"backend/internal/httpapi/token.go":"37be38c1953c3a631c80932341025094589700dd",
"backend/internal/models/models.go":"6f677bc6a485222ea619bcd0873c53568aa1e2f9",
"backend/internal/service/auth.go":"8518b873c1919b095facfc5b048341b06da5de64",
"backend/internal/service/books.go":"ce93b4c5d1f153f34117451e156a7c19ce0c9507",
"backend/internal/service/borrows.go":"63b29eaab9e60b5f443c81dcab94481b67255a69",
"backend/internal/service/flow_test.go":"40f62b0cd1c27dcf8bf6db252e4cd5b3a21407fb",
"backend/internal/service/policy.go":"63db9cfa1ffd9662ce2b22989aac48797952df92",
"backend/internal/service/policy_test.go":"f0ef7a744e386fff75bcb962adc7be9bcca99f31",
"backend/internal/service/readers.go":"6087195cfaf16cb1c36f9bbb18559bf27d5f1853",
"backend/internal/service/service.go":"48c5a40c7a451a95a87d24ef1b719dea93289d9b",
"backend/internal/service/stats.go":"b918dc46f7108be5de10d63dc1f61c44bcaaf035",
"backend/internal/store/books.go":"e61ea19815373e6a626ef0d1d3731a60b8b928b4",
"backend/internal/store/borrows.go":"fd814d1ab48d9925da9f1042b9652dbc43c39b55",
"backend/internal/store/db.go":"098868f23d2a60cae0b941d85fdee430a47ed485",
"backend/internal/store/seed.go":"5d9568b81039202032b44a31e51a18de443fd138",
"backend/internal/store/users.go":"24c4888e9b3300bc79a02e9a642789f6160f2d9a",
"backend/tests/integration/api_test.go":"5804d6f3a75ebf564b951d7065a9365ac1272a3c",
"frontend/.gitignore":"8282fdc43e9a3f1d9df0ad025a0a54793454f678",
"frontend/.vscode/extensions.json":"d4fef4e317512098ded392e804e89c8566a64619",
"frontend/README.md":"31c0cc39b3bdba57c09abac4146d027fbfb2bb31",
"frontend/package.json":"3ae0a082e9e44a343d7a328426e9c13f2b95acaa",
"frontend/src/index.html":"9ebfaaacd550499e87c6484c96de5ee513a1ebf9",
"frontend/src/styles.css":"c9558a2506822ae22cf7419b41f5425ad303e7b9a",
"frontend/src/js/api.js":"4f08ced7ddd8e18f2479a74e1ec6327972dfc513",
"frontend/src/js/app.js":"f7bc0072e2e1069a8a207f32015158f20b769d25",
"frontend/src/js/components.js":"60d83387f678ded7ec428931ca2ab78bcc611996",
"frontend/src/js/login.js":"b1d2904853f430b7c210fd13db777426a5dd1459",
"frontend/src/js/net.js":"14eb8de6b23822d6549c2aab49cace013546aedb",
"frontend/src/js/views/books.js":"52d14bb89017c435ae83fef7f1182a7159de25c5",
"frontend/src/js/views/borrows.js":"8302ef9e6950758ac12f37017ecaf232289387f3",
"frontend/src/js/views/catalog.js":"dd01707ce302fc155f7617c93702acd7d6215184",
"frontend/src/js/views/dashboard.js":"087cabfecab429b08db2147fda19ba13702fab5a",
"frontend/src/js/views/logs.js":"2a86a6809de3d831dfd1cf068fa5dae0b3cabe20",
"frontend/src/js/views/myborrows.js":"658d8f42a810e2e388a5473c69886fe57e8cd092",
"frontend/src/js/views/profile.js":"f1564f0f44b12963fb42503946bcabfdff370216",
"frontend/src/js/views/readers.js":"4e206e6882444a5e7c01e860610ecaa5ec34bf62",
"frontend/src/vendor/entry.js":"6b581e06515382fc2338df141b0b915563614ffa",
"frontend/src/vendor/http.global.js":"e755ce41e3b1caa91ff3d46d6bc75d5ec1644c2e",
"frontend/src-tauri/.gitignore":"d8769d02d149255200f1f89102bb0fe3b9893094",
"frontend/src-tauri/Cargo.toml":"064fde9cc97229c0b5489f94e4376c1d400c8324",
"frontend/src-tauri/build.rs":"2ba80a8bff12786f28420a30260b9949355066b4",
"frontend/src-tauri/tauri.conf.json":"45c351e6465e0329ef27cbe3da3deec86f7bd20a",
"frontend/src-tauri/capabilities/default.json":"cbe06bf1fffe730f77f4ac2a7cd74d785bd95109",
"frontend/src-tauri/src/lib.rs":"9ff85d7acf134d6b9545937d78d25961f6c6103f",
"frontend/src-tauri/src/main.rs":"ba7508d28d89b9a4c0436751ef8961f3d9146068",
}
local = json.load(open(os.path.join(root,"tools","local_inventory.json"),encoding="utf-8"))
missing = [p for p in local if p not in GH]
changed = [p for p in local if p in GH and local[p] != GH[p]]
extra = [p for p in GH if p not in local]
print("missing(%d):"%len(missing)); [print("  ",p) for p in sorted(missing)]
print("changed(%d):"%len(changed)); [print("  ",p) for p in sorted(changed)]
print("extra(%d):"%len(extra)); [print("  ",p) for p in sorted(extra)]

# build push batches from exact local content
todo = sorted(missing+changed)
LIMIT=60000
batches=[]; cur=[]; cursz=0
for rel in todo:
    content = open(os.path.join(root,*rel.split("/")),"rb").read().decode("utf-8")
    size=len(content.encode("utf-8"))
    if cur and cursz+size>LIMIT:
        batches.append(cur); cur=[]; cursz=0
    cur.append({"path":"lms-modern/"+rel,"content":content}); cursz+=size
if cur: batches.append(cur)
for i,b in enumerate(batches,1):
    fp=os.path.join(root,"tools","sync_%02d.json"%i)
    json.dump(b,open(fp,"w",encoding="utf-8"),ensure_ascii=False)
    print("batch",i,"files",len(b),"bytes",sum(len(x['content'].encode()) for x in b),"->",os.path.basename(fp))