package main

import (
	"flag"
	"reflect"
	"testing"
)

func TestParseReqParam(t *testing.T) {
	
	f := flag.CommandLine

	// this one must be first - with no leading clearFlags call it
	// verifies our expectation of default values as we reset by
	// clearFlags
	pkgMap := make(map[string]string)
	expected := map[string]string{}
	err := parseReqParam("", f, pkgMap)
	if err != nil {
		t.Errorf("Test 0: unexpected parse error '%v'", err)
	}
	if !reflect.DeepEqual(pkgMap, expected) {
		t.Errorf("Test 0: pkgMap parse error, expected '%v', got '%v'", expected, pkgMap)
	}
	checkFlags(false, false, "-", "", "apidocs", t, 0)

	clearFlags()
	pkgMap = make(map[string]string)
	expected = map[string]string{"google/api/annotations.proto": "github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api"}
	err = parseReqParam("allow_delete_body,allow_merge,file=./foo.pb,import_prefix=/bar/baz,Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api", f, pkgMap)
	if err != nil {
		t.Errorf("Test 1: unexpected parse error '%v'", err)
	}
	if !reflect.DeepEqual(pkgMap, expected) {
		t.Errorf("Test 1: pkgMap parse error, expected '%v', got '%v'", expected, pkgMap)
	}
	checkFlags(true, true, "./foo.pb", "/bar/baz", "apidocs", t, 1)

	clearFlags()
	pkgMap = make(map[string]string)
	expected = map[string]string{"google/api/annotations.proto": "github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api"}
	err = parseReqParam("allow_delete_body=true,allow_merge=true,merge_file_name=test_name,file=./foo.pb,import_prefix=/bar/baz,Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api", f, pkgMap)
	if err != nil {
		t.Errorf("Test 2: unexpected parse error '%v'", err)
	}
	if !reflect.DeepEqual(pkgMap, expected) {
		t.Errorf("Test 2: pkgMap parse error, expected '%v', got '%v'", expected, pkgMap)
	}
	checkFlags(true, true,"./foo.pb", "/bar/baz", "test_name", t, 2)

	clearFlags()
	pkgMap = make(map[string]string)
	expected = map[string]string{"a/b/c.proto": "github.com/x/y/z", "f/g/h.proto": "github.com/1/2/3/"}
	err = parseReqParam("allow_delete_body=false,allow_merge=false,Ma/b/c.proto=github.com/x/y/z,Mf/g/h.proto=github.com/1/2/3/", f, pkgMap)
	if err != nil {
		t.Errorf("Test 3: unexpected parse error '%v'", err)
	}
	if !reflect.DeepEqual(pkgMap, expected) {
		t.Errorf("Test 3: pkgMap parse error, expected '%v', got '%v'", expected, pkgMap)
	}
	checkFlags(false, false,"stdin", "", "apidocs", t, 3)

	clearFlags()
	pkgMap = make(map[string]string)
	expected = map[string]string{}
	err = parseReqParam("", f, pkgMap)
	if err != nil {
		t.Errorf("Test 4: unexpected parse error '%v'", err)
	}
	if !reflect.DeepEqual(pkgMap, expected) {
		t.Errorf("Test 4: pkgMap parse error, expected '%v', got '%v'", expected, pkgMap)
	}
	checkFlags(false, false, "stdin", "", "apidocs", t, 4)

	clearFlags()
	pkgMap = make(map[string]string)
	expected = map[string]string{}
	err = parseReqParam("unknown_param=17", f, pkgMap)
	if err == nil {
		t.Error("Test 5: expected parse error not returned")
	}
	if !reflect.DeepEqual(pkgMap, expected) {
		t.Errorf("Test 5: pkgMap parse error, expected '%v', got '%v'", expected, pkgMap)
	}
	checkFlags(false, false,"stdin", "", "apidocs", t, 5)

	clearFlags()
	pkgMap = make(map[string]string)
	expected = map[string]string{}
	err = parseReqParam("Mfoo", f, pkgMap)
	if err == nil {
		t.Error("Test 6: expected parse error not returned")
	}
	if !reflect.DeepEqual(pkgMap, expected) {
		t.Errorf("Test 6: pkgMap parse error, expected '%v', got '%v'", expected, pkgMap)
	}
	checkFlags(false, false,"stdin", "", "apidocs", t, 6)

	clearFlags()
	pkgMap = make(map[string]string)
	expected = map[string]string{}
	err = parseReqParam("allow_delete_body,file,import_prefix,allow_merge,merge_file_name", f, pkgMap)
	if err != nil {
		t.Errorf("Test 7: unexpected parse error '%v'", err)
	}
	if !reflect.DeepEqual(pkgMap, expected) {
		t.Errorf("Test 7: pkgMap parse error, expected '%v', got '%v'", expected, pkgMap)
	}
	checkFlags(true, true, "", "", "", t, 7)

}

func checkFlags(allowDeleteV, allowMergeV bool, fileV, importPathV, mergeFileNameV string, t *testing.T, tid int) {
	if *importPrefix != importPathV {
		t.Errorf("Test %v: import_prefix misparsed, expected '%v', got '%v'", tid, importPathV, *importPrefix)
	}
	if *file != fileV {
		t.Errorf("Test %v: file misparsed, expected '%v', got '%v'", tid, fileV, *file)
	}
	if *allowDeleteBody != allowDeleteV {
		t.Errorf("Test %v: allow_delete_body misparsed, expected '%v', got '%v'", tid, allowDeleteV, *allowDeleteBody)
	}
	if *allowMerge != allowMergeV {
		t.Errorf("Test %v: allow_merge misparsed, expected '%v', got '%v'", tid, allowMergeV, *allowMerge)
	}
	if *mergeFileName != mergeFileNameV {
		t.Errorf("Test %v: merge_file_name misparsed, expected '%v', got '%v'", tid, mergeFileNameV, *mergeFileName)
	}
}

func clearFlags() {
	*importPrefix = ""
	*file = "stdin"
	*allowDeleteBody = false
	*allowMerge = false
	*mergeFileName = "apidocs"
}
