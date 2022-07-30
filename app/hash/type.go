package hash

type Type int

const (
	SHA1 Type = iota
	ED2K
	AHash
	DHash
	DHashV
	PHash
	WHash
)

func (t Type) String() string {
	switch t {
	case SHA1:
		return "SHA1"
	case ED2K:
		return "ED2K"
	case AHash:
		return "AHash"
	case DHash:
		return "DHash"
	case DHashV:
		return "DHashV"
	case PHash:
		return "PHash"
	case WHash:
		return "WHash"
	default:
		return ""
	}
}
