// Package lm is used as wrapper for lc, redis, mysql etc...
// Created by simplejia [6/2016]
package lm

import (
	"bytes"
	"encoding/json"

	"github.com/garyburd/redigo/redis"
	"github.com/simplejia/lc"

	"reflect"
	"time"
)

type LmStru struct {
	Input  interface{}
	Output interface{}
	Proc   ft
	Key    mft
	Mc     *McStru
	Lc     *LcStru
}

type McStru struct {
	Expire time.Duration
	Pool   *redis.Pool
	Safety bool // true: nil值不保存
}

type LcStru struct {
	Expire time.Duration
	Safety bool // true: 对lc在并发状态下返回的nil值不接受
}

type ft func(p, r interface{}) error
type mft func(p interface{}) string

func GluesMc(lmStru *LmStru) (err error) {
	ps, result, f, mf, stru := lmStru.Input, lmStru.Output, lmStru.Proc, lmStru.Key, lmStru.Mc
	expire, pool, safety := stru.Expire, stru.Pool, stru.Safety

	rps := reflect.ValueOf(ps)
	num := rps.Len()
	if num == 0 {
		return
	}

	rresult := reflect.Indirect(reflect.ValueOf(result))
	if rresult.IsNil() {
		rresult = reflect.MakeMap(rresult.Type())
		reflect.Indirect(reflect.ValueOf(result)).Set(rresult)
	}

	keys := make([]interface{}, num)
	for i := 0; i < num; i++ {
		rp := rps.Index(i)
		p := rp.Interface()
		key := mf(p)
		keys[i] = key
	}

	conn := pool.Get()
	defer conn.Close()
	vs, err := redis.Strings(conn.Do("MGET", keys...))
	if err != nil {
		return
	}

	rpsNone := reflect.MakeSlice(reflect.TypeOf(ps), 0, 0)
	for i := 0; i < num; i++ {
		rp := rps.Index(i)
		if v := vs[i]; v != "" {
			var rppv reflect.Value
			var pvNew interface{}
			if re := reflect.TypeOf(result).Elem().Elem(); re.Kind() == reflect.Ptr {
				rppv = reflect.New(re.Elem())
				pvNew = rppv.Interface()
			} else {
				rppv = reflect.New(re)
				pvNew = rppv.Interface()
				rppv = reflect.Indirect(rppv)
			}
			json.Unmarshal([]byte(v), &pvNew)
			if pvNew == nil {
				continue
			}
			rresult.SetMapIndex(rp, rppv)
		} else {
			rpsNone = reflect.Append(rpsNone, rp)
			continue
		}
	}

	numNone := rpsNone.Len()
	if numNone == 0 {
		return
	}

	rresultPtrNone := reflect.New(rresult.Type())
	reflect.Indirect(rresultPtrNone).Set(reflect.MakeMap(rresult.Type()))
	rresultNone := reflect.Indirect(rresultPtrNone)
	err = f(rpsNone.Interface(), rresultPtrNone.Interface())
	if err != nil {
		return
	}

	for i := 0; i < rpsNone.Len(); i++ {
		rpNone := rpsNone.Index(i)
		pNone := rpNone.Interface()
		key4mc := mf(pNone)
		expire4mc := int(expire / time.Second)
		rv := rresultNone.MapIndex(rpNone)
		if rv.IsValid() {
			rresult.SetMapIndex(rpNone, rv)
			v, errIgnore := json.Marshal(rv.Interface())
			if errIgnore != nil {
				continue
			}
			conn.Do("SETEX", key4mc, expire4mc, v)
		} else if !safety {
			conn.Do("SETEX", key4mc, expire4mc, "null")
		}
	}

	return
}

func GlueMc(lmStru *LmStru) (err error) {
	p, result, f, mf, stru := lmStru.Input, lmStru.Output, lmStru.Proc, lmStru.Key, lmStru.Mc
	expire, pool, safety := stru.Expire, stru.Pool, stru.Safety

	key4mc := mf(p)
	expire4mc := int(expire / time.Second)

	conn := pool.Get()
	defer conn.Close()
	v, errIgnore := redis.String(conn.Do("GET", key4mc))
	if errIgnore == nil {
		err = json.Unmarshal([]byte(v), &result)
		return
	} else if errIgnore != redis.ErrNil {
		err = errIgnore
		return
	}

	err = f(p, result)
	if err != nil {
		return
	}

	vs, err := json.Marshal(result)
	if err != nil {
		return
	}

	if safety && bytes.Compare(vs, []byte("null")) == 0 {
		return
	}

	_, err = conn.Do("SETEX", key4mc, expire4mc, vs)
	if err != nil {
		return
	}

	return
}

func GluesLc(lmStru *LmStru) (err error) {
	ps, result, f, mf, stru := lmStru.Input, lmStru.Output, lmStru.Proc, lmStru.Key, lmStru.Lc
	expire, safety := stru.Expire, stru.Safety

	rps := reflect.ValueOf(ps)
	num := rps.Len()
	if num == 0 {
		return
	}

	rresult := reflect.Indirect(reflect.ValueOf(result))
	if rresult.IsNil() {
		rresult = reflect.MakeMap(rresult.Type())
		reflect.Indirect(reflect.ValueOf(result)).Set(rresult)
	}

	keys, keysM := make([]string, num), map[string]interface{}{}
	for i := 0; i < num; i++ {
		rp := rps.Index(i)
		p := rp.Interface()
		key := mf(p)
		keys[i] = key
		keysM[key] = p
	}

	vsLc, vsAlterLc := lc.Mget(keys)
	for k, v := range vsLc {
		if v == nil {
			if safety {
				delete(vsLc, k)
				vsAlterLc[k] = nil
			}
			continue
		}
		p := keysM[k]
		rresult.SetMapIndex(reflect.ValueOf(p), reflect.ValueOf(v))
	}

	numNone := num - len(vsLc)
	if numNone == 0 {
		return
	}

	rpsNone := reflect.MakeSlice(reflect.TypeOf(ps), 0, numNone)
	for i := 0; i < num; i++ {
		rp := rps.Index(i)
		p := rp.Interface()
		key := mf(p)
		if _, ok := vsAlterLc[key]; ok {
			rpsNone = reflect.Append(rpsNone, rp)
		}
	}

	rresultPtrNone := reflect.New(rresult.Type())
	reflect.Indirect(rresultPtrNone).Set(reflect.MakeMap(rresult.Type()))
	rresultNone := reflect.Indirect(rresultPtrNone)
	errIgnore := f(rpsNone.Interface(), rresultPtrNone.Interface())
	if errIgnore != nil {
		if safety {
			err = errIgnore
			return
		}
		for k, v := range vsAlterLc {
			if v == nil {
				continue
			}
			p := keysM[k]
			rresult.SetMapIndex(reflect.ValueOf(p), reflect.ValueOf(v))
		}
		return
	}

	for i := 0; i < rpsNone.Len(); i++ {
		rpNone := rpsNone.Index(i)
		pNone := rpNone.Interface()
		key := mf(pNone)
		rv := rresultNone.MapIndex(rpNone)
		if rv.IsValid() {
			rresult.SetMapIndex(rpNone, rv)
			lc.Set(key, rv.Interface(), expire)
		} else {
			lc.Set(key, nil, expire)
		}
	}

	return
}

func GlueLc(lmStru *LmStru) (err error) {
	p, result, f, mf, stru := lmStru.Input, lmStru.Output, lmStru.Proc, lmStru.Key, lmStru.Lc
	expire, safety := stru.Expire, stru.Safety

	rresult := reflect.Indirect(reflect.ValueOf(result))

	key := mf(p)
	vLc, ok := lc.Get(key)
	if vLc != nil {
		rresult.Set(reflect.Indirect(reflect.ValueOf(vLc)))
	}
	if ok {
		if vLc != nil || !safety {
			return
		}
	}

	rresultNone := reflect.New(rresult.Type())
	errIgnore := f(p, rresultNone.Interface())
	if errIgnore != nil {
		if safety {
			err = errIgnore
			return
		}
		return
	}

	rresult.Set(reflect.Indirect(rresultNone))
	lc.Set(key, result, expire)
	return
}

func Glues(lmStru *LmStru) (err error) {
	ps, result, f, mf, mcStru, lcStru := lmStru.Input, lmStru.Output, lmStru.Proc, lmStru.Key, lmStru.Mc, lmStru.Lc
	lmStru = &LmStru{
		Input:  ps,
		Output: result,
		Proc: func(ps, result interface{}) error {
			lmStru := &LmStru{
				Input:  ps,
				Output: result,
				Proc:   f,
				Key:    mf,
				Mc:     mcStru,
			}
			return GluesMc(lmStru)
		},
		Key: mf,
		Lc:  lcStru,
	}
	return GluesLc(lmStru)
}

func Glue(lmStru *LmStru) (err error) {
	p, result, f, mf, mcStru, lcStru := lmStru.Input, lmStru.Output, lmStru.Proc, lmStru.Key, lmStru.Mc, lmStru.Lc
	lmStru = &LmStru{
		Input:  p,
		Output: result,
		Proc: func(p, result interface{}) error {
			lmStru := &LmStru{
				Input:  p,
				Output: result,
				Proc:   f,
				Key:    mf,
				Mc:     mcStru,
			}
			return GlueMc(lmStru)
		},
		Key: mf,
		Lc:  lcStru,
	}
	return GlueLc(lmStru)
}
