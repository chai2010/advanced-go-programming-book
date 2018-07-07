// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

#include "mystring.h"

#include <string.h>

static char buffer[1024];

static char* malloc(int size) {
	return &buffer[0];
}

static void free(void* p) {
	//
}

char* make_string(const char* s) {
	char* p = malloc(strlen(s)+1);
	strcpy(p, s);
	return p;
}

void free_string(char* s) {
	free(s);
}
