package util

func PlayerCountSentence(count int) string {
	switch count {
	case 1:
		return "jednego"

	case 2:
		return "dwóch"

	case 3:
		return "trzech"

	case 4:
		return "czterech"

	case 5:
		return "pięciu"

	case 6:
		return "sześciu"

	case 7:
		return "siedmiu"

	case 8:
		return "ośmiu"

	case 9:
		return "dziewięciu"

	case 10:
		return "dziesięciu"

	case 11:
		return "jedenastu"

	default:
		return "kilkanastu"
	}
}
