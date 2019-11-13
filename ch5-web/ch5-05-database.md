# 5.5 Database and database dealings

This section will do some simple analysis of the `db/sql` official standard library, and introduce some widely used open source ORM and SQL Builder. And from the perspective of enterprise application development and corporate architecture, it is more appropriate to analyze which technology stack is suitable for modern enterprise applications.

## 5.5.1 Speaking from database/sql

Go officially provides the `database/sql` package to work with the database. The `database/sql` library actually only provides a set of interfaces and specifications for operating the database, such as abstract SQL prep (prepare). Connection pool management, data binding, transactions, error handling, and more. The official does not provide specific protocol support for certain database implementations.

To deal with a specific database, such as MySQL, you need to introduce the MySQL driver again, like this:

```go
Import "database/sql"
Import _ "github.com/go-sql-driver/mysql"

Db, err := sql.Open("mysql", "user:password@/dbname")
```

```go
Import _ "github.com/go-sql-driver/mysql"
```

This import statement will call the `init` function of the `mysql` package, and the things it does are simple:

```go
Func init() {
sql.Register("mysql", &MySQLDriver{})
}
```

Register `driver` with the name `mysql` in the global `map` of the `sql` package. `Driver` is an interface in the `sql` package:

```go
Type Driver interface {
Open(name string) (Conn, error)
}
```

The `db` object returned by calling `sql.Open()` is the `Conn` here.

```go
Type Conn interface {
Prepare(query string) (Stmt, error)
Close() error
Begin() (Tx, error)
}
```

It is also an interface. If you look closely at the `database/sql/driver/driver.go` code, you will find that all the members in this file are all interfaces. If you operate on these types, you will still call the specific `driver` method.

From the user's point of view, in the process of using the `database/sql` package, you can use the functions provided in these interfaces. Let's look at a complete example using `database/sql` and `go-sql-driver/mysql`:

```go
Package main

Import (
"database/sql"
_ "github.com/go-sql-driver/mysql"
)

Func main() {
// db is an object of type sql.DB
// The object is thread safe and contains a connection pool internally
// The option to connect to the pool can be set in the sql.DB method, which is omitted here for simplicity.
Db, err := sql.Open("mysql",
"user:password@tcp(127.0.0.1:3306)/hello")
If err != nil {
log.Fatal(err)
}
Defer db.Close()

Var (
Id int
Name string
)
Rows, err := db.Query("select id, name from users where id = ?", 1)
If err != nil {
log.Fatal(err)
}

Defer rows.Close()

/ / must read the contents of the rows, or explicitly call the Close () method,
// Otherwise the connection will never be released until defer's rows.Close() is executed
For rows.Next() {
Err := rows.Scan(&id, &name)
If err != nil {
log.Fatal(err)
}
log.Println(id, name)
}

Err = rows.Err()
If err != nil {
log.Fatal(err)
}
}
```

If you want to know more about the official usage of the `database/sql` library, you can refer to `http://go-database-sql.org/`.

Including the function introduction, usage, precautions and anti-intuitive implementation of the library (for example, the query of `sql.DB` in the same goroutine, may be on multiple connections) is involved, and will not be repeated in this chapter.

As smart as you are, you may have sniffed some bad tastes in the short procedure above. The function provided by the official `db` library is so simple. Do we have to write such a similar code every time we go to the database to read the content? Or if our object is a struct, the work of binding `sql.Rows` to the object becomes more repetitive and boring.

Yes, so the community will have a variety of SQL Builder and ORM.

## 5.5.2 ORM and SQL Builder to improve production efficiency

What is the ORM often mentioned in the field of web development? Let's take a look at the omnipotent Wikipedia:

```
Object Relational Mapping (English: Object Relational Mapping, ORM, or O/RM, or O/R mapping),
A programming technique used to implement the conversion of data between different types of systems in an object-oriented programming language.
In effect, it actually creates a "virtual object database" that can be used in programming languages.
```

The most common ORM is doing a mapping from db to a program's class or struct. So the program at hand may be mapping the classes inside your program from the MySQL table. Let's first take a look at how the ORM in other programming languages ​​is written:

```python
>>> from blog.models import Blog
>>> b = Blog(name='Beatles Blog', tagline='All the latest Beatles news.')
>>> b.save()
```

There is no trace of the database at all. Yes, the purpose of the ORM is to shield the DB layer. The ORM of many languages ​​only defines your class or structure, and then uses a specific syntax to make one-to-one or a pair between the structures. Many relationships are expressed. Then the task is complete. Then you can perform various operations on these objects that map the database tables, such as save, create, retrieve, and delete. As for what insidious things the ORM has done in the back, you are not necessarily clear. When using ORM, we tend to have an intuitive feeling of forgetting the database. For example, we have a need: to show users the latest list of products, we assume that the goods and businesses are 1:1 relationship, we can easily write code like this:

```python
# Fake code
shopList := []
For product in productList {
shopList = append(shopList, product.GetShop)
}
```

Of course, we can't criticize programmers who write code like this as lazy programmers. Because the tool like ORM is to shield sql from the starting point, let us operate the database closer to the human way of thinking. So many programmers who have only touched the ORM and are just entering the line can easily write the above code.

Such code will magnify the database read request by a factor of N. In other words, if your product list has 15 SKUs, each time the user opens the page, at least 1 (query item list) + 15 (query related shop information) query is required. Here N is 16. If your list page is large, say 600 entries, then you must perform at least 1+600 queries. If the maximum simple query that your database can withstand is 120,000 QPS, and the above query is just your most commonly used query, what is the service capacity that you can provide externally? It is 200 qps! One of the taboos of the Internet system is this unwarranted read amplification.

Of course, you can also say that this is not an ORM problem. If you write sql you may still write a similar program, then look at two demos:

```go
o := orm.NewOrm()
Num, err := o.QueryTable("cardgroup").Filter("Cards__Card__Name", cardName).All(&cardgroups)
```

Many ORMs provide this type of Filter query, but behind some ORMs may hide very difficult details, such as the generated SQL statement will automatically `limit 1000`.

Perhaps readers who like ORM will refute it when they read it. You did not read the document carefully. Yes, although these ORM tools show in the documentation that All queries automatically limit 1000 without explicitly specifying Limit, this is still very difficult for many people who have not read the documentation or read the ORM source. "Devil" details. People who like strong typing languages ​​generally don't like what the language implicitly does, such as the implicit type conversion of various languages ​​in the assignment operation and then lose the precision in the conversion, which will definitely cause you a headache. . So the less things a library does in the backside, the better. If you must do it, you must do it in a conspicuous place. For example, in the above example, it is better to remove this default self-acting behavior or to force the user to pass the limit parameter.

In addition to the limit issue, let's take a look at this query below:

```go
Num, err := o.QueryTable("cardgroup").Filter("Cards__Card__Name", cardName).All(&cardgroups)
```

Can you see that this Filter is a join operation? Of course, users who have in-depth experience will still feel that this is nitpicking. But such an analysis wants to prove that ORM wants to hide too much detail from the design. The price of convenience is that the operation behind it is completely out of control. Such a project will become unrecognizable and difficult to maintain after several maintenance personnel.

Of course, we can't deny the progressive significance of ORM. Its original intention is to separate the specific implementation of data operations and storage. But people at the scale of the company have gradually reached a consensus that ORM may be a failed design because of the hidden important details. The important details hidden are critical to the development of scaled systems.

Compared to ORM, SQL Builder achieves a good balance between SQL and project maintainability. First of all, sql builder does not block too much detail like ORM. Secondly, from the perspective of development, SQL Builder can also complete development very efficiently after simple encapsulation, for example:

```go
Where := map[string]interface{} {
"order_id > ?" : 0,
"customer_id != ?" : 0,
}
Limit := []int{0,100}
orderBy := []string{"id asc", "create_time desc"}

Orders := orderModel.GetList(where, limit, orderBy)
```

Write the relevant code of SQL Builder, or read it without any difficulty. Converting these code brains into sql is not too much effort. So through the code you can hit the database index on this query, whether to go over the index, whether it can be analyzed with the joint index.

To put it plainly, SQL Builder is a special dialect of sql in the code. If you don't have DBA, but R&D has the ability to analyze and optimize sql, or your company's DBA has no objection to learning such sql dialects. Then using SQL Builder is a good choice and will not cause any problems.

In addition, in some scenarios that do not require DBA intervention, it is also possible to use SQL Builder. For example, if you want to make a set of operation and maintenance system, and treat MySQL as a component in the system, the QPS of the system is not high, the query is not Complex and so on.

Once you're doing a high-concurrency OLTP online system, and you want to maximize the risk of your system with a clear division of labor, using SQL Builder is not appropriate.

## 5.5.3 Fragile database

Both ORM and SQL Builder have a fatal flaw, that is, there is no way to perform pre-sql audits on the system. Although many ORM and SQL Builder also provide the function of printing sql at runtime, it can only be output when querying. The functionality provided by SQL Builder and ORM itself is too flexible. It makes it impossible for you to enumerate all the sql that might be executed online by testing. For example, you might use SQL Builder to write the following code:

```go
Where := map[string]interface{} {
"product_id = ?" : 10,
"user_id = ?" : 1232 ,
}

If order_id != 0 {
Where["order_id = ?"] = order_id
}

Res, err := historyModel.GetList(where, limit, orderBy)
```

There are a lot of `if`s in your system like the above example, it is difficult to cover all possible SQL combinations through test cases.

Such a system, as long as it is released, has already given birth to a huge initial risk.

For Internet companies that are now 7 by 24 services, service unavailability is a very significant issue. Although the technology stack of the storage layer has undergone many years of development, it is still the most vulnerable part of the whole system. System downtime means direct economic loss for companies that provide services 24 hours a day. The risk can not be ignored.

From the perspective of industry division of labor, today's Internet companies have full-time DBAs. Most DBAs don't necessarily have the ability to write code. There are still a few obstacles to reading the SQL Builder's related "spelling SQL" code. From the DBA point of view, I still hope to have a special ex ante SQL audit mechanism, and let it get all the SQL content of the system at low cost, instead of reading the relevant code of SQL Builder written by the business development.

So nowadays, the core online business of large Internet companies will provide SQL in a prominent position in the code to the DBA review, for example:

```go
Const (
getAllByProductIDAndCustomerID = `select * from p_orders where product_id in (:product_id) and customer_id=:customer_id`
)

// GetAllByProductIDAndCustomerID
@param driver_id
@param rate_date
@return []Order, error
Func GetAllByProductIDAndCustomerID(ctx context.Context, productIDs []uint64, customerID uint64) ([]Order, error) {
Var orderList []Order

Params := map[string]interface{}{
"product_id" : productIDs,
"customer_id": customerID,
}

// getAllByProductIDAndCustomerID is a sql string of const type
Sql, args, err := sqlutil.Named(getAllByProductIDAndCustomerID, params)
If err != nil {
Return nil, err
}

Err = dao.QueryList(ctx, sqldbInstance, sql, args, &orderList)
If err != nil {
Return nil, err
}

Return orderList, err
}
```

For code like this, it is convenient to take the const part of the DAO layer's change set directly to the DBA for review before going online. The sqlutil.Named in the code is similar to the Named function in sqlx and supports the comparison operators and in in the where expression.

For the sake of simplicity, the function is written a little more complicated. If you think about it carefully, the exported function of the query can be further simplified. Please readers try it on their own.