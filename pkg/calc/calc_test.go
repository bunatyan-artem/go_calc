package calc

import (
	"testing"
)

type Test struct {
	expression string
	answer     float64
	err        bool
}

var tests = []Test{
	{"2 + 3", 5, false},
	{"2 +3", 5, false},
	{"2 *6+ 3", 15, false},
	{"2 + 8*3", 26, false},
	{"2 + 3", 5, false},
	{"2 + 3)", 0, true},
	{"(2 + 3)", 5, false},
	{"(2 + 3", 0, true},
	{"((2 + 3)", 0, true},
	{"(2 + 3))", 0, true},
	{"2 + 3()", 0, true},
	{"12/5", 2.4, false},
	{"0 - 12/5", -2.4, false},
	{"12345679*9", 111111111, false},
	{"1/0", 0, true},
	{"", 0, true},
	{"4", 4, false},
	{"*4", 0, true},
	{"(*4)", 0, true},
	{"()*4", 0, true},
	{"+1 2", 0, true},
	{"text", 0, true},
	{"text5", 0, true},
	{"t+ 6", 0, true},
	{"2 + 1*", 0, true},
	{"2 + 3)", 0, true},
	{"(2+(3+4))", 9, false},
	{"3(2 + 3)", 0, true},
	{"(+ 2 + 3)", 0, true},
}

func TestCalc(t *testing.T) {
	for i, test := range tests {
		_TestCalc(t, i, &test)
	}
}

func _TestCalc(t *testing.T, i int, test *Test) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("#%d: panic in test \"%s\" - %s", i, test.expression, r)
		}
	}()

	answer, err := Calc(test.expression)
	if answer != test.answer && !test.err {
		t.Errorf("#%d: %s=%g; want %g", i, test.expression, answer, test.answer)
	}
	if err == nil && test.err {
		t.Errorf("#%d: \"%s\"; throw nil, want error", i, test.expression)
	} else if err != nil && !test.err {
		t.Errorf("#%d: \"%s\"; throw %s, want nil", i, test.expression, err)
	}
}
