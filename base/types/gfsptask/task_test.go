package gfsptask

import (
	"fmt"
	"testing"
	"time"
)

func TestDNSParser(t *testing.T) {
	t1 := time.Now().UnixMilli()
	t2 := time.Now().Unix()
	time.Sleep(1 * time.Second)
	t3 := time.Now().UnixMilli()
	t4 := time.Now().Unix()
	t5 := time.Since(time.UnixMilli(t1))
	t51 := time.Since(time.Unix(t1, 0))
	t6 := time.Since(time.Unix(t2, 0))
	fmt.Println(t1)
	fmt.Println(t2)
	fmt.Println(t3)
	fmt.Println(t4)
	fmt.Println(t3 - t1)
	fmt.Println(t4 - t2)
	fmt.Println(t51.Seconds())
	fmt.Println(t5.Seconds())
	fmt.Println(t6.Seconds())

	t8 := t1 + 2
	fmt.Println(t1)
	fmt.Println(t8)
	fmt.Println(time.Unix(t1, 0))
	fmt.Println(time.Unix(t8, 0))
}
