package trivia

import "src-go/domain/trivia"

func FigureOutType(q *trivia.Question) {
	if q.Type != "" {
		return
	}

	if len(q.IncorrectAnswerMessages) == 1 {
		q.Type = trivia.TrueFalse
	} else {
		q.Type = trivia.MultipleChoice
	}
}
