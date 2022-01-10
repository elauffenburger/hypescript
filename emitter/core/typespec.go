package core

import (
	"reflect"
	"strings"
)

/// typeSpecMatchType is the type of match to use for `satisfies` on `TypeSpec`.
type typeSpecMatchType int

const (
	/// typeSpecMatchTypeExact will match if the two types are identical.
	///
	/// A reference to type T and type T itself will be considered identical.
	typeSpecMatchTypeExact typeSpecMatchType = 0
	/// typeSpecMatchTypeLoose will match if the receiver is a superset of the
	/// argument type.
	typeSpecMatchTypeLoose typeSpecMatchType = 1
)

type TypeSpec struct {
	Function      *Function
	Object        *Object
	Interface     *Interface
	TypeReference *string
	Union         *Union

	unresolved bool
	resolver   func()

	// Redirect is a redirect to another TypeSpec
	Redirect *TypeSpec
}

func (t *TypeSpec) Unresolved() bool {
	return t.unresolved
}

func (t *TypeSpec) MarkResolved() {
	t.resolver()
}

func (t TypeSpec) String() string {
	switch {
	case t.TypeReference != nil:
		return *t.TypeReference
	case t.Interface != nil:
		return t.Interface.Name
	case t.Union != nil:
		union := strings.Builder{}

		i := 0
		n := len(t.Union.Types)
		for t := range t.Union.Types {
			union.WriteString(t.String())

			if i != n-1 {
				union.WriteString("|")
			}

			i++
		}

		return union.String()
	default:
		return "TypeSpec{}"
	}
}

/// Equals returns true if t is a reference to other, or other is a reference to t.
func (t *TypeSpec) Equals(other *TypeSpec) bool {
	return t.satisfies(other, typeSpecMatchTypeExact)
}

/// Satisfies returns true if t can satisfy the requirements of other.
func (t *TypeSpec) Satisfies(other *TypeSpec) bool {
	return t.satisfies(other, typeSpecMatchTypeLoose)
}

/// RefersTo returns true if t is a reference to other or other references the same type as t.
func (t *TypeSpec) RefersTo(other *TypeSpec) bool {
	if t.TypeReference != nil {
		if other.TypeReference != nil && *other.TypeReference == *t.TypeReference {
			return true
		}

		if other.Interface != nil && other.Interface.Name == *t.TypeReference {
			return true
		}
	}

	return false
}

func (t *TypeSpec) satisfies(other *TypeSpec, matchType typeSpecMatchType) bool {
	// If we don't need an exact match, follow redirects.
	if matchType != typeSpecMatchTypeExact {
		for t.Redirect != nil {
			t = t.Redirect
		}

		for other.Redirect != nil {
			other = other.Redirect
		}
	}

	if other.TypeReference != nil && *other.TypeReference == string(TsAny) {
		return true
	}

	if t.RefersTo(other) || other.RefersTo(t) {
		return true
	}

	if other.Object != nil {
		if t.Object != nil {
			return satisfiesMembers(t.Object.AllMembers(), other.Object.AllMembers(), matchType)
		}

		if t.Interface != nil {
			return satisfiesMembers(t.Interface.AllMembers(), other.Object.AllMembers(), matchType)
		}
	}

	if other.Interface != nil {
		if t.Object != nil {
			return satisfiesMembers(t.Object.AllMembers(), other.Interface.AllMembers(), matchType)
		}

		if t.Interface != nil {
			return satisfiesMembers(t.Interface.AllMembers(), other.Interface.AllMembers(), matchType)
		}
	}

	if other.Union != nil {
		if t.Union != nil {
			// If we need an exact match, make sure the unions line up exactly.
			if matchType == typeSpecMatchTypeExact && len(t.Union.Types) != len(other.Union.Types) {
				return false
			}

			// Make sure each type in t is also in other.
			for ut := range t.Union.Types {
				if _, ok := other.Union.Types[ut]; !ok {
					return false
				}
			}

			return true
		}

		// Otherwise, just make sure t can satisfy one of the types in other.
		for ut := range other.Union.Types {
			if t.Satisfies(ut) {
				return true
			}
		}

		return false
	}

	return false
}

func satisfiesMembers(members, targetMembers map[string]*Member, matchType typeSpecMatchType) bool {
	// If we need an exact match, the fields of t must line up exactly (not just be a superset of other).
	if matchType == typeSpecMatchTypeExact && len(members) != len(targetMembers) {
		return false
	}

	// Make sure all the fields line up.
	for name, tgtMember := range targetMembers {
		// Make sure t contains the member.
		member, ok := members[name]
		if !ok {
			return false
		}

		if tgtMember.Field != nil {
			if member.Field == nil {
				return false
			}

			if !tgtMember.Field.Type.satisfies(member.Field.Type, matchType) {
				return false
			}
		}

		if tgtMember.Function != nil {
			if member.Function == nil {
				return false
			}

			if member.Function.Name != tgtMember.Function.Name {
				return false
			}

			if !member.Function.ImplicitReturnType.satisfies(tgtMember.Function.ImplicitReturnType, matchType) {
				return false
			}
		}
	}

	return true
}

/// EqualsStrict returns true if t is deeply equal to other.
///
/// If you want to test if the TypeSpecs are loosely equal via references,
/// use EqualsReferencing.
func (t *TypeSpec) EqualsStrict(other *TypeSpec) bool {
	return reflect.DeepEqual(t, other)
}
