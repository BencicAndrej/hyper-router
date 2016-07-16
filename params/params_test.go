package params

import (
	"context"
	"testing"
)

func TestParamsSetAndRetrieve(t *testing.T) {
	params := Params{
		Param{"foo", "bar"},
		Param{"baz", "foo"},
	}

	tests := []struct {
		key, want string
		ok        bool
	}{
		{"foo", "bar", true},
		{"baz", "foo", true},
		{"fizz", "", false},
	}

	for _, test := range tests {
		got, ok := params.ByName(test.key)

		if got != test.want || ok != test.ok {
			t.Errorf(
				"params.ByName('%s'): ('%s', '%v'), wanted ('%s', '%v')",
				test.key,
				got,
				ok,
				test.want,
				test.ok,
			)
		}
	}
}

func TestContextSetAndRetrieve(t *testing.T) {
	tests := []struct {
		loadParams  Params
		extractOk   bool
		paramsTests []struct {
			key, want string
			ok        bool
		}
	}{
		{
			loadParams: Params{
				Param{"foo", "bar"},
				Param{"baz", "foo"},
			},
			extractOk: true,
			paramsTests: []struct {
				key, want string
				ok        bool
			}{
				{"foo", "bar", true},
				{"baz", "foo", true},
				{"fizz", "", false},
			},
		},
		{
			loadParams:  nil,
			extractOk:   false,
			paramsTests: nil,
		},
	}

	for _, test := range tests {
		ctx := context.Background()

		if test.loadParams != nil {
			ctx = NewContext(ctx, test.loadParams)
		}

		params, ok := FromContext(ctx)
		if ok != test.extractOk {
			t.Errorf("FromContext(ctx): ok is %v, wanted %v")
			continue
		}

		if !ok {
			continue
		}

		for _, paramTest := range test.paramsTests {
			got, ok := params.ByName(paramTest.key)

			if got != paramTest.want || ok != paramTest.ok {
				t.Errorf(
					"params.ByName('%s'): ('%s', '%v'), wanted ('%s', '%v')",
					paramTest.key,
					got,
					ok,
					paramTest.want,
					paramTest.ok,
				)
			}
		}
	}

}
