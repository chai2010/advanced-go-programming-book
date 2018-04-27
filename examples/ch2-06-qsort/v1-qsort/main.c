// Copyright © 2018 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

#include <stdio.h>
#include <stdlib.h>

#define DIM(x) (sizeof(x)/sizeof((x)[0]))

static int cmp(const void* a, const void* b) {
	const int* pa = (int*)a;
	const int* pb = (int*)b;
	return *pa - *pb;
}

int main() {
	int values[] = { 42, 8, 109, 97, 23, 25 };
	int i;

	qsort(values, DIM(values), sizeof(values[0]), cmp);

	for(i = 0; i < DIM(values); i++) {
		printf ("%d ",values[i]);
	}
	printf("\n");

	return 0;
}
