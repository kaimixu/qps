package qps

import (
	"reflect"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type Qps struct {
	nType reflect.Type

	node interface{}

	// qps统计间隔
	interval time.Duration

	// 缓存结点是否持久化，持久化表示调用History接口不删除已拉取的数据，否则删除已拉取的数据
	persist bool

	// 缓存结点个数
	period int

	sync.RWMutex

	// 缓存结点
	history []*Node
}

type Node struct {
	data interface{}
	ct   time.Time
}

func (n *Node) GetData() interface{} {
	return n.data
}

func (n *Node) GetCt() time.Time {
	return n.ct
}

func (q *Qps) Inc(tag string) {
	q.Add(tag, 1)
}

func (q *Qps) Add(tag string, n uint) {
	et := reflect.TypeOf(q.node).Elem()
	ev := reflect.ValueOf(q.node).Elem()

	for i := 0; i < et.NumField(); i++ {
		if et.Field(i).Tag.Get("qps") == tag {
			f := ev.Field(i)
			if !f.CanSet() {
				panic("tag " + tag + " cann't set")
			}

			ptr := unsafe.Pointer(f.UnsafeAddr())
			switch f.Kind() {
			case reflect.Uint:
				if unsafe.Sizeof(int(0)) == 4 {
					atomic.AddUint32((*uint32)(ptr), uint32(n))
				} else {
					atomic.AddUint64((*uint64)(ptr), uint64(n))
				}
			case reflect.Uint32:
				atomic.AddUint32((*uint32)(ptr), uint32(n))
			case reflect.Uint64:
				atomic.AddUint64((*uint64)(ptr), uint64(n))
			default:
				panic("not support reflect.Value " + f.Kind().String())
			}
		}
	}
}

func (q *Qps) History(n int) []*Node {
	if n < 0 {
		panic("invalid argument n")
	}

	q.Lock()
	defer q.Unlock()
	if len(q.history) == 0 {
		return []*Node{}
	}

	if n > len(q.history) {
		n = len(q.history)
	}
	data := make([]*Node, 0, n)
	data = append(data, q.history[:n]...)

	if !q.persist {
		q.history = q.history[n:]
	}

	return data
}

func (q *Qps) stat() {
	itr := reflect.New(q.nType).Interface()
	nes := reflect.ValueOf(itr).Elem()
	es := reflect.ValueOf(q.node).Elem()

	for i := 0; i < es.NumField(); i++ {
		var v uint64

		ptr := unsafe.Pointer(es.Field(i).UnsafeAddr())
		switch es.Field(i).Kind() {
		case reflect.Uint:
			if unsafe.Sizeof(int(0)) == 4 {
				v = uint64(atomic.SwapUint32((*uint32)(ptr), 0))
			} else {
				v = atomic.SwapUint64((*uint64)(ptr), 0)
			}
		case reflect.Uint32:
			v = uint64(atomic.SwapUint32((*uint32)(ptr), 0))
		case reflect.Uint64:
			v = atomic.SwapUint64((*uint64)(ptr), 0)
		}

		nes.Field(i).SetUint(v)
	}

	newn := &Node{
		data: nes.Interface(),
		ct:   time.Now(),
	}
	q.Lock()
	defer q.Unlock()

	q.history = append(q.history, newn)
	if len(q.history) > q.period {
		q.history = q.history[1:]
	}
}

func (s *Qps) cron() {
	tk := time.NewTicker(s.interval)
	for {
		<-tk.C
		s.stat()
	}
}

func New(interval time.Duration, period int, persist bool, node interface{}) *Qps {
	q := &Qps{
		nType:    reflect.TypeOf(node),
		node:     reflect.New(reflect.TypeOf(node)).Interface(),
		persist:  persist,
		interval: interval,
		period:   period,
		history:  make([]*Node, 0, period),
	}

	go q.cron()
	return q
}
