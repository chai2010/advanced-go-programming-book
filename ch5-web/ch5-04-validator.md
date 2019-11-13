# 5.4 validator request check

Someone in the community used *Figure 5-10* to mock PHP:

![validate process](../images/ch6-04-validate.jpg)

*Figure 5-10 validator process*

This is actually a language-independent scenario. There are many situations where field validation is required. Form or JSON submission of a Web system is just a typical example. We use Go to write a verification example similar to the one above. Then study how to improve it step by step.

## 5.4.1 Refactoring request validation function

Suppose our data has been bound to a specific structure through an open source binding library.

```go
Type RegisterReq struct {
Username string `json:"username"`
PasswordNew string `json:"password_new"`
PasswordRepeat string `json:"password_repeat"`
Email string `json:"email"`
}

Func register(req RegisterReq) error{
If len(req.Username) > 0 {
If len(req.PasswordNew) > 0 && len(req.PasswordRepeat) > 0 {
If req.PasswordNew == req.PasswordRepeat {
If emailFormatValid(req.Email) {
createUser()
Return nil
} else {
Return errors.New("invalid email")
}
} else {
Return errors.New("password and reinput must be the same")
}
} else {
Return errors.New("password and password reinput must be longer than 0")
}
} else {
Return errors.New("length of username cannot be 0")
}
}
```

We used Go to successfully write the arrow type code that waved open. . How is this code optimized?

Quite simply, the scheme has been given in the book Refactoring: [Guard Clauses] (https://refactoring.com/catalog/replaceNestedConditionalWithGuardClauses.html).

```go
Func register(req RegisterReq) error{
If len(req.Username) == 0 {
Return errors.New("length of username cannot be 0")
}

If len(req.PasswordNew) == 0 || len(req.PasswordRepeat) == 0 {
Return errors.New("password and password reinput must be longer than 0")
}

If req.PasswordNew != req.PasswordRepeat {
Return errors.New("password and reinput must be the same")
}

If emailFormatValid(req.Email) {
Return errors.New("invalid email")
}

createUser()
Return nil
}
```

The code is cleaner and doesn't look so awkward. This is a more general refactoring concept. Although the refactoring method is used to make our validation process code look elegant, we still have to write a similar set of `validate()` functions for each `http` request. Is there a better way? To help us lift this manual labor? The answer is the validator.

## 5.4.2 Emancipate physical labor with validator

From a design perspective, we will definitely declare a structure for each request. The verification scenarios mentioned in the previous section can all be done through the validator. Also take the structure in the previous article as an example. For the sake of beauty, we will first omit the json tag.

Here we introduce a new validator library:

> https://github.com/go-playground/validator

```go
Import "gopkg.in/go-playground/validator.v9"

Type RegisterReq struct {
// The gt=0 of the string indicates that the length must be > 0, gt = greater than
Username string `validate:"gt=0"`
// Same as above
PasswordNew string `validate:"gt=0"`
// eqfield cross-field equality check
PasswordRepeat string `validate:"eqfield=PasswordNew"`
// legal email format check
Email string `validate:"email"`
}

Validate := validator.New()

Func validate(req RegisterReq) error {
Err := validate.Struct(req)
If err != nil {
doSomething()
Return err
}
...
}

```

This eliminates the need to write a duplicate `validate()` function before each request enters the business logic. In this example, only a few features of this validator are listed.

We tried to run this program and the input parameters were set to:

```go
//...

Var req = RegisterReq {
Username : "Xargin",
PasswordNew : "ohno",
PasswordRepeat : "ohn",
Email : "alex@abc.com",
}

Err := validate(req)
fmt.Println(err)

// Key: 'RegisterReq.PasswordRepeat' Error: Field validation for
// 'PasswordRepeat' failed on the 'eqfield' tag
```

If you feel that the error message provided by this `validator` is not user-friendly, for example, to return the error message to the user, you should not display the English directly. Error information can be customized for each tag, and readers can explore it on their own.

## 5.4.3 Principle

From a structural point of view, each structure can be seen as a tree. Suppose we have a structure defined as follows:

```go
Type Nested struct {
Email string `validate:"email"`
}
Type T struct {
Age int `validate:"eq=10"`
Nested Nested
}
```

Draw this structure as a tree, see *Figure 5-11*:

![struct-tree](../images/ch6-04-validate-struct-tree.png)

*Figure 5-11 validator tree*

From the perspective of field validation requirements, it is possible to traverse this tree of structures, whether we use depth-first search or breadth-first search.

Let's write a traversal example of a recursive depth-first search:

```go
Package main

Import (
"fmt"
"reflect"
"regexp"
"strconv"
"strings"
)

Type Nested struct {
Email string `validate:"email"`
}
Type T struct {
Age int `validate:"eq=10"`
Nested Nested
}

Func validateEmail(input string) bool {
If pass, _ := regexp.MatchString(
`^([\w\.\_]{2,10})@(\w{1,}).([a-z]{2,4})$`, input,
Pass;
Return true
}
Return false
}

Func validate(v interface{}) (bool, string) {
validateResult := true
Errmsg := "success"
Vt := reflect.TypeOf(v)
Vv := reflect.ValueOf(v)
For i := 0; i < vv.NumField(); i++ {
fieldVal := vv.Field(i)
tagContent := vt.Field(i).Tag.Get("validate")
k := fieldVal.Kind()

Switch k {
Case reflect.Int:
Val := fieldVal.Int()
tagValStr := strings.Split(tagContent, "=")
tagVal, _ := strconv.ParseInt(tagValStr[1], 10, 64)
If val != tagVal {
Errmsg = "validate int failed, tag is: "+ strconv.FormatInt(
tagVal, 10,
)
validateResult = false
}
Case reflect.String:
Val := fieldVal.String()
tagValStr := tagContent
Switch tagValStr {
Case "email":
nestedResult := validateEmail(val)
If nestedResult == false {
Errmsg = "validate mail failed, field val is: "+ val
validateResult = false
}
}
Case reflect.Struct:
// If there is an embedded struct, then depth-first traversal
// is a recursive process
valInter := fieldVal.Interface()
nestedResult, msg := validate(valInter)
If nestedResult == false {
validateResult = false
Errmsg = msg
}
}
}
Return validateResult, errmsg
}

Func main() {
Var a = T{Age: 10, Nested: Nested{Email: "abc@abc.com"}}

validateResult, errmsg := validate(a)
fmt.Println(validateResuLt, errmsg)
}
```

Here we simply support the two tags `eq=x` and `email`. The reader can make a simple modification to this program to see the specific validate effect. In order to demonstrate the streamlining of error handling and complex processing, such as `reflect.Int8/16/32/64`, `reflect.Ptr` and other types of processing, if you write a verification library for the production environment, please be sure to do Functional perfection and fault tolerance.

The open source validation component introduced in the previous section is far more functionally complex than our example here. But the principle is very simple, that is, tree traversal of the structure with reflection. A thoughtful reader may have a problem at this time. We use a large amount of reflection when verifying the structure, and Go's reflection is not very good in performance, and sometimes even affects the performance of our program. This kind of consideration does have some truths, but the scene that requires a lot of verification of the structure often appears in the Web service. This is not necessarily the performance bottleneck of the program. The actual effect is to make a more accurate judgment from pprof.

What if the reflection-based verification really becomes the performance bottleneck for your service? There is also an idea to avoid reflection: use Go's built-in Parser to scan the source code and then generate validation code based on the definition of the structure. We can put all the structures that need to be verified in a separate package. This is left to the reader to explore.