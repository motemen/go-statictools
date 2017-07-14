package testdata

// TODO: recursively nested struct
// TODO: importing packages

type Foo struct {
	A int
	B int
	C int
}

type Bar struct {
	Foo *Foo
	X   bool
}

func func1() {
	f := Foo{
		A: 1,
		B: 2,
		C: 3,
	}

	// +test
	// name=Foo
	// fields=A,B,C
	g := Foo{}

	// +test
	// name=Foo
	// fields=B,C
	h := Foo{
		A: 1,
	}

	_, _, _ = f, g, h

	// +test
	// name=Bar
	// fields=X
	x := Bar{
		Foo: &Foo{
			A: 1, B: 2, C: 3,
		},
	}

	y := Bar{
		// +test
		// name=Foo
		// fields=A
		Foo: &Foo{
			B: 2, C: 3,
		},
		X: true,
	}

	_, _ = x, y
}
