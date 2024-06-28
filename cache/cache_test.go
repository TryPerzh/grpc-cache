package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := New(10*time.Minute, 1*time.Minute)

	key := "test_key"
	value := "test_value"

	cache.Set(key, value, 0)

	val, found := cache.Get(key)
	assert.True(t, found, "Очікувалось, що ключ буде знайдений")
	assert.Equal(t, value, val, "Отримане значення має бути рівним встановленому")
}

func TestCache_Add(t *testing.T) {
	cache := New(10*time.Minute, 1*time.Minute)

	key := "test_key"
	value := "test_value"

	err := cache.Add(key, value, 0)
	assert.NoError(t, err, "Не очікувалася помилка при додаванні першого разу")

	err = cache.Add(key, "new_value", 0)
	assert.Error(t, err, "Очікувалася помилка при спробі додати існуючий ключ")
}

func TestCache_Replace(t *testing.T) {
	cache := New(10*time.Minute, 1*time.Minute)

	key := "test_key"
	value := "test_value"

	cache.Set(key, value, 0)

	err := cache.Replace(key, "new_value", 0)
	assert.NoError(t, err, "Не очікувалася помилка при заміні існуючого ключа")

	_, found := cache.Get(key)
	assert.True(t, found, "Очікувалось, що ключ залишиться в кеші після заміни")
}

func TestCache_Delete(t *testing.T) {
	cache := New(10*time.Minute, 1*time.Minute)

	key := "test_key"
	value := "test_value"

	cache.Set(key, value, 0)

	cache.Delete(key)

	_, found := cache.Get(key)
	assert.False(t, found, "Очікувалось, що ключ буде видалений з кешу")
}

func TestCache_Count(t *testing.T) {
	cache := New(10*time.Minute, 1*time.Minute)

	key1 := "key1"
	key2 := "key2"
	value := "test_value"

	cache.Set(key1, value, 0)
	cache.Set(key2, value, 0)

	count := cache.Count()
	assert.Equal(t, 2, count, "Очікувалось, що кількість елементів в кеші буде дорівнювати 2")
}
