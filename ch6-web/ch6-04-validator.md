# 6.4. validator 请求校验

社区里曾经有人用这张图来嘲笑 PHP：

![validate 流程](../images/ch6-04-validate.jpg)

实际上这是一个语言无关的场景，需要进行字段校验的情况有很多，web 系统的 Form/json 提交只是一个典型的例子。我们用 go 来写一个类似上图的校验 demo。然后研究怎么一步步对其进行改进。

## 重构请求校验函数

假设我们的数据已经通过某个 binding 库绑定到了具体的 struct 上。

```go
type RegisterReq struct {
    Username        string   `json:"username"`
    PasswordNew     string   `json:"password_new"`
    PasswordRepeat  string   `json:"password_repeat"`
    Email           string   `json:"email"`
}

func register(req RegisterReq) error{
    if len(req.Username) > 0 {
        if len(req.PasswordNew) > 0 && len(req.PasswordRepeat) > 0 {
            if req.PasswordNew == req.PasswordRepeat {
                if emailFormatValid(req.Email) {
                    createUser()
                    return nil
                } else {
                    return errors.New("invalid email")
                }
            } else {
                return errors.New("password and reinput must be the same")
            }
        } else {
            return errors.New("password and password reinput must be longer than 0")
        }
    } else {
        return errors.New("length of username cannot be 0")
    }
}
```

我们在 golang 里成功写出了 hadoken 开路的箭头型代码。。这种代码一般怎么进行优化呢？

很简单，在《重构》一书中已经给出了方案：[Guard Clauses](https://refactoring.com/catalog/replaceNestedConditionalWithGuardClauses.html)。

```go
func register(req RegisterReq) error{
    if len(req.Username) == 0 {
        return errors.New("length of username cannot be 0")
    }

    if len(req.PasswordNew) == 0 || len(req.PasswordRepeat) == 0 {
        return errors.New("password and password reinput must be longer than 0")
    }

    if req.PasswordNew != req.PasswordRepeat {
        return errors.New("password and reinput must be the same")
    }

    if emailFormatValid(req.Email) {
        return errors.New("invalid email")
    }

    createUser()
    return nil
}
```

代码更清爽，看起来也不那么别扭了。这是比较通用的重构理念。虽然使用了重构方法使我们的 validate 过程看起来优雅了，但我们还是得为每一个 http 请求都去写这么一套差不多的 validate 函数，有没有更好的办法来帮助我们解除这项体力劳动？答案就是 validator。

## 用 validator 解放体力劳动
