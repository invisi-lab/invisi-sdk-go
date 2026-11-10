package provenance

func CanonicalWord(value string) (string, error)                   { return word(value) }
func HashFields(values ...string) (string, error)                  { return hash(values...) }
func DigestLimbs(value string) (string, string, error)             { return limbs(value) }
func EqualNumber(a, b string) bool                                 { return equalNumber(a, b) }
func CheckInteger(value string, bits int, nonzero bool) error      { return check(value, bits, nonzero) }
func CertificateToWire(c RootCertificate) (RootCertificate, error) { return certificateToWire(c) }
func AssertPublicSubmission(value any) error                       { return assertPublicSubmission(value) }

func DecodeFixed(value string, size int) ([]byte, error) { return decodeFixed(value, size) }
