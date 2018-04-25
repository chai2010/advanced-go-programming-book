// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

#include "./my_buffer.h"

extern "C" {
	#include "./my_buffer_capi.h"
}

struct MyBuffer_T: MyBuffer {
	MyBuffer_T(int size): MyBuffer(size) {}
	~MyBuffer_T() {}
};

MyBuffer_T* NewMyBuffer(int size) {
	auto p = new MyBuffer_T(size);
	return p;
}
void DeleteMyBuffer(MyBuffer_T* p) {
	delete p;
}

char* MyBuffer_Data(MyBuffer_T* p) {
	return p->Data();
}
int MyBuffer_Size(MyBuffer_T* p) {
	return p->Size();
}
