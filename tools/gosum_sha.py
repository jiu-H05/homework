import hashlib
p=r"C:\Users\Lenovo\Desktop\PJ\lms-modern\backend\go.sum"
d=open(p,"rb").read()
print("LOCAL_GOSUM_SHA",hashlib.sha1(b"blob %d\0"%len(d)+d).hexdigest(),"bytes",len(d))