package sqlcompiler

import (
	"testing"

	"github.com/infobloxopen/seal/pkg/lexer"
	"github.com/sirupsen/logrus"
)

func TestJSONBReplacer(t *testing.T) {
	logrus.SetLevel(logrus.InfoLevel)
	//logrus.SetLevel(logrus.TraceLevel)

	tests := []struct {
		dialect   SQLDialectEnum
		jsonbOp   string
		isNumKey  bool
		input     string
		expected  string
		shouldErr bool
	}{
		{
			dialect:   DialectUnknown, // Dialect doesn't support JSONB
			input:     `ctx.tags["endangered"]`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  false,
			expected:  ``,
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags[endangered]`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  false,
			expected:  `ctx.tags->'endangered'`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags["endangered"]`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  false,
			expected:  `ctx.tags->'endangered'`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags["endangered"]`,
			jsonbOp:   JSONBTextOperator,
			isNumKey:  false,
			expected:  `ctx.tags->>'endangered'`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags["endangered"]`,
			jsonbOp:   JSONBExistsOperator,
			isNumKey:  false,
			expected:  `ctx.tags?'endangered'`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags[0]`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  false,
			expected:  `ctx.tags->'0'`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags["0"]`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  true,
			expected:  `ctx.tags->0`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags[0]`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  true,
			expected:  `ctx.tags->0`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags["non_numeric_index"]`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  true,
			expected:  ``,
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags[non_numeric_index]`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  true,
			expected:  ``,
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags[3.14]`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  true,
			expected:  ``,
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags[-1]`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  true,
			expected:  ``,
			shouldErr: true,
		},
	}

	for idx, tst := range tests {
		sqlc := NewSQLCompiler(WithDialect(tst.dialect))
		jsonbReplacer := NewJSONBReplacer(
			WithJSONBOperator(tst.jsonbOp),
			WithIsNumericKey(tst.isNumKey),
		)
		id := tst.input
		idParts := lexer.SplitIdentifier(id)
		replaced, err := jsonbReplacer(sqlc, "", idParts, id)
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
