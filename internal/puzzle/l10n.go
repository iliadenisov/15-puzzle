package puzzle

type langCode string

const (
	langCodeEn = "en"
	langCodeRu = "ru"
)

func l10nCode(lc string) langCode {
	switch langCode(lc) {
	case langCodeRu:
		return langCodeRu
	case langCodeEn:
		fallthrough
	default:
		return langCodeEn
	}
}

func l10nGameTitle(lc langCode) string {
	switch lc {
	case langCodeRu:
		return "Пятнашки"
	case langCodeEn:
		fallthrough
	default:
		return "15 Puzzle"
	}
}

func l10nSilent(lc langCode) string {
	switch lc {
	case langCodeRu:
		return "Тишина"
	case langCodeEn:
		fallthrough
	default:
		return "Silent"
	}
}

func l10nMoves(lc langCode) string {
	switch lc {
	case langCodeRu:
		return "Ходы"
	case langCodeEn:
		fallthrough
	default:
		return "Moves"
	}
}

func l10nRank(lc langCode) string {
	switch lc {
	case langCodeRu:
		return "Рейтинг"
	case langCodeEn:
		fallthrough
	default:
		return "Rank"
	}
}

func l10nWins(lc langCode) string {
	switch lc {
	case langCodeRu:
		return "Побед"
	case langCodeEn:
		fallthrough
	default:
		return "Wins"
	}
}
