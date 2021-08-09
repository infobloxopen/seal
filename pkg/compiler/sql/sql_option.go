package sqlcompiler

import "github.com/sirupsen/logrus"

// SQLOption is option function for SQL conversion parameters
type SQLOption func(sqlc *SQLCompiler)

// WithLogger is an option to specify SQL dialect
func WithLogger(logger *logrus.Logger) SQLOption {
	return func(sqlc *SQLCompiler) {
		sqlc.Logger = logger
	}
}

// WithDialect is an option to specify SQL dialect
func WithDialect(dialect SQLDialectEnum) SQLOption {
	return func(sqlc *SQLCompiler) {
		sqlc.Dialect = dialect
	}
}

// WithIdentifierReplacer is an option to specify SQLReplacerFn for identifiers.
// This option can be specified multiple times, and will be executed in the order specified.
func WithIdentifierReplacer(replacerFn SQLReplacerFn) SQLOption {
	return func(sqlc *SQLCompiler) {
		if sqlc.IdentifierReplacers == nil {
			sqlc.IdentifierReplacers = []SQLReplacerFn{}
		}
		sqlc.IdentifierReplacers = append(sqlc.IdentifierReplacers, replacerFn)
	}
}

// WithLiteralReplacer is an option to specify SQLReplacerFn for literal values.
// This option can be specified multiple times, and will be executed in the order specified.
func WithLiteralReplacer(replacerFn SQLReplacerFn) SQLOption {
	return func(sqlc *SQLCompiler) {
		if sqlc.LiteralReplacers == nil {
			sqlc.LiteralReplacers = []SQLReplacerFn{}
		}
		sqlc.LiteralReplacers = append(sqlc.LiteralReplacers, replacerFn)
	}
}
