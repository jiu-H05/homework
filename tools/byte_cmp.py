import urllib.request
url="https://raw.githubusercontent.com/jiu-H05/homework/main/lms-modern/backend/internal/config/config.go"
rem=urllib.request.urlopen(url,timeout=30).read()
loc=open(r"C:\Users\Lenovo\Desktop\PJ\lms-modern\backend\internal\config\config.go","rb").read()
print("remote bytes",len(rem),"CRLF",rem.count(b"\r\n"),"loneLF",rem.count(b"\n")-rem.count(b"\r\n"))
print("local  bytes",len(loc),"CRLF",loc.count(b"\r\n"),"loneLF",loc.count(b"\n")-loc.count(b"\r\n"))
n=min(len(rem),len(loc))
for i in range(n):
    if rem[i]!=loc[i]:
        print("first diff at",i,"remote",rem[max(0,i-15):i+15],"local",loc[max(0,i-15):i+15]); break
else:
    print("common prefix equal up to",n,"; tails rem",rem[n:n+20],"loc",loc[n:n+20])