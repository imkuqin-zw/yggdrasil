package xstrings

import (
	"regexp"
	"strings"
)

// StringMatcher contains match criteria for matching a string, and is an
// internal representation of the `StringMatcher` proto defined at
// https://github.com/envoyproxy/envoy/blob/main/api/envoy/type/matcher/v3/string.proto.
type StringMatcher struct {
	// Since these match fields are part of a `oneof` in the corresponding xDS
	// proto, only one of them is expected to be set.
	exactMatch    *string
	prefixMatch   *string
	suffixMatch   *string
	regexMatch    *regexp.Regexp
	containsMatch *string
	// If true, indicates the exact/prefix/suffix/contains matching should be
	// case insensitive. This has no effect on the regex match.
	ignoreCase bool
}

func NewExactMatcher(exact *string, ignoreCase bool) StringMatcher {
	sm := StringMatcher{
		exactMatch: exact,
		ignoreCase: ignoreCase,
	}
	if ignoreCase {
		*sm.exactMatch = strings.ToLower(*exact)
	}
	return sm
}

func NewPrefixMatcher(prefix *string, ignoreCase bool) StringMatcher {
	sm := StringMatcher{
		prefixMatch: prefix,
		ignoreCase:  ignoreCase,
	}
	if ignoreCase {
		*sm.prefixMatch = strings.ToLower(*prefix)
	}
	return sm
}

func NewSuffixMatcher(suffix *string, ignoreCase bool) StringMatcher {
	sm := StringMatcher{
		prefixMatch: suffix,
		ignoreCase:  ignoreCase,
	}
	if ignoreCase {
		*sm.prefixMatch = strings.ToLower(*suffix)
	}
	return sm
}

func NewContainsMatcher(contains *string, ignoreCase bool) StringMatcher {
	sm := StringMatcher{
		prefixMatch: contains,
		ignoreCase:  ignoreCase,
	}
	if ignoreCase {
		*sm.prefixMatch = strings.ToLower(*contains)
	}
	return sm
}

func NewRegexMatcher(regex *regexp.Regexp) StringMatcher {
	return StringMatcher{regexMatch: regex}
}

// Match returns true if input matches the criteria in the given StringMatcher.
func (sm StringMatcher) Match(input string) bool {
	if sm.ignoreCase {
		input = strings.ToLower(input)
	}
	switch {
	case sm.exactMatch != nil:
		return input == *sm.exactMatch
	case sm.prefixMatch != nil:
		return strings.HasPrefix(input, *sm.prefixMatch)
	case sm.suffixMatch != nil:
		return strings.HasSuffix(input, *sm.suffixMatch)
	case sm.regexMatch != nil:
		return FullMatchWithRegex(sm.regexMatch, input)
	case sm.containsMatch != nil:
		return strings.Contains(input, *sm.containsMatch)
	}
	return false
}

// ExactMatch returns the value of the configured exact match or an empty string
// if exact match criteria was not specified.
func (sm StringMatcher) ExactMatch() string {
	if sm.exactMatch != nil {
		return *sm.exactMatch
	}
	return ""
}

// Equal returns true if other and sm are equivalent to each other.
func (sm StringMatcher) Equal(other StringMatcher) bool {
	if sm.ignoreCase != other.ignoreCase {
		return false
	}

	if (sm.exactMatch != nil) != (other.exactMatch != nil) ||
		(sm.prefixMatch != nil) != (other.prefixMatch != nil) ||
		(sm.suffixMatch != nil) != (other.suffixMatch != nil) ||
		(sm.regexMatch != nil) != (other.regexMatch != nil) ||
		(sm.containsMatch != nil) != (other.containsMatch != nil) {
		return false
	}

	switch {
	case sm.exactMatch != nil:
		return *sm.exactMatch == *other.exactMatch
	case sm.prefixMatch != nil:
		return *sm.prefixMatch == *other.prefixMatch
	case sm.suffixMatch != nil:
		return *sm.suffixMatch == *other.suffixMatch
	case sm.regexMatch != nil:
		return sm.regexMatch.String() == other.regexMatch.String()
	case sm.containsMatch != nil:
		return *sm.containsMatch == *other.containsMatch
	}
	return true
}
