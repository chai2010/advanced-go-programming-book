# 6.4 Distributed Search Engine

In the Web chapter, we mentioned that MySQL is very fragile. The database system itself is guaranteed to be real-time and strongly consistent, so its functionality is designed to meet this consistency requirement. For example, the design of write ahead log, the index and data organization based on B+ tree, and the transaction based on MVCC.

Relational databases are generally used to implement OLTP systems, called OLTP, citing wikipedia:

> Online transaction processing (OLTP) refers to the processing of general and immediate job data by means of information systems, computer networks and databases, and the operation of large batches of earlier traditional database systems. the same. OLTP is often used for automated data processing tasks such as order entry, financial transactions, etc. In contrast to online analytical processing (OLAP), which is at the decision analysis level.

In the business scenario of the Internet, there are also scenarios where the real-time requirements are not high (a delay of many seconds can be accepted), but the query complexity is high. For example, in an e-commerce WMS system, or in most CRM or customer service systems with rich business scenarios, it may be necessary to provide random combination query functions for dozens of fields. The data dimensions of such a system are inherently numerous, such as a description of a piece of goods in an e-commerce WMS, which may have the following fields:

> Warehouse id, warehousing time, location id, storage shelf id, warehousing operator id, outbound operator id, inventory quantity, expiration time, SKU type, product brand, product category, number of internals

In addition to the above information, if the goods are circulating in the warehouse. There may be associated process ids, current flow status, and so on.

Imagine if we were running a large e-commerce company with tens of millions of orders per day, it would be very difficult to query and build the appropriate index in this database.

In CRM or customer service systems, there is often a need to search by keyword, and large Internet companies receive tens of thousands of user complaints every day. Considering the source of the incident, the user's complaint must be at least 2 to 3 years. It is also tens of millions or even hundreds of millions of data. Performing a like query based on the keyword may hang the entire MySQL directly.

At this time we need a search engine to save the game.

## search engine

Elasticsearch is the leader of the open source distributed search engine, which relies on the Lucene implementation and has made many optimizations in deployment and operation and maintenance. Building a distributed search engine today is much easier than the Sphinx era. Simply configure the client IP and port.

### Inverted list

Although es is customized for the search scenario, as mentioned earlier, es is often used as a database in practical applications because of the nature of the inverted list. You can understand the inverted index with a simpler perspective:

![posting-list](../images/ch6-posting_list.png)

*Figure 6-10 Inverted list*

When querying data in Elasticsearch, the essence is to find a sequence of multiple ordered sequences. Non-numeric type fields involve word segmentation problems. In most internal usage scenarios, we can use the default bi-gram word segmentation directly. What is a bi-gram participle:

Putting all `Ti` and `T(i+1)` into one word (called term in Elasticsearch), and then rearranging its inverted list, so our inverted list is probably like this:

![terms](../images/ch6-terms.png)

*Figure 6-11 The result of the word segmentation "Good weather today"*

When the user searches for 'the weather is good', it is actually seeking: the weather, the gas is very good, the intersection of the three groups of inverted lists, but the equality judgment logic here is somewhat special, with pseudo code:

```go
Func equal() {
If postEntry.docID of 'weather' == postEntry.docID of '气很' &&
postEntry.offset + 1 of 'weather' == postEntry.offset of '气很' {
Return true
}

If postEntry.docID of '气很' == postEntry.docID of 'very good' &&
postEntry.offset + 1 of '气很' == postEntry.offset of '好' {
Return true
}

If postEntry.docID of 'weather' == postEntry.docID of 'very good' &&
postEntry.offset + 2 of 'weather' == postEntry.offset of 'very good' {
Return true
}

Return false
}
```

The time complexity of multiple ordered lists to find intersections is: `O(N * M)`, where N is the smallest set of elements in a given list, and M is the number of given lists.

One of the decisive factors in the whole algorithm is the length of the shortest inverted list, followed by the sum of words, and the general number of words is not very large (Imagine, would you type hundreds of words in the search engine to search?) , so the decisive role is generally the length of the shortest one of all the inverted list.

Therefore, in the case where the total number of documents is large, the search speed is also very fast when the shortest one of the inverted lists of search terms is not long. If you use a relational database, you need to scan slowly by index (if any).

### Query DSL

Es defines a set of query DSL, when we use es as a database, we need to use its bool query. for example:

```json
{
  "query": {
    "bool": {
      "must": [
        {
          "match": {
            "field_1": {
              "query": "1",
              "type": "phrase"
            }
          }
        },
        {
          "match": {
            "field_2": {
              "query": "2",
              "type": "phrase"
            }
          }
        },
        {
          "match": {
            "field_3": {
              "query": "3",
              "type": "phrase"
            }
          }
        },
        {
          "match": {
            "field_4": {
              "query": "4",
              "type": "phrase"
            }
          }
        }
      ]
    }
  },
  "from": 0,
  "size": 1
}
```

It seems cumbersome, but the meaning of the expression is simple:

```go
If field_1 == 1 && field_2 == 2 && field_3 == 3 && field_4 == 4 {
    Return true
}
```

Use bool should query to represent the logic of or:

```json
{
  "query": {
    "bool": {
      "should": [
        {
          "match": {
            "field_1": {
              "query": "1",
              "type": "phrase"
            }
          }
        },
        {
          "match": {
            "field_2": {
              "query": "3",
              "type": "phrase"
            }
          }
        }
      ]
    }
  },
  "from": 0,
  "size": 1
}
```

What is shown here is similar:

```go
If field_1 == 1 || field_2 == 2 {
Return true
}
```

The expressions followed by `if` in these Go codes have proper nouns in the programming language to express `Boolean Expression`:

```go
4 > 1
5 == 2
3 < i && x > 10
```

The `Bool Query` scheme of es uses json to express the Boolean Expression in this programming language. Why can this be done? Because json itself can express the tree structure, our program code will become AST after being parse by the compiler, and the AST abstract syntax tree, as its name implies, is a tree structure. In theory, json can fully express the result of a piece of program code being parse. The Boolean Expression here is also generated by the compiler Parse and generates a similar tree structure, and is only a small subset of the entire compiler implementation.

### Based on client SDK for development

initialization:

```go
// When using the elastic version
// Note that it corresponds to the elasticsearch you use.
Import (
Elastic "gopkg.in/olivere/elastic.v3"
)

Var esClient *elastic.Client

Func initElasticsearchClient(host string, port string) {
Var err error
esClient, err = elastic.NewClient(
elastic.SetURL(fmt.Sprintf("http://%s:%s", host, port)),
elastic.SetMaxRetries(3),
)

If err != nil {
// log error
}
}
```

insert:

```go
Func insertDocument(db string, table string, obj map[string]interface{}) {

Id := obj["id"]

Var indexName, typeName string
// The database/table concept in the database can be simply mapped to the index and type of es
// but note, because the _type in es is essentially just a field of document
// So too much single index content can cause performance issues
// in the new version type has been deprecated
// In order to make the data of different tables fall into different indexes, here we use table+name as the name of index
indexName = fmt.Sprintf("%v_%v", db, table)
typeName = table

	// normal circumstances
Res, err := esClient.Index().Index(indexName).Type(typeName).Id(id).BodyJson(obj).Do()
If err != nil {
// handle error
} else {
// insert success
}
}
```

Obtain:

```go
Func query(indexName string, typeName string) (*elastic.SearchResult, error) {
// Add bool query condition via bool must and bool should
q := elastic.NewBoolQuery().Must(elastic.NewMatchPhraseQuery("id",1),
elastic.NewBoolQuery().Must(elastic.NewMatchPhraseQuery("male", "m")))

q = q.Should(
elastic.NewMatchPhraseQuery("name", "alex"),
elastic.NewMatchPhraseQuery("name", "xargin"),
)

searchService := esClient.Search(indexName).Type(typeName)
Res, err := searchService.Query(q).Do()
If err != nil {
// log error
Return nil, err
}

Return res, nil
}
```

delete:

```go
Func deleteDocument(
indexName string, typeName string, obj map[string]interface{},
) {
Id := obj["id"]

Res, err := esClient.Delete().Index(indexName).Type(typeName).Id(id).Do()
If err != nil {
// handle error
} else {
// delete success
}
}
```

Because of the nature of Lucene, the data in the search engine is essentially immutable, so if you want to update the document, Lucene internally performs full coverage according to the id (essentially the data in the latest segment of the same id), so The same is true for the insertion.

When using es as a database, you need to be aware that because es has an operation of index merging, it takes a while for the data to be inserted into es to be queried (determined by ref's refresh_interval). So don't use es as a strong and consistent relational database.

### Convert sql to DSL

For example, we have a bool expression, `user_id = 1 and (product_id = 1 and (star_num = 4 or star_num = 5) and banned = 1)`, written in SQL as follows:

```sql
Select * from xxx where user_id = 1 and (
Product_id = 1 and (star_num = 4 or star_num = 5) and banned = 1
)
```

The DSL written as es is of the form:

```json
{
  "query": {
    "bool": {
      "must": [
        {
          "match": {
            "user_id": {
              "query": "1",
              "type": "phrase"
            }
          }
        },
        {
          "match": {
            "product_id": {
              "query": "1",
              "type": "phrase"
            }
          }
        },
        {
          "bool": {
            "should": [
              {
                "match": {
                  "star_num": {
                    "query": "4",
                    "type": "phrase"
                  }
                }
              },
              {
                "match": {
                  "star_num": {
                    "query": "5",
                    "type": "phrase"
                  }
                }
              }
            ]
          }
        },
        {
          "match": {
            "banned": {
              "query": "1",
              "type": "phrase"
            }
          }
        }
      ]
    }
  },
  "from": 0,
  "size": 1
}
```

Although es DSL is well understood, it is very hard to write. The SDK-based approach was provided earlier, but it is not flexible enough.

The where part of SQL is the boolean expression. As we mentioned before, this bool expression is similar to the es DSL structure after being parsed. Can we directly help us convert SQL to DSL through this "almost" guess?

Of course, we can compare the structure of SQL with the structure of Parse and the structure of es DSL:

![ast](../images/ch6-ast-dsl.png)

*Figure 6-12 Correspondence between AST and DSL*

Since the structure is completely consistent, we can logically convert each other. We traverse the AST tree with breadth first, then convert the binary expression into a json string, and then assemble it. Due to space limitations, no examples are given in this article. Readers can check:

> github.com/cch123/elasticsql

To learn the specific implementation.

## Heterogeneous data synchronization

In practical applications, we rarely write data directly to search engines. A more common way is to synchronize data from MySQL or other relational data into a search engine. The user of the search engine can only query the data and cannot modify and delete it.

There are two common synchronization schemes:

### Incremental data synchronization via timestamp

![sync to es](../images/ch6-sync.png)

*Figure 6-13 Time-based data synchronization*

This synchronization method is strongly bound to the business, such as the outbound order in the WMS system, we do not need very real time, a little delay is acceptable, then we can get the last ten from the MySQL outbound order table every minute. All the outbound orders created in minutes are taken out and stored in es in batches. The logic that needs to be executed for the operations of fetching data can be expressed as the following SQL:

```sql
Select * from wms_orders where update_time >= date_sub(now(), interval 10 minute);
```

Of course, considering the boundary situation, we can make the data of this time period overlap with the previous one:

```sql
Select * from wms_orders where update_time >= date_sub(
Now(), interval 11 minute
);
```

Update the data coverage changed to es in the last 11 minutes. The shortcomings of this approach are obvious, and we must require business data to strictly adhere to certain specifications. For example, there must be an update_time field, and each time it is created and updated, the field must have the correct time value. Otherwise our synchronization logic will lose data.

### Synchronizing data with binlog

![binlog-sync](../images/ch6-binlog-sync.png)

*Figure 6-13 Data synchronization based on binlog*

The industry uses more of Ali's open source Canal for binlog parsing and synchronization. Canal will pretend to be a MySQL slave library, then parse the bincode of the line format and send it to the message queue in a more easily parsed format (such as json).

The downstream Kafka consumer is responsible for writing the self-incrementing primary key of the upstream data table as the id of the es document, so that each time the binlog is received, the corresponding id data is updated to the latest. MySQL's Row format binlog will provide all the fields of each record to the downstream, so when synchronizing data to heterogeneous data targets, you don't need to consider whether the data is inserted or updated, as long as you always cover by id.

This model also requires the business to adhere to a data table specification, that is, the table must have a unique primary key id to ensure that the data we enter into es will not be duplicated. Once the specification is not followed, it will result in data duplication when synchronizing. Of course, you can also customize the consumer logic for each required table, which is not the scope of the general system discussion.