package imprint

import "hash"

// HashOptions mirrors hashstructure.HashOptions for migration convenience.
type HashOptions struct {
	Hasher          hash.Hash64
	TagName         string
	ZeroNil         bool
	IgnoreZeroValue bool
	SlicesAsSets    bool
	UseStringer     bool
}

// FromHashOptions converts a HashOptions struct into functional Option values.
// This provides a migration path from hashstructure v2.
func FromHashOptions(hopts *HashOptions) []Option {
	if hopts == nil {
		return nil
	}

	var opts []Option

	if hopts.Hasher != nil {
		opts = append(opts, WithHasher(hopts.Hasher))
	}
	if hopts.TagName != "" {
		opts = append(opts, WithTagName(hopts.TagName))
	}
	if hopts.ZeroNil {
		opts = append(opts, WithZeroNil(true))
	}
	if hopts.IgnoreZeroValue {
		opts = append(opts, WithIgnoreZeroValue(true))
	}
	if hopts.SlicesAsSets {
		opts = append(opts, WithSlicesAsSets(true))
	}
	if hopts.UseStringer {
		opts = append(opts, WithUseStringer(true))
	}

	return opts
}
