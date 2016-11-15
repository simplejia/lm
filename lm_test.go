package lm

import (
	"fmt"
	"strconv"
	"time"

	"github.com/simplejia/lc"

	"github.com/garyburd/redigo/redis"
)

var pool *redis.Pool

func init() {
	pool = &redis.Pool{
		Dial: func() (c redis.Conn, err error) {
			c, err = redis.Dial("tcp", ":40003")
			if err != nil {
				return
			}
			return
		},
	}

	lc.Init(65536)
}

type s1 struct {
	A int
}

func f0(ids []uint64, r *map[uint64]s1) (err error) {
	m := map[uint64]s1{
		33: s1{
			A: 9,
		},
		34: s1{
			A: 10,
		},
	}
	for _, id := range ids {
		if d, ok := m[id]; ok {
			(*r)[id] = d
		}
	}
	return
}

func f1(ids []uint64, r *map[uint64]*s1) (err error) {
	m := map[uint64]*s1{
		33: &s1{
			A: 9,
		},
		34: &s1{
			A: 10,
		},
	}
	for _, id := range ids {
		if d, ok := m[id]; ok {
			(*r)[id] = d
		}
	}
	return
}

func f2(id uint64, r *s1) (err error) {
	m := map[uint64]*s1{
		33: &s1{
			A: 9,
		},
	}
	if v, ok := m[id]; ok {
		*r = *v
	}
	return
}

func f3(id uint64, r **s1) (err error) {
	m := map[uint64]*s1{
		33: &s1{
			A: 9,
		},
	}
	*r = m[id]
	return
}

func f4(id, rid uint64, r *[]*s1) (err error) {
	m := map[string][]*s1{
		"33_44": []*s1{
			&s1{
				A: 9,
			},
			&s1{
				A: 10,
			},
		},
		"55_66": []*s1{
			&s1{
				A: 9,
			},
			&s1{
				A: 10,
			},
		},
	}
	if v, ok := m[fmt.Sprintf("%v_%v", id, rid)]; ok {
		*r = v
	}
	return
}

func f5(id, rid uint64, r *[]s1) (err error) {
	m := map[string][]s1{
		"33_44": []s1{
			s1{
				A: 9,
			},
			s1{
				A: 10,
			},
		},
	}
	if v, ok := m[fmt.Sprintf("%v_%v", id, rid)]; ok {
		*r = v
	}
	return
}

func mf1(kind string, id uint64) (ret string) {
	ret = "lm_" + kind + strconv.FormatUint(id, 10)
	return
}

func mf2(kind string, id, rid uint64) (ret string) {
	ret = "lm_" + kind + strconv.FormatUint(id, 10) + "_" + strconv.FormatUint(rid, 10)
	return
}

func tGluesLc() {
	ids := []uint64{1, 2, 33}
	r := map[uint64]*s1{}

	lmStru := &LmStru{
		Input:  ids,
		Output: &r,
		Proc: func(ps, result interface{}) error {
			return f1(ps.([]uint64), result.(*map[uint64]*s1))
		},
		Key: func(p interface{}) string {
			return mf1("GluesLc", p.(uint64))
		},
		Lc: &LcStru{
			Expire: time.Second,
			Safety: false,
		},
	}
	err := GluesLc(lmStru)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(r)
}

func tGlueLc() {
	id := uint64(33)
	var r *s1

	lmStru := &LmStru{
		Input:  id,
		Output: &r,
		Proc: func(p, r interface{}) error {
			return f3(p.(uint64), r.(**s1))
		},
		Key: func(p interface{}) string {
			return mf1("GlueLc", p.(uint64))
		},
		Lc: &LcStru{
			Expire: time.Second,
			Safety: false,
		},
	}
	err := GlueLc(lmStru)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(r)
}

func tGlueFlatLc() {
	id := uint64(33)
	rid := uint64(44)
	var r []*s1

	lmStru := &LmStru{
		Input:  []interface{}{id, rid},
		Output: &r,
		Proc: func(p, r interface{}) error {
			return f4(p.([]interface{})[0].(uint64), p.([]interface{})[1].(uint64), r.(*[]*s1))
		},
		Key: func(p interface{}) string {
			return mf2("GlueFlatLc", p.([]interface{})[0].(uint64), p.([]interface{})[1].(uint64))
		},
		Lc: &LcStru{
			Expire: time.Second,
			Safety: false,
		},
	}
	err := GlueLc(lmStru)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(r)
}

func tGluesMc() {
	ids := []uint64{1, 2, 33}
	var r map[uint64]*s1

	lmStru := &LmStru{
		Input:  ids,
		Output: &r,
		Proc: func(ps, result interface{}) error {
			return f1(ps.([]uint64), result.(*map[uint64]*s1))
		},
		Key: func(p interface{}) string {
			return mf1("GluesMc", p.(uint64))
		},
		Mc: &McStru{
			Expire: time.Second,
			Pool:   pool,
		},
	}
	err := GluesMc(lmStru)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(r)
}

func tGlueMc() {
	id := uint64(33)
	var r *s1

	lmStru := &LmStru{
		Input:  id,
		Output: &r,
		Proc: func(p, r interface{}) error {
			return f3(p.(uint64), r.(**s1))
		},
		Key: func(p interface{}) string {
			return mf1("GlueMc", p.(uint64))
		},
		Mc: &McStru{
			Expire: time.Second,
			Pool:   pool,
		},
	}
	err := GlueMc(lmStru)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(r)
}

func tGlueFlatMc() {
	id := uint64(33)
	rid := uint64(44)
	var r []*s1

	lmStru := &LmStru{
		Input:  []interface{}{id, rid},
		Output: &r,
		Proc: func(p, r interface{}) error {
			return f4(p.([]interface{})[0].(uint64), p.([]interface{})[1].(uint64), r.(*[]*s1))
		},
		Key: func(p interface{}) string {
			return mf2("GlueFlatMc", p.([]interface{})[0].(uint64), p.([]interface{})[1].(uint64))
		},
		Mc: &McStru{
			Expire: time.Second,
			Pool:   pool,
		},
	}
	err := GlueMc(lmStru)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(r)
}

func tGlues() {
	ids := []uint64{1, 2, 33}
	var r map[uint64]*s1

	lmStru := &LmStru{
		Input:  ids,
		Output: &r,
		Proc: func(ps, r interface{}) error {
			return f1(ps.([]uint64), r.(*map[uint64]*s1))
		},
		Key: func(p interface{}) string {
			return mf1("Glues", p.(uint64))
		},
		Mc: &McStru{
			Expire: time.Second * 2,
			Pool:   pool,
		},
		Lc: &LcStru{
			Expire: time.Second,
			Safety: false,
		},
	}
	err := Glues(lmStru)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(r)
}

func tGlue() {
	id := uint64(33)
	var r *s1

	lmStru := &LmStru{
		Input:  id,
		Output: &r,
		Proc: func(p, r interface{}) error {
			return f3(p.(uint64), r.(**s1))
		},
		Key: func(p interface{}) string {
			return mf1("Glue", p.(uint64))
		},
		Mc: &McStru{
			Expire: time.Second * 2,
			Pool:   pool,
		},
		Lc: &LcStru{
			Expire: time.Second,
			Safety: false,
		},
	}
	err := Glue(lmStru)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(r)
}

func tGlueFlat() {
	id := uint64(55)
	rid := uint64(66)
	var r []*s1

	lmStru := &LmStru{
		Input:  []interface{}{id, rid},
		Output: &r,
		Proc: func(p, r interface{}) error {
			return f4(p.([]interface{})[0].(uint64), p.([]interface{})[1].(uint64), r.(*[]*s1))
		},
		Key: func(p interface{}) string {
			return mf2("GlueFlat", p.([]interface{})[0].(uint64), p.([]interface{})[1].(uint64))
		},
		Mc: &McStru{
			Expire: time.Second * 2,
			Pool:   pool,
		},
		Lc: &LcStru{
			Expire: time.Second,
			Safety: false,
		},
	}
	err := Glue(lmStru)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(r)
}

func ExampleGlueLc() {
	tGlueLc()
	return
}

func ExampleGlueFlatLc() {
	tGlueFlatLc()
	return
}

func ExampleGluesLc() {
	tGluesLc()
	return
}

func ExampleGlueMc() {
	tGlueMc()
	return
}

func ExampleGlueFlatMc() {
	tGlueFlatMc()
	return
}

func ExampleGluesMc() {
	tGluesMc()
	return
}

func ExampleGlue() {
	tGlue()
	return
}

func ExampleGlueFlat() {
	tGlueFlat()
	return
}

func ExampleGlues() {
	tGlues()
	return
}
