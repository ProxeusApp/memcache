package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	c := New(3 * time.Second)
	c.Put("myKey", "my Value")
	c.PutWithOtherExpiry("myKey2", "my Value2", 8*time.Second)
	c.PutWithOtherExpiry(123, 456, 10*time.Second)
	c.Put(1.456, 7.89)
	c.OnExpired = func(key interface{}, val interface{}) {
		fmt.Println("on expired", key, val)
	}

	var myVal string
	err := c.Get("myKey", &myVal)
	if err != nil {
		t.Error(err, 1)
	}
	if myVal == "" {
		t.Error("myVal shouldn't be empty", myVal)
	}
	myVal = ""
	time.Sleep(6 * time.Second)
	err = c.Get("myKey", &myVal)
	if err == nil {
		t.Error(err, 2)
	}
	if myVal != "" {
		t.Error("myVal should be empty because of timeout", myVal)
	}
	myVal = ""
	err = c.Get("myKey2", &myVal)
	if err != nil {
		t.Error(err, 3)
	}
	if myVal != "my Value2" {
		t.Error("myVal2 shouldn't be empty")
	}
	var intval int
	err = c.Get(123, &intval)
	if err != nil {
		t.Error(err, 3)
	}
	if intval != 456 {
		t.Error("intval shouldn't be empty", intval)
	}
}

func BenchmarkNewCache(b *testing.B) {
	//TODO impl
	b.StartTimer()
	c := New(3 * time.Second)
	c.Put("myKey", "my Value")
	c.PutWithOtherExpiry("myKey2", "my Value2", 8*time.Second)
	c.PutWithOtherExpiry(123, 456, 10*time.Second)
	c.Put(1.456, 7.89)
	c.OnExpired = func(key interface{}, val interface{}) {
		fmt.Println("on expired", key, val)
	}

	var myVal string
	err := c.Get("myKey", &myVal)
	if err != nil {
		return
	}
	if myVal == "" {
		fmt.Println("myVal shouldn't be empty", myVal)
	}
	myVal = ""
	time.Sleep(6 * time.Second)
	err = c.Get("myKey", &myVal)
	if err == nil {
		fmt.Println(err, 2)
	}
	if myVal != "" {
		fmt.Println("myVal should be empty because of timeout", myVal)
	}
	myVal = ""
	err = c.Get("myKey2", &myVal)
	if err != nil {
		fmt.Println(err, 3)
	}
	if myVal != "my Value2" {
		fmt.Println("myVal2 shouldn't be empty")
	}
	var intval int
	err = c.Get(123, &intval)
	if err != nil {
		fmt.Println(err, 3)
	}
	if intval != 456 {
		fmt.Println("intval shouldn't be empty", intval)
	}
	b.StopTimer()
}
