#include "person.h"

#include <stdio.h>

extern "C" void Main() {
	auto p = Person::New("gopher", 10);

	char buf[64];
	char* name = p->GetName(buf, sizeof(buf)-1);
	int age = p->GetAge();

	printf("%s, %d years old.\n", name, age);
	p->Delete();
}
