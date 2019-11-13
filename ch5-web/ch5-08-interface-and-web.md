# 5.8 Interface and Table Driven Development

In the web project, you will often encounter changes in the external dependency environment, such as:

1. The company's old storage system has been in disrepair for a long time. Now no one has maintained it. The new system has not been considered for smooth migration. However, the ultimatum has been completed and the migration is required within N days.
2. The old user system of the platform department has been in disrepair for a long time, and now no one has maintained it. It is a sad story. The new system did not consider compatibility with the old interface, but the ultimatum was already down, requiring migration within N months.
3. The company's old news queue is cool and has been in disrepair. The new technical elites did not consider forward compatibility, but the ultimatum has been lowered and the migration is required within half a year.

Well, so you see, our external dependencies are always constantly upgraded for our own sake, and do not want to be forward compatible, and then give us an ultimatum. If our department is saturated and the leadership is strong, then sometimes the relying party can be forced to do compatibility. But the world is not necessarily like people, even if our leadership is strong, the leadership of readers and friends may still be recognized.

We can think about how to alleviate this problem.

## 5.8.1 Business System Development Process

As long as Internet companies can live for three years, the primary problem facing engineering is code bloat. After the system's code is inflated, the parts of the system that are not related to the business's own process can be disassembled and asynchronous. What is business-related, such as some statistics, anti-cheating, marketing billing, price calculation, user status update and so on. These requirements often depend on the data of the main process, but they are only the side branches of the main process, and they are self-contained.

At this time we can dismantle these side branches and deploy, develop and maintain them as independent systems. The delay of these side-by-side processes is very sensitive. For example, if the user clicks a button on the interface and needs to return immediately (price calculation, payment), then RPC communication with the main process system is required, and when the communication fails, the result is directly returned. To the user. If the delay is not sensitive, such as the lottery system, and the results are published later, or non-real-time statistical systems, then there is no need to do an RPC process for each system in the main process. We only need to package the data needed downstream into a message and pass it into the message queue. The subsequent things have nothing to do with the main process (of course, the follow-up process with the user still needs to be done).

Although some problems have been solved through disassembly and asynchronousization, they cannot solve all problems. As the business develops, the modules of single responsibility will become more and more complex, which is an inevitable trend. If one thing becomes complicated, then disassembly and asynchronousization will not work. We still have to do a certain degree of encapsulation abstraction on the thing itself.

## 5.8.2 Using functions to encapsulate business processes

In the most basic packaging process, we put similar behaviors together, and then package them into a single function, so that our messy code becomes like this:

```go
Func BusinessProcess(ctx context.Context, params Params) (resp, error){
ValidateLogin()
ValidateParams()
AntispamCheck()
GetPrice()
CreateOrder()
UpdateUserStatus()
NotifyDownstreamSystems()
}
```

No matter how complex the business is, the logic within the system can be broken down into processes like `step1 -> step2 -> step3 ...`.

There are also complex processes within each step, such as:

```go
Func CreateOrder() {
ValidateDistrict() // Determine if it is a regionally qualified item
ValidateVIPProduct() // Check if it is only available for vip
GetUserInfo() // Get more detailed user information from the user system
GetProductDesc() // Get the details of the item at that point in time from the item system
DecrementStorage() // deducting inventory
CreateOrderSnapshot() // Create an order snapshot
Return CreateSuccess
}
```

When reading the business process code, we can read the function name to know what has been done in the process. If you need to modify the details, then go to each business step to see the specific process. A well-written business process code will stack all the processes in a few functions, resulting in hundreds or even thousands of rows of functions. This spaghetti-style code reading and maintenance can be very painful. In the development process, a simple package like this one should be performed immediately if there are conditions.

## 5.8.3 Using interfaces to abstract

In the early stage of business development, it is not suitable to introduce interfaces. In many cases, business processes change greatly. Introducing interfaces too early will increase the business system itself by adding unnecessary stratification, resulting in almost complete negation of each modification. Previous work.

When the business develops to a certain stage and the main process is stable, the interface can be used for abstraction. The stability here means that most of the business steps of the main process have been determined. Even if the modifications are made, there will be no large-scale changes, but only minor repairs, or just adding or deleting a small number of business steps.

If we have already packaged the business steps well during the development process, it is very easy to abstract the interface at this time. The pseudo code:

```go
// OrderCreator creates an order process
Type OrderCreator interface {
ValidateDistrict() // Determine if it is a regionally qualified item
ValidateVIPProduct() // Check if it is only available for vip
GetUserInfo() // Get more detailed user information from the user system
GetProductDesc() // Get the details of the item at that point in time from the item system
DecrementStorage() // deducting inventory
CreateOrderSnapshot() // Create an order snapshot
}
```

We can complete the abstraction by referring to the step function signatures we have written before.

Before we abstract, we should understand that the introduction of interfaces makes sense for our system itself, which is to be analyzed according to the scene. If our system only serves one product line, and the internal code is only customized for a very specific scenario, then the introduction of the interface will not bring any benefits. As for whether it is convenient to test, we will talk about this in the following chapters.

If we are doing a platform system that requires a platform to define uniform business processes and business specifications, then interface-based abstraction makes sense. for example:

![interface-impl](../images/ch6-interface-impl.uml.png)

*Figure 5-19 Implementing a public interface*

The platform needs to serve multiple lines of business, but the data definition needs to be unified, so I hope to follow the platform-defined process. As a platform side, we can define a set of interfaces similar to the above, and then require the access side's business to implement these interfaces. If the interface has its unwanted steps, just return `nil`, or ignore it.

When the business is iterating, the platform code is not modified, so we introduce these access services as plugins for platform code. What if we don't have an interface?

```go
Import (
"sample.com/travelorder"
"sample.com/marketorder"
)

Func CreateOrder() {
Switch businessType {
Case TravelBusiness:
travelorder.CreateOrder()
Case MarketBusiness:
marketorder.CreateOrderForMarket()
Default:
Return errors.New("not supported business")
}
}

Func ValidateUser() {
Switch businessType {
Case TravelBusiness:
travelorder.ValidateUserVIP()
Case MarketBusiness:
marketorder.ValidateUserRegistered()
Default:
Return errors.New("not supported business")
}
}

// ...
Switch ...
Switch ...
Switch ...
```

That's right, there is endless `switch`, and endless garbage code. After the introduction of the interface, our `switch` only needs to be done once at the business portal.

```go
Type BusinessInstance interface {
ValidateLogin()
ValidateParams()
AntispamCheck()
GetPrice()
CreateOrder()
UpdateUserStatus()
NotifyDownstreamSystems()
}

Func entry() {
Var bi BusinessInstance
Switch businessType {
Case TravelBusiness:
Bi = travelorder.New()
Case MarketBusiness:
Bi = marketorder.New()
Default:
Return errors.New("not supported business")
}
}

Func BusinessProcess(bi BusinessInstance) {
bi.ValidateLogin()
bi.ValidateParams()
bi.AntispamCheck()
bi.GetPrice()
bi.CreateOrder()
bi.UpdateUserStatus()
bi.NotifyDownstreamSystems()
}
```

Interface-oriented programming, do not care about the specific implementation. If the corresponding service is modified in the iteration, all logic is completely transparent to the platform side.

## 5.8.4 Advantages and Disadvantages of Interface

The most popular place for Go is the orthogonality of its interface design. Modules do not need to know each other's existence. A module defines the interface, and B module can implement this interface. If there is no data type defined in the A module in the interface, then `import A` is not even used in the B module. For example, `io.Writer` in the standard library:

```go
Type Writer interface {
Write(p []byte) (n int, err error)
}
```

We can implement the `io.Writer` interface in our own module:

```go
Type MyType struct {}

Func (m MyType) Write(p []byte) (n int, err error) {
Return 0, nil
}
```

Then we can pass our own `MyType` to any function that uses `io.Writer` as a parameter, such as:

```go
Package log

Func SetOutput(w io.Writer) {
Output = w
}
```

then:

```go
Package my-business

Import "xy.com/log"

Func init() {
log.SetOutput(MyType)
}
```

In the place defined by `MyType`, you can directly implement the `io.Writer` interface without `import "io"`. We can also combine many functions at will to implement various types of interfaces, and interface implementers and interfaces. The definition side does not need to establish the dependencies generated by the import. So many people think that this orthogonality of Go is a very good design.

But this "orthogonal" nature will also bring us some trouble. When we take over a system with hundreds of thousands of rows, if we see an interface that defines a lot of interfaces, such as an order process, we hope to find out directly which objects are implemented by those objects. But until now, this simple requirement has only been implemented by Goland, and the experience is acceptable. Visual Studio Code needs to scan the project globally to see which structures implement all the functions of the interface. Languages ​​that explicitly implement interfaces are much more friendly to IDE interface lookups. On the other hand, we see a structure and hope to know the structure immediately. Which interfaces are implemented, but also have the same problems as mentioned earlier.

Despite the inconvenience, the benefits brought by the interface are self-evident: First, relying on inversion, which is the impact of the interface on software projects in most languages, in the design of Go's orthogonal interface. It is even possible to remove dependencies; the second is that the compiler helps us to check for errors like "not fully implemented interfaces" at compile time, if the business does not implement a process, but uses its instance as an interface forcibly. :

```go
Package main

Type OrderCreator interface {
ValidateUser()
CreateOrder()
}

Type BookOrderCreator struct{}

Func (boc BookOrderCreator) ValidateUser() {}

Func createOrder(oc OrderCreator) {
oc.ValidateUser()
oc.CreateOrder()
}

Func main() {
createOrder(BookOrderCreator{})
}
```

The following error will be reported.

```shell
# command-line-arguments
./a.go:18:30: cannot use BookOrderCreator literal (type BookOrderCreator) as type OrderCreator in argument to createOrder:
BookOrderCreator does not implement OrderCreator (missing CreateOrder method)
```

Therefore, the interface can also be considered as a type-safe means of checking at compile time.

## 5.8.5 Table Driven Development

Students who are familiar with open source lint tools should have seen the complexity of the circle. If there are `if` and `switch` in the function, the complexity of the function will increase, so students with obsessive-compulsive disorder have a function at the entrance. There is `switch` in it, or do you want to kill this `switch`, is there any way? Of course, there are table-driven ways to store the examples we need:

```go
Func entry() {
Var bi BusinessInstance
Switch businessType {
Case TravelBusiness:
Bi = travelorder.New()
Case MarketBusiness:
Bi = marketorder.New()
Default:
Return errors.New("not supported business")
}
}
```

Can be modified to:

```go
Var businessInstanceMap = map[int]BusinessInstance {
TravelBusiness : travelorder.New(),
MarketBusiness : marketorder.New(),
}

Func entry() {
Bi := businessInstanceMap[businessType]
}
```

Table-driven design, many design-related books do not use it as a design pattern, but I think this is still a very important means to help us simplify the code. In the daily development work, you can think more about which unnecessary `switch case` can be easily solved with a dictionary and a line of code.

Of course, table-driven is not a disadvantage, because you need to calculate the hash of the input `key`, in the case of performance-sensitive, you need to consider more.