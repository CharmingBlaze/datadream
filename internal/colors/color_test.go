package colors

import "testing"

func TestParseHex(t *testing.T) {
	cases := []struct {
		hex              string
		r, g, b, a       uint8
	}{
		{"#000000", 0, 0, 0, 255},
		{"#FFFFFF", 255, 255, 255, 255},
		{"#FF0000", 255, 0, 0, 255},
		{"#FF000080", 255, 0, 0, 128},
		{"#F00", 255, 0, 0, 255},
		{"#F008", 255, 0, 0, 136},
	}
	for _, tc := range cases {
		c, err := ParseHex(tc.hex)
		if err != nil {
			t.Fatalf("%s: %v", tc.hex, err)
		}
		if c.R != tc.r || c.G != tc.g || c.B != tc.b || c.A != tc.a {
			t.Fatalf("%s: got {%d,%d,%d,%d}", tc.hex, c.R, c.G, c.B, c.A)
		}
	}
}

func TestRGBAEquivalence(t *testing.T) {
	half, _ := ParseHex("#FF000080")
	r128, _ := ParseRGBA(255, 0, 0, 128, false)
	r05, _ := ParseRGBA(255, 0, 0, 0.5, true)
	cssHalf, _ := ParseCSS("rgba(255, 0, 0, 0.5)")
	for _, c := range []Color{half, r128, r05, cssHalf} {
		if !half.Equal(c) {
			t.Fatalf("expected half red, got %+v", c)
		}
	}
}

func TestCSSNamed(t *testing.T) {
	red, _ := ParseCSS("red")
	hex, _ := ParseHex("#FF0000")
	if !red.Equal(hex) {
		t.Fatalf("css red != #FF0000")
	}
	trans, _ := ParseCSS("transparent")
	if trans.A != 0 {
		t.Fatalf("transparent alpha")
	}
	rebecca, _ := ParseCSS("rebeccapurple")
	ns, ok := ResolveNamespace("rebeccaPurple")
	if !ok || !rebecca.Equal(ns) {
		t.Fatalf("rebecca purple mismatch")
	}
	corn, ok := ResolveNamespace("cornflowerBlue")
	cssCorn, _ := ParseCSS("cornflowerblue")
	if !ok || !corn.Equal(cssCorn) {
		t.Fatalf("cornflowerBlue mismatch")
	}
}

func TestAssertions(t *testing.T) {
	black, _ := ParseHex("#000000")
	rgbBlack, _ := ParseRGB(0, 0, 0)
	if !black.Equal(rgbBlack) {
		t.Fatal("#000000 != rgb(0,0,0)")
	}
	white, _ := ParseHex("#FFFFFF")
	rgbWhite, _ := ParseRGB(255, 255, 255)
	if !white.Equal(rgbWhite) {
		t.Fatal("#FFFFFF != rgb(255,255,255)")
	}
}

func TestInvalidHex(t *testing.T) {
	_, err := ParseHex("#GG0000")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInvalidRGB(t *testing.T) {
	_, err := ParseRGB(300, 0, 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInvalidAlpha(t *testing.T) {
	_, err := ParseRGBA(255, 0, 0, 2.5, true)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWithAlpha(t *testing.T) {
	red, _ := ParseHex("#FF0000")
	half, _ := red.WithAlpha(0.5, true)
	want, _ := ParseHex("#FF000080")
	if !half.Equal(want) {
		t.Fatalf("withAlpha 0.5: got %+v", half)
	}
}

func TestCSSCaseInsensitive(t *testing.T) {
	for _, name := range []string{"rebeccapurple", "RebeccaPurple", "REBECCAPURPLE", "Rebecca Purple"} {
		c, err := ParseCSS(name)
		if err != nil {
			t.Fatalf("%q: %v", name, err)
		}
		want, _ := ParseHex("#663399")
		if !c.Equal(want) {
			t.Fatalf("%q: got %+v", name, c)
		}
	}
}
