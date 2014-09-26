package miniredis

import (
	"testing"

	"github.com/garyburd/redigo/redis"
)

// Test EXPIRE. Keys with an expiration are called volatile in Redis parlance.
func TestExpire(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()
	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)

	// Not volatile yet
	{
		equals(t, 0, s.Expire("foo"))
		b, err := redis.Int(c.Do("TTL", "foo"))
		ok(t, err)
		equals(t, -2, b)
	}

	// Set something
	{
		_, err := c.Do("SET", "foo", "bar")
		ok(t, err)
		// Key exists, but no Expire set yet.
		b, err := redis.Int(c.Do("TTL", "foo"))
		ok(t, err)
		equals(t, -1, b)

		n, err := redis.Int(c.Do("EXPIRE", "foo", "1200"))
		ok(t, err)
		equals(t, 1, n) // EXPIRE returns 1 on success.

		equals(t, 1200, s.Expire("foo"))
		b, err = redis.Int(c.Do("TTL", "foo"))
		ok(t, err)
		equals(t, 1200, b)
	}

	// A SET resets the expire.
	{
		_, err := c.Do("SET", "foo", "bar")
		ok(t, err)
		b, err := redis.Int(c.Do("TTL", "foo"))
		ok(t, err)
		equals(t, -1, b)
	}

	// Set a non-existing key
	{
		n, err := redis.Int(c.Do("EXPIRE", "nokey", "1200"))
		ok(t, err)
		equals(t, 0, n) // EXPIRE returns 0 on failure.
	}

	// Remove an expire
	{

		// No key yet
		n, err := redis.Int(c.Do("PERSIST", "exkey"))
		ok(t, err)
		equals(t, 0, n)

		_, err = c.Do("SET", "exkey", "bar")
		ok(t, err)

		// No timeout yet
		n, err = redis.Int(c.Do("PERSIST", "exkey"))
		ok(t, err)
		equals(t, 0, n)

		_, err = redis.Int(c.Do("EXPIRE", "exkey", "1200"))
		ok(t, err)

		// All fine now
		n, err = redis.Int(c.Do("PERSIST", "exkey"))
		ok(t, err)
		equals(t, 1, n)

		// No TTL left
		b, err := redis.Int(c.Do("TTL", "exkey"))
		ok(t, err)
		equals(t, -1, b)
	}

	// Hash key works fine, too
	{
		_, err := c.Do("HSET", "wim", "zus", "jet")
		ok(t, err)
		b, err := redis.Int(c.Do("EXPIRE", "wim", "1234"))
		ok(t, err)
		equals(t, 1, b)
	}
}

func TestExpireat(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()
	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)

	// Not volatile yet
	{
		equals(t, 0, s.Expire("foo"))
		b, err := redis.Int(c.Do("TTL", "foo"))
		ok(t, err)
		equals(t, -2, b)
	}

	// Set something
	{
		_, err := c.Do("SET", "foo", "bar")
		ok(t, err)
		// Key exists, but no Expire set yet.
		b, err := redis.Int(c.Do("TTL", "foo"))
		ok(t, err)
		equals(t, -1, b)

		n, err := redis.Int(c.Do("EXPIREAT", "foo", 1234567890))
		ok(t, err)
		equals(t, 1, n) // EXPIREAT returns 1 on success.

		equals(t, 1234567890, s.Expire("foo"))
		b, err = redis.Int(c.Do("TTL", "foo"))
		ok(t, err)
		equals(t, 1234567890, b)
		equals(t, 1234567890, s.Expire("foo"))
	}
}

func TestPexpire(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()
	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)

	// Key exists
	{
		ok(t, s.Set("foo", "bar"))
		b, err := redis.Int(c.Do("PEXPIRE", "foo", 12))
		ok(t, err)
		equals(t, 1, b)

		e, err := redis.Int(c.Do("PTTL", "foo"))
		ok(t, err)
		equals(t, 12, e)
	}
	// Key doesn't exist
	{
		b, err := redis.Int(c.Do("PEXPIRE", "nosuch", 12))
		ok(t, err)
		equals(t, 0, b)

		e, err := redis.Int(c.Do("PTTL", "nosuch"))
		ok(t, err)
		equals(t, -2, e)
	}

	// No expire
	{
		s.Set("aap", "noot")
		e, err := redis.Int(c.Do("PTTL", "aap"))
		ok(t, err)
		equals(t, -1, e)
	}
}

func TestDel(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()
	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)

	s.Set("foo", "bar")
	s.HSet("aap", "noot", "mies")
	s.Set("one", "two")
	s.SetExpire("one", 1234)
	s.Set("three", "four")
	r, err := redis.Int(c.Do("DEL", "one", "aap", "nosuch"))
	ok(t, err)
	equals(t, 2, r)
	equals(t, 0, s.Expire("one"))

	// Direct also works:
	s.Set("foo", "bar")
	s.Del("foo")
	got, err := s.Get("foo")
	equals(t, ErrKeyNotFound, err)
	equals(t, "", got)
}

func TestType(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()
	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)

	// String key
	{
		s.Set("foo", "bar!")
		v, err := redis.String(c.Do("TYPE", "foo"))
		ok(t, err)
		equals(t, "string", v)
	}

	// Hash key
	{
		s.HSet("aap", "noot", "mies")
		v, err := redis.String(c.Do("TYPE", "aap"))
		ok(t, err)
		equals(t, "hash", v)
	}

	// New key
	{
		v, err := redis.String(c.Do("TYPE", "nosuch"))
		ok(t, err)
		equals(t, "none", v)
	}

	// Wrong usage
	{
		_, err := redis.Int(c.Do("TYPE"))
		assert(t, err != nil, "do TYPE error")
		_, err = redis.Int(c.Do("TYPE", "spurious", "arguments"))
		assert(t, err != nil, "do TYPE error")
	}

	// Direct usage:
	{
		equals(t, "hash", s.Type("aap"))
		equals(t, "", s.Type("nokey"))
	}
}

func TestExists(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()
	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)

	// String key
	{
		s.Set("foo", "bar!")
		v, err := redis.Int(c.Do("EXISTS", "foo"))
		ok(t, err)
		equals(t, 1, v)
	}

	// Hash key
	{
		s.HSet("aap", "noot", "mies")
		v, err := redis.Int(c.Do("EXISTS", "aap"))
		ok(t, err)
		equals(t, 1, v)
	}

	// New key
	{
		v, err := redis.Int(c.Do("EXISTS", "nosuch"))
		ok(t, err)
		equals(t, 0, v)
	}

	// Wrong usage
	{
		_, err := redis.Int(c.Do("EXISTS"))
		assert(t, err != nil, "do EXISTS error")
		_, err = redis.Int(c.Do("EXISTS", "spurious", "arguments"))
		assert(t, err != nil, "do EXISTS error")
	}

	// Direct usage:
	{
		equals(t, true, s.Exists("aap"))
		equals(t, false, s.Exists("nokey"))
	}
}

func TestMove(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()
	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)

	// No problem.
	{
		s.Set("foo", "bar!")
		v, err := redis.Int(c.Do("MOVE", "foo", 1))
		ok(t, err)
		equals(t, 1, v)
	}

	// Src key doesn't exists.
	{
		v, err := redis.Int(c.Do("MOVE", "nosuch", 1))
		ok(t, err)
		equals(t, 0, v)
	}

	// Target key already exists.
	{
		s.DB(0).Set("two", "orig")
		s.DB(1).Set("two", "taken")
		v, err := redis.Int(c.Do("MOVE", "two", 1))
		ok(t, err)
		equals(t, 0, v)
		s.CheckGet(t, "two", "orig")
	}

	// Wrong usage
	{
		_, err := redis.Int(c.Do("MOVE"))
		assert(t, err != nil, "do MOVE error")
		_, err = redis.Int(c.Do("MOVE", "foo"))
		assert(t, err != nil, "do MOVE error")
		_, err = redis.Int(c.Do("MOVE", "foo", "noint"))
		assert(t, err != nil, "do MOVE error")
		_, err = redis.Int(c.Do("MOVE", "foo", 2, "toomany"))
		assert(t, err != nil, "do MOVE error")
	}
}

func TestKeys(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()
	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)

	s.Set("foo", "bar!")
	s.Set("foobar", "bar!")
	s.Set("barfoo", "bar!")
	s.Set("fooooo", "bar!")

	{
		v, err := redis.Strings(c.Do("KEYS", "foo"))
		ok(t, err)
		equals(t, []string{"foo"}, v)
	}

	// simple '*'
	{
		v, err := redis.Strings(c.Do("KEYS", "foo*"))
		ok(t, err)
		equals(t, []string{"foo", "foobar", "fooooo"}, v)
	}
	// simple '?'
	{
		v, err := redis.Strings(c.Do("KEYS", "fo?"))
		ok(t, err)
		equals(t, []string{"foo"}, v)
	}

	// Don't die on never-matching pattern.
	{
		v, err := redis.Strings(c.Do("KEYS", `f\`))
		ok(t, err)
		equals(t, []string{}, v)
	}

	// Wrong usage
	{
		_, err := redis.Int(c.Do("KEYS"))
		assert(t, err != nil, "do KEYS error")
		_, err = redis.Int(c.Do("KEYS", "foo", "noint"))
		assert(t, err != nil, "do KEYS error")
	}
}

func TestRandom(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()
	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)

	// Empty db.
	{
		v, err := c.Do("RANDOMKEY")
		ok(t, err)
		equals(t, nil, v)
	}

	s.Set("one", "bar!")
	s.Set("two", "bar!")
	s.Set("three", "bar!")

	// No idea which key will be returned.
	{
		v, err := redis.String(c.Do("RANDOMKEY"))
		ok(t, err)
		assert(t, v == "one" || v == "two" || v == "three", "RANDOMKEY looks sane")
	}

	// Wrong usage
	{
		_, err = redis.Int(c.Do("RANDOMKEY", "spurious"))
		assert(t, err != nil, "do RANDOMKEY error")
	}
}

func TestRename(t *testing.T) {
	s, err := Run()
	ok(t, err)
	defer s.Close()
	c, err := redis.Dial("tcp", s.Addr())
	ok(t, err)

	// Non-existing key
	{
		_, err := redis.Int(c.Do("RENAME", "nosuch", "to"))
		assert(t, err != nil, "do RENAME error")
	}

	// Same key
	{
		_, err := redis.Int(c.Do("RENAME", "from", "from"))
		assert(t, err != nil, "do RENAME error")
	}

	// Move a string key
	{
		s.Set("from", "value")
		str, err := redis.String(c.Do("RENAME", "from", "to"))
		ok(t, err)
		equals(t, "OK", str)
		equals(t, false, s.Exists("from"))
		equals(t, true, s.Exists("to"))
		s.CheckGet(t, "to", "value")
	}

	// Move a hash key
	{
		s.HSet("from", "key", "value")
		str, err := redis.String(c.Do("RENAME", "from", "to"))
		ok(t, err)
		equals(t, "OK", str)
		equals(t, false, s.Exists("from"))
		equals(t, true, s.Exists("to"))
		equals(t, "value", s.HGet("to", "key"))
	}

	// Move over something which exists
	{
		s.Set("from", "string value")
		s.HSet("to", "key", "value")
		s.SetExpire("from", 999999)

		str, err := redis.String(c.Do("RENAME", "from", "to"))
		ok(t, err)
		equals(t, "OK", str)
		equals(t, false, s.Exists("from"))
		equals(t, true, s.Exists("to"))
		s.CheckGet(t, "to", "string value")
		equals(t, 0, s.Expire("from"))
		equals(t, 999999, s.Expire("to"))
	}

	// Wrong usage
	{
		_, err := redis.Int(c.Do("MOVE"))
		assert(t, err != nil, "do MOVE error")
		_, err = redis.Int(c.Do("MOVE", "too few"))
		assert(t, err != nil, "do MOVE error")
		_, err = redis.Int(c.Do("MOVE", "some", "spurious", "arguments"))
		assert(t, err != nil, "do MOVE error")
	}
}
