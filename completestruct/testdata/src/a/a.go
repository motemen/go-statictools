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

	g := Foo{} // want `struct fields missing in Foo{} literal: A, B, C`

	h := Foo{ // want `struct fields missing in Foo{} literal: B, C`
		A: 1,
	}

	_, _, _ = f, g, h

	x := Bar{ // want `struct fields missing in Bar{} literal: X`
		Foo: &Foo{
			A: 1, B: 2, C: 3,
		},
	}

	y := Bar{
		Foo: &Foo{ // want `struct fields missing in Foo{} literal: A`
			B: 2, C: 3,
		},
		X: true,
	}

	_, _ = x, y
}
