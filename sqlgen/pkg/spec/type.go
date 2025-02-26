package spec

import (
	"database/sql"
	"fmt"

	"github.com/xyzbit/codegen/sqlgen/pkg/parameter"
	"github.com/pingcap/parser/mysql"
)

const (
	// TypeNullLongLong is a type extension for mysql.TypeLongLong.
	TypeNullLongLong byte = 0xf0
	// TypeNullDecimal is a type extension for mysql.TypeDecimal.
	TypeNullDecimal byte = 0xf1
	// TypeNullString is a type extension for mysql.TypeString.
	TypeNullString byte = 0xf2
)

const defaultThirdDecimalPkg = "github.com/shopspring/decimal"

type typeKey struct {
	tp            byte
	signed        bool
	thirdPkg      string
	aggregateCall bool
	sql.NullFloat64
}

var typeMapper = map[typeKey]string{
	{tp: mysql.TypeTiny}:                   "int8",
	{tp: mysql.TypeTiny, signed: true}:     "uint8",
	{tp: mysql.TypeShort}:                  "int16",
	{tp: mysql.TypeShort, signed: true}:    "uint16",
	{tp: mysql.TypeLong}:                   "int32",
	{tp: mysql.TypeLong, signed: true}:     "uint32",
	{tp: mysql.TypeFloat}:                  "float64",
	{tp: mysql.TypeDouble}:                 "float64",
	{tp: mysql.TypeTimestamp}:              "time.Time",
	{tp: mysql.TypeLonglong}:               "int64",
	{tp: mysql.TypeLonglong, signed: true}: "uint64",
	{tp: mysql.TypeInt24}:                  "int32",
	{tp: mysql.TypeInt24, signed: true}:    "uint32",
	{tp: mysql.TypeDate}:                   "time.Time",
	{tp: mysql.TypeDuration}:               "time.Time",
	{tp: mysql.TypeDatetime}:               "time.Time",
	{tp: mysql.TypeYear}:                   "string",
	{tp: mysql.TypeVarchar}:                "string",
	{tp: mysql.TypeBit}:                    "byte",
	{tp: mysql.TypeJSON}:                   "string",
	{
		tp:       mysql.TypeNewDecimal,
		thirdPkg: defaultThirdDecimalPkg,
	}: "decimal.Decimal",
	{
		tp:       TypeNullDecimal,
		thirdPkg: defaultThirdDecimalPkg,
	}: "decimal.NullDecimal",
	{tp: mysql.TypeEnum}:       "string",
	{tp: mysql.TypeSet}:        "string",
	{tp: mysql.TypeTinyBlob}:   "string",
	{tp: mysql.TypeMediumBlob}: "string",
	{tp: mysql.TypeLongBlob}:   "string",
	{tp: mysql.TypeBlob}:       "string",
	{tp: mysql.TypeVarString}:  "string",
	{tp: mysql.TypeString}:     "string",
	{tp: TypeNullString}:       "sql.NullString",

	// aggregate functions
	{tp: mysql.TypeTiny, aggregateCall: true}:     "sql.NullInt16",
	{tp: mysql.TypeShort, aggregateCall: true}:    "sql.NullInt16",
	{tp: mysql.TypeLong, aggregateCall: true}:     "sql.NullInt32",
	{tp: mysql.TypeFloat, aggregateCall: true}:    "sql.NullInt32",
	{tp: mysql.TypeDouble, aggregateCall: true}:   "sql.NullFloat64",
	{tp: mysql.TypeLonglong, aggregateCall: true}: "sql.NullInt64",
	{tp: mysql.TypeInt24, aggregateCall: true}:    "sql.NullInt32",
	{tp: mysql.TypeYear, aggregateCall: true}:     "sql.NullString",
	{tp: mysql.TypeVarchar, aggregateCall: true}:  "sql.NullString",
	{tp: mysql.TypeBit, aggregateCall: true}:      "sql.NullInt16",
	{tp: mysql.TypeJSON, aggregateCall: true}:     "sql.NullString",
	{
		tp:            mysql.TypeNewDecimal,
		thirdPkg:      defaultThirdDecimalPkg,
		aggregateCall: true,
	}: "decimal.NullDecimal",
	{
		tp:            TypeNullDecimal,
		thirdPkg:      defaultThirdDecimalPkg,
		aggregateCall: true,
	}: "decimal.NullDecimal",
	{tp: mysql.TypeEnum, aggregateCall: true}:       "sql.NullString",
	{tp: mysql.TypeSet, aggregateCall: true}:        "sql.NullString",
	{tp: mysql.TypeTinyBlob, aggregateCall: true}:   "sql.NullString",
	{tp: mysql.TypeMediumBlob, aggregateCall: true}: "sql.NullString",
	{tp: mysql.TypeLongBlob, aggregateCall: true}:   "sql.NullString",
	{tp: mysql.TypeBlob, aggregateCall: true}:       "sql.NullString",
	{tp: mysql.TypeVarString, aggregateCall: true}:  "sql.NullString",
	{tp: mysql.TypeString, aggregateCall: true}:     "sql.NullString",
	{tp: mysql.TypeString, aggregateCall: true}:     "sql.NullString",
	{tp: TypeNullLongLong}:                          "sql.NullInt64",
	{tp: TypeNullDecimal}:                           "decimal.NullDecimal",
	{tp: TypeNullString}:                            "sql.NullString",
	{tp: TypeNullLongLong, aggregateCall: true}:     "sql.NullInt64",
	{tp: TypeNullDecimal, aggregateCall: true}:      "decimal.NullDecimal",
	{tp: TypeNullString, aggregateCall: true}:       "sql.NullString",
}

// Type is the type of the column.
type Type byte

// DataType returns the Go type, third-package of the column.
func (c Column) DataType() (parameter.Parameter, error) {
	key := typeKey{tp: c.TP, signed: c.Unsigned, aggregateCall: c.AggregateCall}
	if c.AggregateCall {
		key = typeKey{tp: c.TP, aggregateCall: c.AggregateCall}
	}
	if c.TP == mysql.TypeNewDecimal {
		key.thirdPkg = defaultThirdDecimalPkg
	}

	goType, ok := typeMapper[key]
	if !ok {
		return parameter.Parameter{}, fmt.Errorf("unsupported type: %v", c.TP)
	}

	return NewParameter(c.Name, goType, key.thirdPkg), nil
}

// GoType returns the Go type of the column.
func (c Column) GoType() (string, error) {
	p, err := c.DataType()
	return p.Type, err
}

func (c Column) HasComment() bool {
	return len(c.Comment) > 0
}

func isNullType(tp byte) bool {
	return tp >= TypeNullLongLong && tp <= TypeNullString
}
