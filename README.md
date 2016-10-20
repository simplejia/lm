# [lm](http://github.com/simplejia/lm) (lc+redis+[mysql|http] glue)
## 实现初衷
> 写redis+mysql代码时（还可能加上lc），示意代码如下：
```
func orig(key string) (value string) {
    value = redis.Get(key)
    if value != "" {
        return
    }
    value = mysql.Get(key)
    redis.Set(key, value)
    return
}

// 如果再加上lc的话
func orig(key string) (value string) {
    value = lc.Get(key)
    if value != "" {
        return
    }
    value = redis.Get(key)
    if value != "" {
        lc.Set(key, value)
        return
    }
    value = mysql.Get(key)
    redis.Set(key, value)
    lc.Set(key, value)
    return
}
```
> 有了lm，再写上面的代码时，一切变的那么简单
[lm_test.go](http://github.com/simplejia/lm/lm_test.go)
```
func tGlue(key, value string) (err error) {
	err = Glue(
		key,
		&value,
		func(p, r interface{}) error {
            _r := r.(*string)
            *_r = "test value"
			return nil
		},
		func(p interface{}) string {
			return fmt.Sprintf("tGlue:%v", p)
		},
		&LcStru{
			Expire: time.Millisecond * 500,
			Safety: false,
		},
		&McStru{
			Expire: time.Minute,
            Pool: pool,
		},
	)
	if err != nil {
		return
	}
    return
}
```
## 功能
* 自动添加缓存代码，支持lc, redis，减轻你的心智负担，让你的代码更加简单可靠，少了大段的冗余代码，复杂的事全交给lm自动帮你做了
* 支持Glue[Lc|Mc]及相应批量操作Glues[Lc|Mc]，详见lm_test.go示例代码

## 注意
* lm.LcStru.Safety，当置为true时，对lc在并发状态下返回的nil值不接受，因为lc.Get在并发状态下，同一个key返回的value有可能是nil，并且ok状态为true，Safety置为true后，对以上情况不接受，会继续调用下一层逻辑，

## LICENSE
lm is licensed under the Apache Licence, Version 2.0
(http://www.apache.org/licenses/LICENSE-2.0.html)
