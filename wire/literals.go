// SPDX-License-Identifier: Apache-2.0

package wire

import (
	"fmt"

	"github.com/substrait-io/substrait-go/v9/expr"
	"github.com/substrait-io/substrait-go/v9/types"
	proto "github.com/substrait-io/substrait-protobuf/go/substraitpb"
)

// LiteralToProto encodes a literal as its protobuf message.
func LiteralToProto(l expr.Literal) *proto.Expression_Literal {
	switch l := l.(type) {
	case *expr.NullLiteral:
		return nullLiteralToProto(l)
	case *expr.PrimitiveLiteral[bool]:
		return primitiveLiteralToProto(l)
	case *expr.PrimitiveLiteral[int8]:
		return primitiveLiteralToProto(l)
	case *expr.PrimitiveLiteral[int16]:
		return primitiveLiteralToProto(l)
	case *expr.PrimitiveLiteral[int32]:
		return primitiveLiteralToProto(l)
	case *expr.PrimitiveLiteral[int64]:
		return primitiveLiteralToProto(l)
	case *expr.PrimitiveLiteral[float32]:
		return primitiveLiteralToProto(l)
	case *expr.PrimitiveLiteral[float64]:
		return primitiveLiteralToProto(l)
	case *expr.PrimitiveLiteral[string]:
		return primitiveLiteralToProto(l)
	case *expr.PrimitiveLiteral[types.Timestamp]:
		return primitiveLiteralToProto(l)
	case *expr.PrimitiveLiteral[types.Date]:
		return primitiveLiteralToProto(l)
	case *expr.PrimitiveLiteral[types.Time]:
		return primitiveLiteralToProto(l)
	case *expr.PrimitiveLiteral[types.FixedChar]:
		return primitiveLiteralToProto(l)
	case *expr.PrimitiveLiteral[types.TimestampTz]:
		return primitiveLiteralToProto(l)
	case *expr.NestedLiteral[expr.StructLiteralValue]:
		return nestedLiteralToProto(l)
	case *expr.NestedLiteral[expr.ListLiteralValue]:
		return nestedLiteralToProto(l)
	case *expr.ByteSliceLiteral[[]byte]:
		return byteSliceLiteralToProto(l)
	case *expr.ByteSliceLiteral[types.FixedBinary]:
		return byteSliceLiteralToProto(l)
	case *expr.ByteSliceLiteral[types.UUID]:
		return byteSliceLiteralToProto(l)
	case *expr.MapLiteral:
		return mapLiteralToProto(l)
	default:
		panic(fmt.Sprintf("wire: unhandled literal %T", l))
	}
}

func nullLiteralToProto(n *expr.NullLiteral) *proto.Expression_Literal {
	return &proto.Expression_Literal{
		Nullable:               true,
		TypeVariationReference: n.Type.GetTypeVariationReference(),
		LiteralType:            &proto.Expression_Literal_Null{Null: TypeToProto(n.Type)},
	}
}

func primitiveLiteralToProto[T expr.PrimitiveLiteralValue](l *expr.PrimitiveLiteral[T]) *proto.Expression_Literal {
	lit := &proto.Expression_Literal{
		Nullable:               l.Type.GetNullability() == types.NullabilityNullable,
		TypeVariationReference: l.Type.GetTypeVariationReference(),
	}

	switch v := any(l.Value).(type) {
	case bool:
		lit.LiteralType = &proto.Expression_Literal_Boolean{Boolean: v}
	case int8:
		lit.LiteralType = &proto.Expression_Literal_I8{I8: int32(v)}
	case int16:
		lit.LiteralType = &proto.Expression_Literal_I16{I16: int32(v)}
	case int32:
		lit.LiteralType = &proto.Expression_Literal_I32{I32: v}
	case int64:
		lit.LiteralType = &proto.Expression_Literal_I64{I64: v}
	case float32:
		lit.LiteralType = &proto.Expression_Literal_Fp32{Fp32: v}
	case float64:
		lit.LiteralType = &proto.Expression_Literal_Fp64{Fp64: v}
	case string:
		lit.LiteralType = &proto.Expression_Literal_String_{String_: v}
	case types.Timestamp:
		lit.LiteralType = &proto.Expression_Literal_Timestamp{Timestamp: int64(v)}
	case types.Date:
		lit.LiteralType = &proto.Expression_Literal_Date{Date: int32(v)}
	case types.Time:
		lit.LiteralType = &proto.Expression_Literal_Time{Time: int64(v)}
	case types.FixedChar:
		lit.LiteralType = &proto.Expression_Literal_FixedChar{FixedChar: string(v)}
	case types.TimestampTz:
		lit.LiteralType = &proto.Expression_Literal_TimestampTz{TimestampTz: int64(v)}
	default:
		panic("invalid primitive literal type")
	}

	return lit
}

func nestedLiteralToProto[T expr.StructLiteralValue | expr.ListLiteralValue](l *expr.NestedLiteral[T]) *proto.Expression_Literal {
	lit := &proto.Expression_Literal{
		Nullable:               l.Type.GetNullability() == types.NullabilityNullable,
		TypeVariationReference: l.Type.GetTypeVariationReference(),
	}

	vals := make([]*proto.Expression_Literal, len(l.Value))
	for i, v := range l.Value {
		vals[i] = LiteralToProto(v)
	}

	switch any(l.Value).(type) {
	case expr.StructLiteralValue:
		lit.LiteralType = &proto.Expression_Literal_Struct_{
			Struct: &proto.Expression_Literal_Struct{Fields: vals},
		}
	case expr.ListLiteralValue:
		if len(vals) == 0 {
			lit.LiteralType = &proto.Expression_Literal_EmptyList{
				EmptyList: TypeToProto(l.Type).GetList(),
			}
		} else {
			lit.LiteralType = &proto.Expression_Literal_List_{
				List: &proto.Expression_Literal_List{Values: vals},
			}
		}
	}

	return lit
}

func mapLiteralToProto(l *expr.MapLiteral) *proto.Expression_Literal {
	lit := &proto.Expression_Literal{
		Nullable:               l.Type.GetNullability() == types.NullabilityNullable,
		TypeVariationReference: l.Type.GetTypeVariationReference(),
	}

	if len(l.Value) == 0 {
		lit.LiteralType = &proto.Expression_Literal_EmptyMap{
			EmptyMap: TypeToProto(l.Type).GetMap(),
		}
	} else {
		kv := make([]*proto.Expression_Literal_Map_KeyValue, len(l.Value))
		for i, v := range l.Value {
			kv[i] = &proto.Expression_Literal_Map_KeyValue{
				Key:   LiteralToProto(v.Key),
				Value: LiteralToProto(v.Value),
			}
		}
		lit.LiteralType = &proto.Expression_Literal_Map_{
			Map: &proto.Expression_Literal_Map{KeyValues: kv},
		}
	}

	return lit
}

func byteSliceLiteralToProto[T ~[]byte](l *expr.ByteSliceLiteral[T]) *proto.Expression_Literal {
	lit := &proto.Expression_Literal{
		Nullable:               l.Type.GetNullability() == types.NullabilityNullable,
		TypeVariationReference: l.Type.GetTypeVariationReference(),
	}

	switch v := any(l.Value).(type) {
	case []byte:
		lit.LiteralType = &proto.Expression_Literal_Binary{Binary: v}
	case types.FixedBinary:
		lit.LiteralType = &proto.Expression_Literal_FixedBinary{FixedBinary: v}
	case types.UUID:
		lit.LiteralType = &proto.Expression_Literal_Uuid{Uuid: v}
	}

	return lit
}
