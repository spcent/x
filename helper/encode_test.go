package helper

import "testing"

func TestEncodeURIComponent_Basic(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"ABC abc 123-_.!~*'()", "ABC%20abc%20123-_.!~*'()"},
		{"你好, world!", "%E4%BD%A0%E5%A5%BD%2C%20world!"},
		{"email=a@b.com&x=1+2", "email%3Da%40b.com%26x%3D1%2B2"},
		{"~tilde stays", "~tilde%20stays"},
		{"! ' () * kept", "!%20'%20()%20*%20kept"},
	}

	for _, c := range cases {
		got := EncodeURIComponent(c.in)
		if got != c.want {
			t.Fatalf("EncodeURIComponent(%q) = %q; want %q", c.in, got, c.want)
		}
		// round-trip check
		back, err := DecodeURIComponent(got)
		if err != nil {
			t.Fatalf("DecodeURIComponent(%q) error: %v", got, err)
		}
		if back != c.in {
			t.Fatalf("round-trip failed: %q -> %q -> %q", c.in, got, back)
		}
	}
}

func TestDecodeURIComponent_PlusIsPlus(t *testing.T) {
	// encodeURIComponent never turns space into '+', so '+' must be preserved
	in := "a+b"
	enc := EncodeURIComponent(in) // expect "a%2Bb"
	result := "a%2Bb"
	if enc != result {
		t.Fatalf("EncodeURIComponent(%q) = %q; want `%s`", in, enc, result)
	}
	dec, err := DecodeURIComponent(enc)
	if err != nil || dec != in {
		t.Fatalf("DecodeURIComponent(%q) = %q (err=%v); want %q", enc, dec, err, in)
	}
}

func TestEncodeURIComponent_Emoji(t *testing.T) {
	in := "emoji: 😄"
	enc := EncodeURIComponent(in)
	wantPrefix := "emoji%3A%20"
	if len(enc) < len(wantPrefix) || enc[:len(wantPrefix)] != wantPrefix {
		t.Fatalf("prefix mismatch: got %q", enc)
	}
	dec, err := DecodeURIComponent(enc)
	if err != nil || dec != in {
		t.Fatalf("round-trip failed: %q -> %q (err=%v)", enc, dec, err)
	}
}

func BenchmarkEncodeURIComponent(b *testing.B) {
	benchStr := "中文 + emoji 😄 + query: a=1&b=2&url=https://example.com/path?q=xy"
	for i := 0; i < b.N; i++ {
		_ = EncodeURIComponent(benchStr)
	}
}

func BenchmarkDecodeURIComponent(b *testing.B) {
	benchStr := "中文 + emoji 😄 + query: a=1&b=2&url=https://example.com/path?q=xy"
	enc := EncodeURIComponent(benchStr)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DecodeURIComponent(enc)
	}
}
