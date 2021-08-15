package sqlcompiler

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestTypeMapper(t *testing.T) {
	logrus.SetLevel(logrus.InfoLevel)
	//logrus.SetLevel(logrus.TraceLevel)

	tests := []struct {
		dialect   SQLDialectEnum
		swtype    string
		jsonbOp   string
		intFlag   bool
		input     string
		expected  string
		shouldErr bool
		asterisk  bool
	}{
		{
			dialect:   DialectUnknown, // Dialect doesn't support JSONB
			swtype:    `contacts.profile`,
			input:     `ctx.tags["endangered"]`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  ``,
			shouldErr: true,
			asterisk:  false,
		},
		{
			dialect:   DialectPostgres,
			swtype:    `contacts.address`, // unmatched swagger type
			input:     `ctx.tags["endangered"]`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  `ctx.tags["endangered"]`,
			shouldErr: false,
			asterisk:  false,
		},
		{
			dialect:   DialectPostgres,
			swtype:    `contacts.profile`,
			input:     `ctx.taggs["endangered"]`, // unmatched swagger property
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  `ctx.taggs["endangered"]`,
			shouldErr: false,
			asterisk:  false,
		},
		{
			dialect:   DialectPostgres,
			swtype:    `contacts.profile`,
			input:     `ctx.taggs["endangered"]`, // swagger property matches "*"
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  `profile.taggs->'endangered'`,
			shouldErr: false,
			asterisk:  true,
		},
		{
			dialect:   DialectPostgres,
			swtype:    `contacts.profile`,
			input:     `ctx.tags[endangered]`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  `profile.tagz->'endangered'`,
			shouldErr: false,
			asterisk:  false,
		},
		{
			dialect:   DialectPostgres,
			swtype:    `contacts.profile`,
			input:     `ctx.tags["endangered"]`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  `profile.tagz->'endangered'`,
			shouldErr: false,
			asterisk:  false,
		},
		{
			dialect:   DialectPostgres,
			swtype:    `contacts.profile`,
			input:     `ctx.tags["endangered"]`,
			jsonbOp:   JSONBTextOperator,
			intFlag:   false,
			expected:  `profile.tagz->>'endangered'`,
			shouldErr: false,
			asterisk:  false,
		},
		{
			dialect:   DialectPostgres,
			swtype:    `contacts.profile`,
			input:     `ctx.tags["endangered"]`,
			jsonbOp:   JSONBExistsOperator,
			intFlag:   false,
			expected:  `profile.tagz?'endangered'`,
			shouldErr: false,
			asterisk:  false,
		},
		{
			dialect:   DialectPostgres,
			swtype:    `contacts.profile`,
			input:     `ctx.tags[0]`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  `profile.tagz->'0'`,
			shouldErr: false,
			asterisk:  false,
		},
		{
			dialect:   DialectPostgres,
			swtype:    `contacts.profile`,
			input:     `ctx.tags["0"]`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   true,
			expected:  `profile.tagz->0`,
			shouldErr: false,
			asterisk:  false,
		},
		{
			dialect:   DialectPostgres,
			swtype:    `contacts.profile`,
			input:     `ctx.tags[0]`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   true,
			expected:  `profile.tagz->0`,
			shouldErr: false,
			asterisk:  false,
		},
		{
			dialect:   DialectPostgres,
			swtype:    `contacts.profile`,
			input:     `ctx.tags["non_numeric_index"]`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   true,
			expected:  ``,
			shouldErr: true,
			asterisk:  false,
		},
		{
			dialect:   DialectPostgres,
			swtype:    `contacts.profile`,
			input:     `ctx.tags[non_numeric_index]`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   true,
			expected:  ``,
			shouldErr: true,
			asterisk:  false,
		},
		{
			dialect:   DialectPostgres,
			swtype:    `contacts.profile`,
			input:     `ctx.tags[3.14]`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   true,
			expected:  ``,
			shouldErr: true,
			asterisk:  false,
		},
		{
			dialect:   DialectPostgres,
			swtype:    `contacts.profile`,
			input:     `ctx.tags[-1]`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   true,
			expected:  ``,
			shouldErr: true,
			asterisk:  false,
		},
	}

	for idx, tst := range tests {
		tmpr := NewTypeMapper("contacts.profile").ToSQLTable("profile").
			WithPropertyMapper(NewPropertyMapper("tags").ToSQLColumn("tagz").
				UseJSONBOperator(tst.jsonbOp).
				UseJSONBIntKeyFlag(tst.intFlag),
			)
		if tst.asterisk {
			tmpr = tmpr.WithPropertyMapper(NewPropertyMapper("*").ToSQLColumn("*").
				UseJSONBOperator(tst.jsonbOp).
				UseJSONBIntKeyFlag(tst.intFlag),
			)
		}
		_ = NewSQLCompiler().WithDialect(tst.dialect).WithTypeMapper(tmpr)
		id := tst.input
		replaced, err := tmpr.ReplaceIdentifier(tst.swtype, id)
		if err != nil && !tst.shouldErr {
			t.Errorf("Test#%d: failure: unexpected err=%s for input=%s\n",
				idx, err, tst.input)
		} else if err == nil && tst.shouldErr {
			t.Errorf("Test#%d: failure: expected error for input=%s\n", idx, tst.input)
		} else if err == nil && tst.expected != replaced {
			t.Errorf("Test#%d: failure: input=%s expected=%s actual=%s\n",
				idx, tst.input, tst.expected, replaced)
		} else if err != nil && tst.shouldErr {
			t.Logf("Test#%d: success: expected error and got err=%s\n", idx, err)
		} else {
			t.Logf("Test#%d: success: replaced=%s\n", idx, replaced)
		}
	}
}
