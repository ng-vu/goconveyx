package goconveyx_test

import (
	"strings"
	"testing"
	"time"

	. "github.com/ng-vu/goconveyx"
)

type M map[string]interface{}

type S struct {
	ID    int
	Value string
	Time  time.Time
}

func TestShouldDeepEqual(t *testing.T) {
	ict := time.FixedZone("ICT", 7*60*60)

	t.Run("OK", func(t *testing.T) {
		now := time.Now()
		msg := ShouldDeepEqual(S{ID: 1, Time: now}, S{ID: 1, Time: now})
		assertEmpty(t, msg)
	})
	t.Run("Different ID", func(t *testing.T) {
		now := time.Now()
		msg := ShouldDeepEqual(S{ID: 1, Time: now}, S{ID: 1, Time: now})
		assertEmpty(t, msg)
	})
	t.Run("Different timezone", func(t *testing.T) {
		now := time.Now()
		msg := ShouldDeepEqual(S{ID: 1, Time: now}, S{ID: 1, Time: now.In(ict)})
		assertContains(t, msg, "Should deep equal")
	})
}

func TestShouldResembleSlice(t *testing.T) {
	t.Run("Different length", func(t *testing.T) {
		msg := ShouldResembleSlice([]int{1, 1, 2, 2}, []int{2, 1, 1})
		assertContains(t, msg, "Length not equal")
	})
	t.Run("Same order", func(t *testing.T) {
		ok := ShouldResembleSlice([]int{1, 2, 3}, []int{1, 2, 3}) == ""
		assertTrue(t, ok)
	})
	t.Run("Reversed order", func(t *testing.T) {
		ok := ShouldResembleSlice([]int{3, 2, 1}, []int{1, 2, 3}) == ""
		assertTrue(t, ok)
	})
	t.Run("Different quantity", func(t *testing.T) {
		msg := ShouldResembleSlice([]int{1, 2, 2}, []int{1, 1, 2})
		assertContains(t, msg, "Not match 1 item")
	})
	t.Run("Random order", func(t *testing.T) {
		ok := ShouldResembleSlice([]int{3, 2, 1, 4, 2, 6}, []int{1, 2, 3, 4, 2, 6}) == ""
		assertTrue(t, ok)
	})
	t.Run("Random order, different quantity", func(t *testing.T) {
		msg := ShouldResembleSlice([]int{3, 2, 1, 4, 2, 6}, []int{1, 2, 3, 4, 1, 6})
		assertContains(t, msg, "Not match 3 items")
	})
}

func TestShouldResembleByKey(t *testing.T) {
	t.Run("Both must be slice", func(t *testing.T) {
		msg := ShouldResembleByKey("id")([]int{1, 2}, 2)
		assertContains(t, msg, "Both must be slice)!")
	})
	t.Run("Both must be slice of struct, *struct, map or interface", func(t *testing.T) {
		msg := ShouldResembleByKey("id")([]int{1, 2}, []string{"a"})
		assertContains(t, msg, "Both must be slice of struct, *struct, map or interface")
	})
	t.Run("Length not equal", func(t *testing.T) {
		actual := []M{{"id": 1}}
		expect := []M{{"id": 1}, {"id": 2}}
		msg := ShouldResembleByKey("id")(expect, actual)
		assertContains(t, msg, "Length not equal")
	})
	t.Run("All items must not be nil", func(t *testing.T) {
		actual := []M{{"id": 1}, nil}
		expect := []M{{"id": 1}, {"id": 2}}
		msg := ShouldResembleByKey("id")(actual, expect)
		assertContains(t, msg, "All items must not be nil (actual[1] is nil)")
	})
	t.Run("Could not get key from map", func(t *testing.T) {
		actual := []M{{"id": 1}, {"id": 2}}
		expect := []M{{"id": 1}, {"no_key": 1}}
		msg := ShouldResembleByKey("id")(actual, expect)
		assertContains(t, msg, "Could not get key from expected[1]")
	})
	t.Run("Could not get key from struct", func(t *testing.T) {
		actual := []S{{ID: 1}, {ID: 2}}
		expect := []S{{ID: 1}, {Value: "foo"}}
		msg := ShouldResembleByKey("id")(actual, expect)
		assertContains(t, msg, "Key `id` not found in struct (but it has `ID`)")
	})
	t.Run("All item keys must not be nil", func(t *testing.T) {
		actual := []M{{"id": 1}, {"id": nil}}
		expect := []M{{"id": 1}, {"id": 2}}
		msg := ShouldResembleByKey("id")(actual, expect)
		assertContains(t, msg, "All item keys must not be nil (actual[1].id is nil)")
	})
	t.Run("All item keys must be comparable", func(t *testing.T) {
		actual := []M{{"id": 1}, {"id": make(map[string]string)}}
		expect := []M{{"id": 1}, {"id": 2}}
		msg := ShouldResembleByKey("id")(actual, expect)
		assertContains(t, msg, "All item keys must be comparable (actual[1].id is not, type is `map[string]string`)")
	})
	t.Run("expected[0] and expected[1] has duplicated keys", func(t *testing.T) {
		actual := []M{{"id": 10}, {"id": 20}}
		expect := []M{{"id": 10}, {"id": 10}}
		msg := ShouldResembleByKey("id")(actual, expect)
		assertContains(t, msg, "expected[0] and expected[1] has duplicated keys: `10`")
	})
	t.Run("Expected item with id but not found (1)", func(t *testing.T) {
		actual := []M{{"id": 10}, {"id": 30}}
		expect := []M{{"id": 10}, {"id": 20}}
		msg := ShouldResembleByKey("id")(actual, expect)
		assertContains(t, msg, "Expected item with id=`20` but not found")
	})
	t.Run("Expected item with id but not found (the first)", func(t *testing.T) {
		actual := []M{{"id": 30}, {"id": 50}, {"id": 20}, {"id": 40}, {"id": 60}}
		expect := []M{{"id": 10}, {"id": 20}, {"id": 30}, {"id": 40}, {"id": 50}}
		msg := ShouldResembleByKey("id")(actual, expect)
		assertContains(t, msg, "Expected item with id=`10` but not found")
	})
	t.Run("Expected item with id but not found (the last)", func(t *testing.T) {
		actual := []M{{"id": 30}, {"id": 60}, {"id": 20}, {"id": 40}, {"id": 10}}
		expect := []M{{"id": 10}, {"id": 20}, {"id": 30}, {"id": 40}, {"id": 50}}
		msg := ShouldResembleByKey("id")(actual, expect)
		assertContains(t, msg, "Expected item with id=`50` but not found")
	})
	t.Run("Item is different", func(t *testing.T) {
		actual := []M{{"id": 10}, {"id": 20}, {"id": 30, "x": nil}}
		expect := []M{{"id": 10}, {"id": 20}, {"id": 30}}
		msg := ShouldResembleByKey("id")(actual, expect)
		assertContains(t, msg, "Item with id=`30` is different")
		assertContains(t, msg, "map[x]: <nil> != <does not have key>")
	})
	t.Run("Different order (ok)", func(t *testing.T) {
		actual := []M{{"id": 30}, {"id": 10}, {"id": 20}}
		expect := []M{{"id": 10}, {"id": 20}, {"id": 30}}
		msg := ShouldResembleByKey("id")(actual, expect)
		assertEmpty(t, msg)
	})
	t.Run("Slice of interface (ok)", func(t *testing.T) {
		actual := []interface{}{M{"id": 30}, M{"id": 10}, M{"id": 20}}
		expect := []interface{}{M{"id": 10}, M{"id": 20}, M{"id": 30}}
		msg := ShouldResembleByKey("id")(actual, expect)
		assertEmpty(t, msg)
	})
	t.Run("Slice of struct (ok)", func(t *testing.T) {
		actual := []S{{ID: 30}, {ID: 10}, {ID: 20}}
		expect := []S{{ID: 10}, {ID: 20}, {ID: 30}}
		msg := ShouldResembleByKey("ID")(actual, expect)
		assertEmpty(t, msg)
	})
	t.Run("Slice of interface struct (ok)", func(t *testing.T) {
		actual := []interface{}{S{ID: 30}, S{ID: 10}, S{ID: 20}}
		expect := []interface{}{S{ID: 10}, S{ID: 20}, S{ID: 30}}
		msg := ShouldResembleByKey("ID")(actual, expect)
		assertEmpty(t, msg)
	})
	t.Run("Slice of *struct (ok)", func(t *testing.T) {
		actual := []*S{{ID: 30}, {ID: 10}, {ID: 20}}
		expect := []*S{{ID: 10}, {ID: 20}, {ID: 30}}
		msg := ShouldResembleByKey("ID")(actual, expect)
		assertEmpty(t, msg)
	})
	t.Run("Slice of interface *struct (ok)", func(t *testing.T) {
		actual := []interface{}{&S{ID: 30}, &S{ID: 10}, &S{ID: 20}}
		expect := []interface{}{&S{ID: 10}, &S{ID: 20}, &S{ID: 30}}
		msg := ShouldResembleByKey("ID")(actual, expect)
		assertEmpty(t, msg)
	})
}

func assertTrue(t *testing.T, b bool) {
	if !b {
		t.Errorf("Expect true. Got: %v", b)
	}
}

func assertEmpty(t *testing.T, s string) {
	if s != "" {
		t.Errorf("Expect empty. Got:")
	}
	t.Log(s)
}

func assertContains(t *testing.T, s string, substr string) {
	if !strings.Contains(s, substr) {
		t.Errorf("Expect message contains `%v`. Got:", substr)
	}
	t.Log(s)
}
