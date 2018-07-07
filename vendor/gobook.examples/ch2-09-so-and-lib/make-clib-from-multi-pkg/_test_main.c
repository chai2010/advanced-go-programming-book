// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

#include "main.h"
#include "./number/number.h"

#include <stdio.h>

int main() {
	int a = 10;
	int b = 5;
	int c = 12;

	int x = number_add_mod(a, b, c);
	printf("(%d+%d)%%%d = %d\n", a, b, c, x);

	goPrintln("done");
	return 0;
}
