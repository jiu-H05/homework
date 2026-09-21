#include <windows.h>
__declspec(dllexport) int add(int a, int b) { return a + b; }
BOOL WINAPI DllMain(HINSTANCE h, DWORD reason, LPVOID r) { (void)h;(void)reason;(void)r; return TRUE; }
