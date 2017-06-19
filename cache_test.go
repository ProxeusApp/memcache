package cache

import(
	"testing"
	"fmt"
	"time"
)

func TestNewCache(t *testing.T){
	c := NewCache(3 * time.Second)
	c.Put("myKey", "my Value")
	c.PutWithOtherExpiry("myKey", "my Value", 6*time.Second)
	c.Put(123, 456)
	c.Put(1.456, 7.89)
	c.OnExpired = func(key interface{}, val interface{}){
		fmt.Println("on expired", key, val)
	}
	time.Sleep(10*time.Second)
	fmt.Println(c.Get("myKey"))
}
