// SPDX-License-Identifier: Apache-2.0

// Package codec converts between substrait-go's domain types and the generated
// Substrait protobuf messages. It is a separate module so that serialization can
// move out of the core.
//
// Enum conversions come in pairs and preserve values in both directions, including
// numbers a newer version of the spec may define.
//
// Message conversions come in pairs too. A nil message stays nil in both directions,
// so a field the spec makes optional round trips absent instead of turning into an
// empty message.
//
// codec imports the core. The core must never import codec.
package codec
