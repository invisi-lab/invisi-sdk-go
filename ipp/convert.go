package ipp

// PaymentTypesFromStrings casts wire/storage strings (proto repeated string,
// varchar[] columns) into typed PaymentTypes. It is the single source for the
// []string→[]PaymentType cast shared across services; do NOT duplicate the loop
// per service. This is parse-don't-validate: it neither drops nor rejects
// unknown values — legality is the caller's job via PaymentType.Valid() at the
// validation boundary.
func PaymentTypesFromStrings(in []string) []PaymentType {
	out := make([]PaymentType, len(in))
	for i, s := range in {
		out[i] = PaymentType(s)
	}
	return out
}

// PaymentTypesToStrings flattens typed PaymentTypes back to wire/storage strings
// (e.g. for persisting into a varchar[] column). Single source for the reverse
// cast; callers needing a named array type (e.g. pq.StringArray) wrap the result.
func PaymentTypesToStrings(in []PaymentType) []string {
	out := make([]string, len(in))
	for i, t := range in {
		out[i] = string(t)
	}
	return out
}
