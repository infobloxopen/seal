package sqlcompiler

import (
	"strings"
	"testing"

	"github.com/infobloxopen/seal/pkg/lexer"
	"github.com/sirupsen/logrus"
)

func TestCompileCondition(t *testing.T) {
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
			dialect:   DialectPostgres,
			input:     `age > 18`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  false,
			expected:  ``,
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `type:contacts.profile; foobar.qwerty == "there's a single-quote in this string"`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  false,
			expected:  `(foobar.qwerty = 'there''s a single-quote in this string')`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `subject.nbf < 123 and ctx.description == "string with subject. in it"`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  false,
			expected:  `(mytable.nbf < 123 AND (mytable.description = 'string with subject. in it'))`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `not subject.iss == "string with ctx. in it" and ctx.name =~ ".*goofy.*"`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  false,
			expected:  `((NOT (mytable.iss = 'string with ctx. in it')) AND (mytable.name ~ '.*goofy.*'))`,
			shouldErr: false,
		},
		{
			dialect:   DialectUnknown, // Dialect doesn't support JSONB
			input:     `ctx.tags["endangered"] == "true"`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  false,
			expected:  ``,
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags["endangered"] == "true"`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  false,
			expected:  `(mytable.tags->'endangered' = 'true')`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags["endangered"] == "true"`,
			jsonbOp:   JSONBTextOperator,
			isNumKey:  false,
			expected:  `(mytable.tags->>'endangered' = 'true')`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags["zero"] == "true"`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  true,
			expected:  ``,
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags["0"] == "true"`, // Invalid SEAL index key: "0"
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  true,
			expected:  ``,
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags[0] == "true"`, // Invalid SEAL index key: 0
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  false,
			expected:  ``,
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags["endangered"] == 123`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  false,
			expected:  `(mytable.tags->'endangered' = 123)`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags["endangered"] == qwerty`, // invalid SEAL: unquoted literal string
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  false,
			expected:  ``,
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tagz["endangered"] == "true"`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  false,
			expected:  ``,
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.id in "tag-manage", "tag-view"`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  false,
			expected:  ``, // TODO: expect `(ctx.id IN ('tag-manage', 'tag-view'))`,
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags["endangered"] == "foo in the foobar"`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  false,
			expected:  `(mytable.tags->'endangered' = 'bar in the barbar')`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `ctx.tags["endangered"] == 314159`,
			jsonbOp:   JSONBObjectOperator,
			isNumKey:  false,
			expected:  `(mytable.tags->'endangered' = 271828)`,
			shouldErr: false,
		},
	}

	idNameReplacer := strings.NewReplacer(
		"ctx.", "mytable.",
		"subject.", "mytable.",
	)

	literalReplacer := strings.NewReplacer(
		"foo", "bar",
		"314159", "271828",
	)

	for idx, tst := range tests {
		sqlc := NewSQLCompiler(
			WithDialect(tst.dialect),
			WithIdentifierReplacer(func(sqlc *SQLCompiler, idParts *lexer.IdentifierParts, id string) (string, error) {
				if idParts.Field != "tags" {
					return id, nil
				}
				jsonbReplacer := NewJSONBReplacer(
					WithJSONBOperator(tst.jsonbOp),
					WithIsNumericKey(tst.isNumKey),
				)
				return jsonbReplacer(sqlc, idParts, id)
			}),
			WithIdentifierReplacer(func(sqlc *SQLCompiler, idParts *lexer.IdentifierParts, id string) (string, error) {
				return idNameReplacer.Replace(id), nil
			}),
			WithLiteralReplacer(func(sqlc *SQLCompiler, idParts *lexer.IdentifierParts, s string) (string, error) {
				return literalReplacer.Replace(s), nil
			}),
		)
		where, err := sqlc.CompileCondition(tst.input)
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
