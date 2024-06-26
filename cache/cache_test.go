package cache

import (
	"testing"
	"time"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.

func TestAdd(t *testing.T) {

	var c *Cache = New(10*time.Minute, 10*time.Minute)

	t.Run("Add one elemnt", func(t *testing.T) {
		c.Set("1", "Test Data", 5*time.Minute)
	})

	t.Run("Get one elemnt", func(t *testing.T) {
		item, f := c.Get("1")
		if !f {
			t.Error("Елемент відсутній")
		}
		if item.(string) != "Test Data" {
			t.Error("Значення помелкове")
		}
	})

	// name := "Gladys"
	// want := regexp.MustCompile(`\b`+name+`\b`)
	// msg, err := Hello("Gladys")
	// if !want.MatchString(msg) || err != nil {
	//     t.Fatalf(`Hello("Gladys") = %q, %v, want match for %#q, nil`, msg, err, want)
	// }
}
