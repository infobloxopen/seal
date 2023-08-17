package sqlcompiler

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestCompileCondition(t *testing.T) {
	logrus.SetLevel(logrus.InfoLevel)
	//logrus.SetLevel(logrus.TraceLevel)

	tests := []struct {
		dialect   SQLDialectEnum
		jsonbOp   string
		intFlag   bool
		input     string
		expected  string
		shouldErr bool
	}{
		{
			dialect:   DialectPostgres,
			input:     `type:contacts.profile; age > 18`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  ``,
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `type:eos.vpnpolicy; foobar.qwerty == "there's a single-quote in this string"`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  `(foobar.qwerty = 'there''s a single-quote in this string')`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `type:contacts.profile; subject.nbf < 123 and ctx.description == "string with subject. in it"`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  `(profile.nbf < 123 AND (profile.description = 'string with subject. in it'))`,
			shouldErr: false,
		},
		{
			dialect:   DialectUnknown,
			input:     `type:contacts.profile; not subject.iss == "string with ctx. in it" and ctx.name =~ ".*goofy.*"`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  ``,
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `type:contacts.profile; not subject.iss == "string with ctx. in it" and ctx.name =~ ".*goofy.*"`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  `((NOT (profile.iss = 'string with ctx. in it')) AND (profile.name ~ '.*goofy.*'))`,
			shouldErr: false,
		},
		{
			dialect:   DialectUnknown, // Dialect doesn't support JSONB
			input:     `type:contacts.profile; ctx.tags["endangered"] == "true"`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  ``,
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `type:contacts.profile; ctx.tags["endangered"] == "true"`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  `(profile.tagz->'endangered' = 'true')`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `type:contacts.profile; ctx.tags["endangered"] == "true"`,
			jsonbOp:   JSONBTextOperator,
			intFlag:   false,
			expected:  `(profile.tagz->>'endangered' = 'true')`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `type:contacts.profile; ctx.tags["zero"] == "true"`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   true,
			expected:  ``,
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `type:contacts.profile; ctx.tags["0"] == "true"`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   true,
			expected:  `(profile.tagz->0 = 'true')`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `type:contacts.profile; ctx.tags[0] == "true"`, // Invalid SEAL index key: 0
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  ``,
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `type:contacts.profile; ctx.tags["endangered"] == 123`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  `(profile.tagz->'endangered' = 123)`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `type:contacts.profile; ctx.tags["endangered"] == qwerty`, // invalid SEAL: unquoted literal string
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  ``,
			shouldErr: true,
		},
		{
			dialect:   DialectPostgres,
			input:     `type:contacts.address; ctx.tags["endangered"] == "true"`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  `(address.labels->'endangered' = 'true')`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `type:ddi.ipam; ctx.notes["color"] == "true"`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  `(ipam.notes->'color' = 'true')`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `type:contacts.profile; ctx.id in ["tag-manage", "tag-view", 123]`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  `(profile.id IN ('tag-manage','tag-view',123))`,
			shouldErr: false,
		},
		{
			dialect:   DialectPostgres,
			input:     `type:petstore.pet; "boss" in subject.groups`,
			jsonbOp:   JSONBObjectOperator,
			intFlag:   false,
			expected:  ``, // TODO UNSUPPORTED; may be expected should be ('boss' IN (SELECT groups FROM subject))
			shouldErr: true,
		},
	}

	for idx, tst := range tests {
		sqlc := NewSQLCompiler().WithDialect(tst.dialect).
			WithTypeMapper(NewTypeMapper("contacts.*").ToSQLTable("*").
				WithPropertyMapper(NewPropertyMapper("tags").ToSQLColumn("labels").
					UseJSONBOperator(tst.jsonbOp).
					UseJSONBIntKeyFlag(tst.intFlag),
				),
			).
			WithTypeMapper(NewTypeMapper("contacts.profile").ToSQLTable("profile").
				WithPropertyMapper(NewPropertyMapper("*").ToSQLColumn("*").
					UseJSONBOperator(tst.jsonbOp).
					UseJSONBIntKeyFlag(tst.intFlag),
				).
				WithPropertyMapper(NewPropertyMapper("tags").ToSQLColumn("tagz").
					UseJSONBOperator(tst.jsonbOp).
					UseJSONBIntKeyFlag(tst.intFlag),
				),
			).
			WithTypeMapper(NewTypeMapper("ddi.ipam").ToSQLTable("*").
				WithPropertyMapper(NewPropertyMapper("notes").ToSQLColumn("*").
					UseJSONBOperator(tst.jsonbOp).
					UseJSONBIntKeyFlag(tst.intFlag),
				),
			)
		where, err := sqlc.CompileCondition(tst.input)
		if err != nil && !tst.shouldErr {
			t.Errorf("Test#%d: failure: unexpected err=%s for input=%s\n",
				idx, err, tst.input)
		} else if err == nil && tst.shouldErr {
			t.Errorf("Test#%d: failure: expected error for input=%s and got where=%s\n",
				idx, tst.input, where)
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
