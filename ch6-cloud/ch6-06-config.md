# 6.6 Distributed Configuration Management

In distributed systems, there are often problems that are bothering us. Although there are some elegant restart schemes at present, the actual application may be limited by the internal operation of our system and there is no way to achieve true "elegance". For example, in order to limit the flow to downstream, some data is accumulated in the memory, and the threshold of the time or total amount is set for the accumulation. After the arbitrary threshold is reached, the data is sent to the downstream uniformly, so as to avoid frequent requests exceeding the downstream carrying capacity and smashing downstream. In this case, it is more difficult to restart to be elegant.

Therefore, our goal is to avoid adopting or bypassing the online method and make some modifications to the online program. A typical modification is the configuration item of the program.

## 6.6.1 Scene Example

### 6.6.1.1 Reporting System

In some OLAP or offline data platforms, after a long period of iterative development, the functional modules of the entire system have gradually stabilized. Variable items only appear in the data layer, and most of the changes in the data layer can be considered as changes in SQL. Architects naturally think about pulling these changes out of the system. For example, the configuration management system described in this section.

When the business puts forward new requirements, our requirement is to enter new SQL into the system, or simply modify the old SQL. These changes can be done directly without going online.

### 6.6.1.2 Business Configuration

The platform department of a large company serves a number of business lines, and each business line is assigned a unique id within the platform. The platform itself is also made up of multiple modules that need to share the same line of business definition (or else it's messed up). When the company opens a new product line, it needs to be able to get through the process of all platform systems in a short time. At this time, it is definitely too late for each system to go online. In addition, this kind of public configuration needs to be managed in a unified manner, and its addition and subtraction logic is also managed in a unified manner. When this information is changed, it is necessary to automatically notify the business party's system without human intervention (or only a very simple intervention, such as click auditing).

In addition to line-of-business management, many Internet companies spread their business in accordance with the city. Before a city is opened, all modules in theory should think that the data with the city id is dirty data and is automatically filtered out. If the business is opened, the new city id should be automatically added to the whitelist in the system. This way the business process can run automatically.

For another example, there are various types of operational activities in the operating system of an Internet company. Some operational activities may have unexpected events (such as public relations crisis), and the system needs to be urgently taken offline. At this time, some switches will be used to quickly turn off the corresponding functions. Or quickly remove the activity id you want to cull from the whitelist. In the AB Test section of the Web chapter, we also mentioned that sometimes there is a need to have such a system to tell us how much traffic is currently required to be put on the corresponding function code. We can use remote RPC to learn this information in the same section, but at the same time, we can actively pull this information in conjunction with the distributed configuration system.

## 6.6.2 Using etcd to implement configuration updates

We use etcd to implement a simple configuration read and dynamic update process to understand the online configuration update process.

### 6.6.2.1 Configuration Definition

Simple configuration, you can store the content completely in etcd. such as:

```shell
Etcdctl get /configs/remote_config.json
{
"addr" : "127.0.0.1:1080",
"aes_key" : "01B345B7A9ABC00F0123456789ABCDAF",
"https" : false,
"secret" : "",
"private_key_path" : "",
"cert_file_path" : ""
}
```

### 6.6.2.2 Creating a new etcd client

```go
Cfg := client.Config{
Endpoints: []string{"http://127.0.0.1:2379"},
Transport: client.DefaultTransport,
HeaderTimeoutPerRequest: time.Second,
}
```

Directly initialized with the structure in the etcd client package, nothing to say.

### 6.6.2.3 Configuration Acquisition

```go
Resp, err = kapi.Get(context.Background(), "/path/to/your/config", nil)
If err != nil {
log.Fatal(err)
} else {
log.Printf("Get is done. Metadata is %q\n", resp)
log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)
}
```

Obtaining the `Get()` method that uses the etcd KeysAPI configuration is relatively simple.

### 6.6.2.4 Configuring an update subscription

```go
Kapi := client.NewKeysAPI(c)
w := kapi.Watcher("/path/to/your/config", nil)
Go func() {
For {
Resp, err := w.Next(context.Background())
log.Println(resp, err)
log.Println("new values ​​is ", resp.Node.Value)
}
}()
```

By subscribing to the change event of the config path, when the content changes in the path, the client side can receive the change notification and receive the changed string value.

### 6.6.2.5 Integrate

```go
Package main

Import (
"log"
"time"

"golang.org/x/net/context"
"github.com/coreos/etcd/client"
)

Var configPath = `/configs/remote_config.json`
Var kapi client.KeysAPI

Type ConfigStruct struct {
Addr string `json:"addr"`
AesKey string `json:"aes_key"`
HTTPS bool `json:"https"`
Secret string `json:"secret"`
PrivateKeyPath string `json:"private_key_path"`
CertFilePath string `json:"cert_file_path"`
}

Var appConfig ConfigStruct

Func init() {
Cfg := client.Config{
Endpoints: []string{"http://127.0.0.1:2379"},
Transport: client.DefaultTransport,
HeaderTimeoutPerRequest: time.Second,
}

c, err := client.New(cfg)
If err != nil {
log.Fatal(err)
}
Kapi = client.NewKeysAPI(c)
initConfig()
}

Func watchAndUpdate() {
w := kapi.Watcher(configPath, nil)
Go func() {
// watch every change under this node
For {
Resp, err := w.Next(context.Background())
If err != nil {
log.Fatal(err)
}
log.Println("new values ​​is ", resp.Node.Value)

Err = json.Unmarshal([]byte(resp.Node.Value), &appConfig)
If err != nil {
log.Fatal(err)
}
}
}()
}

Func initConfig() {
Resp, err = kapi.Get(context.Background(), configPath, nil)
If err != nil {
log.Fatal(err)
}

Err := json.Unmarshal(resp.Node.Value, &appConfig)
If err != nil {
log.Fatal(err)
}
}

Func getConfig() ConfigStruct {
Return appConfig
}

Func main() {
// init your app
}
```

If the business is small, use the examples in this section to implement the functionality.

Just to note here, we have a series of operations when updating the configuration: watch response, json parsing, these operations are not atomic. When the config is acquired multiple times in a single service request process, there may be a logical inconsistency between the individual requests before and after the config changes in the middle. Therefore, when you use a similar approach to update your configuration, you need to use the same configuration for the lifetime of a single request. The specific implementation manner may be that the configuration is obtained only once when the request starts, and then transparently transmitted downwards in sequence, etc., and the specific situation is specifically analyzed.

## 6.6.3 Configuring bloat

As the business grows, the pressure on the configuration system itself may become larger and larger, and the configuration files may be tens of thousands. The client is also tens of thousands, and storing the configuration content inside etcd is no longer appropriate. As the number of configuration files expands, in addition to the throughput problems of the storage system itself, there are management issues for configuration information. We need to manage the rights of the corresponding configuration, and we need to configure the storage cluster according to the traffic. If there are too many clients, which causes the configuration storage system to be unable to withstand a large amount of QPS, it may also need to perform cache optimization on the client side, and so on.

That's why big companies have to develop a complex configuration system for their business.

## 6.6.4 Configuring version management

In the configuration management process, it is inevitable that the user may misuse the operation. For example, when updating the configuration, a configuration that cannot be resolved is input. In this case we can solve by configuring the checksum.

Sometimes the wrong configuration may not be a problem with the format, but a logical problem. For example, when we write SQL, we select a field less. When we update the configuration, we accidentally drop a field in the json string and cause the program to understand the new configuration and enter the weird logic. The fastest and most effective way to stop losses quickly is to version management and support rollback by version.

When the configuration is updated, we will assign a version number to each new content of the configuration, record the content and version number before the modification, and roll back in time when it finds a problem with the new configuration.

A common practice is to use MySQL to store different versions of the configuration file or configuration string. When you need to roll back, just do a simple query.

## 6.6.5 Client Fault Tolerance

After the configuration of the business system is stripped to the configuration center, it does not mean that our system can sit back and relax. When the configuration center itself is down, we also need some fault tolerance, at least to ensure that the business can still operate during its downtime. This requires our system to get the required configuration information when the configuration center is down. Even if this information is not new enough.

Specifically, when providing a configuration read SDK for a service, it is preferable to cache the obtained configuration on the disk of the business machine. When the remote configuration center is not available, you can directly use the contents of the hard disk to do the bottom. When reconnecting the configuration center, the corresponding content is updated.

Be sure to consider the data consistency problem after adding the cache. When individual business machines are inconsistent with other machine configurations due to network errors, we should also be able to know from the monitoring system.

We use a means to solve the pain points of our configuration update, but at the same time we may bring us new problems because of the means used. In actual development, we have to think a lot about each step of the decision so that we are not at a loss when the problem comes.