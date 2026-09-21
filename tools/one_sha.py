import os,hashlib
p=r"C:\Users\Lenovo\Desktop\PJ\lms-modern\backend\internal\config\config.go"
d=open(p,"rb").read(); s=hashlib.sha1(b"blob %d\0"%len(d)+d).hexdigest(); print("LOCAL",s)