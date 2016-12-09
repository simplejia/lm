[中文 README](#中文)


# [lm](http://github.com/simplejia/lm) (lc+redis+[mysql|http] glue)
## Original Intention
> When coding with redis and mysql(maybe also add lc cache support), as following:
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
// add lc cache support
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
> Having lm, the code above will be very easy:
[lm_test.go](http://github.com/simplejia/lm/tree/master/lm_test.go)
```
func tGlue(key, value string) (err error) {
	lmStru := &LmStru{
		Input:  key,
		Output: &value,
        Proc: func(p, r interface{}) error {
            _r := r.(*string)
            *_r = "test value"
			return nil
		},
        Key: func(p interface{}) string {
			return fmt.Sprintf("tGlue:%v", p)
		},
		Mc: &McStru{
			Expire: time.Minute,
			Pool:   pool,
		},
		Lc: &LcStru{
			Expire: time.Millisecond * 500,
			Safety: false,
		},
	}
	err = Glue(lmStru)
	if err != nil {
		return
	}
    return
}
```
## Features
* It can automatically add cache feature and support lc and redis, which will make your code more simpler and reliable and reduce large segment of redundant code.
* It supports Glue[Lc|Mc] and corresponding batch operation(Glues[Lc|Mc]), for the details, please refer to the instance code of lm_test.go

## Notice
* lm.LcStru.Safety parameter，When it is true, the nil value returned by the LC under concurrent state is not accepted, because when lc.Get is under concurrent state, the returned value by the same key maybe nil but ok parameter is true. When safety is set as true, all the conditions above will not be accepted. It will continue the next logic.

---
中文
===

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
[lm_test.go](http://github.com/simplejia/lm/tree/master/lm_test.go)
```
func tGlue(key, value string) (err error) {
	lmStru := &LmStru{
		Input:  key,
		Output: &value,
        Proc: func(p, r interface{}) error {
            _r := r.(*string)
            *_r = "test value"
			return nil
		},
        Key: func(p interface{}) string {
			return fmt.Sprintf("tGlue:%v", p)
		},
		Mc: &McStru{
			Expire: time.Minute,
			Pool:   pool,
		},
		Lc: &LcStru{
			Expire: time.Millisecond * 500,
			Safety: false,
		},
	}
	err = Glue(lmStru)
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
* lm.LcStru.Safety，当置为true时，对lc在并发状态下返回的nil值不接受，因为lc.Get在并发状态下，同一个key返回的value有可能是nil，并且ok状态为true，Safety置为true后，对以上情况不接受，会继续调用下一层逻辑

## 案例分享
* 一天一个用户只容许投一次票
```
func f(uid string) (err error) {
	lmStru := &lm.LmStru{
		Input: uid,
		Output: &struct{}{},
        Proc: func(p, r interface{}) error {
			// 略掉这部分逻辑: 可以把投票入库
			// ...
			return nil
		},
        Key: func(p interface{}) string {
			return fmt.Sprintf("pkg:f:%v", p)
		},
		Mc: &lm.McStru{
			Expire: time.Hour * 24,
			Pool:   pool,
		},
	}
	err = lm.GlueMc(lmStru)
	if err != nil {
		return
	}
    return
}
```

