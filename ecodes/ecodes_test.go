package ecodes

import (
	"testing"
)

func TestGPBTNIndexingFunction(t *testing.T) {
  var list []int
  for i := 0x130; i <= 0x13e; i += 1 {
    ni := MapGPToIndex(i)
    list = append(list, ni)
  }
  for i := 0x220; i <= 0x223; i += 1 {
    ni := MapGPToIndex(i)
    list = append(list, ni)
  }
  var ni2 int
  ni2 = MapGPToIndex(0x00)
  list = append(list, ni2)
  ni2 = MapGPToIndex(0x01)
  list = append(list, ni2)
  ni2 = MapGPToIndex(0x03)
  list = append(list, ni2)
  ni2 = MapGPToIndex(0x04)
  list = append(list, ni2)
  for i := -4; i <= -1; i += 1 {
    ni := MapGPToIndex(i)
    list = append(list, ni)
  }
  for i := -10; i <= -7; i += 1 {
    ni := MapGPToIndex(i)
    list = append(list, ni)
  }
  for i := GP_INDEX_BASE; i < GP_INDEX_MAX; i += 1 {
    if i < len(list) {
      if i != list[i] {
        t.Log("list is out of order")
        t.Fatal(list)
      }
    } else {
      t.Log("list is not long enough")
      t.Fatal(list)
    }
  } 

  t.Log(list)
}
