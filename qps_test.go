package qps

import (
	"sync"
	"testing"
	"time"
)

type HttpQps struct {
	In  uint64 `qps:"in"`
	Out uint64 `qps:"out"`
}

func TestQpsAndPersist(t *testing.T) {
	httpOps := New(1*time.Second, 60, true, HttpQps{})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func() {
			wg.Done()
			for j := 0; j < 100; j++ {
				httpOps.Inc("in")
			}
		}()
	}
	wg.Wait()

	httpOps.Add("in", 2)
	httpOps.Add("out", 1)

	time.Sleep(3 * time.Second)
	n1 := httpOps.History(1)
	if n1[0].GetData().(HttpQps).In != 1002 {
		t.Error("In unequal to 1002")
	}
	if n1[0].GetData().(HttpQps).Out != 1 {
		t.Error("Out unequal to 1")
	}

	n2 := httpOps.History(1000)
	if n2[0].GetData().(HttpQps).In != 1002 {
		t.Error("In unequal to 1002")
	}
}

func TestNonPersist(t *testing.T) {
	httpOps := New(2*time.Second, 60, false, HttpQps{})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func() {
			wg.Done()
			for j := 0; j < 100; j++ {
				httpOps.Inc("in")
			}
		}()
	}
	wg.Wait()

	time.Sleep(3 * time.Second)
	n1 := httpOps.History(1)
	if n1[0].GetData().(HttpQps).In != 1000 {
		t.Error("In unequal to 1000")
	}

	n2 := httpOps.History(1)
	if len(n2) > 0 && n2[0].GetData().(HttpQps).In != 0 {
		t.Error("In unequal to 1000")
	}
}
