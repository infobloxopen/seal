package sqlcompiler

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestCompileCondition(t *testing.T) {
	logrus.SetLevel(logrus.InfoLevel)
	//logrus.SetLevel(logrus.TraceLevel)

	tests := []struct {
		dialect   int
		input     string
		expected  string
		shouldErr bool
	}{
		{
			dialect:   DialectPostgres,
			input:     `age > 18`,
			expected:  ``,
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `foobar.qwerty == "there's a single-quote in this string"`,
			expected:  `(foobar.qwerty = 'there''s a single-quote in this string')`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `subject.nbf < 123 and ctx.description == "string with subject. in it"`,
			expected:  `(mysqltable.nbf < 123 AND (mysqltable.description = 'string with subject. in it'))`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `not subject.iss == "string with ctx. in it" and ctx.name =~ ".*goofy.*"`,
			expected:  `((NOT (mysqltable.iss = 'string with ctx. in it')) AND (mysqltable.name ~ '.*goofy.*'))`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags["endangered"] == "true"`,
			expected:  ``, // ``(ctx.tags["endangered"] = 'true')`, // TODO: invalid SQL?
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.id in "tag-manage", "tag-view"`,
			expected:  ``, // TODO: expect `(ctx.id IN ('tag-manage', 'tag-view'))`,
			shouldErr: true,
		},
	}

	colNameReplacer := strings.NewReplacer(
		"ctx.", "mysqltable.",
		"subject.", "mysqltable.",
	)

	for idx, tst := range tests {
		where, err := CompileCondition(tst.dialect, tst.input, colNameReplacer)
		if err != nil && !tst.shouldErr {
			t.Errorf("Test#%d: failure: unexpected err=%s for input=%s\n",
				idx, err, tst.input)
		} else if err == nil && tst.shouldErr {
			t.Errorf("Test#%d: failure: expected error for input=%s\n", idx, tst.input)
		} else if err == nil && tst.expected != where {
			t.Errorf("Test#%d: failure: input=%s expected=%s actual=%s\n",
				idx, tst.input, tst.expected, where)
		} else if err != nil && tst.shouldErr {
			t.Logf("Test#%d: success: expected error and got err=%s\n", idx, err)
		} else {
			t.Logf("Test#%d: success: where=%s\n", idx, where)
		}
	}
}
