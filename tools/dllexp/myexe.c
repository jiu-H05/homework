__declspec(dllimport) int add(int a, int b);
#include <stdio.h>
int main(void) { int r = add(2, 3); printf("add=%d\n", r); return r == 5 ? 0 : 1; }
