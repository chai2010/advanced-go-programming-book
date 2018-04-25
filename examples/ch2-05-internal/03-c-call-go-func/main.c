// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

#include <stdio.h>

int main() {
	extern int sum(int a, int b);
	printf("1+1=%d\n", sum(1, 1));
	return 0;
}
