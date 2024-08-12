package trivia

import (
	"src-go/messages"
	"src-go/util"
)

// GetInvalidAnswerPhraseParts returns a random phrase indicating an invalid answer
// based on the question type and correctness.
//
// For MultipleChoice questions, it selects phrases from the "MultipleLeft" category.
// For TrueFalse questions, it selects phrases from the "BooleanTrue" category if the
// correct answer is true, or from the "BooleanFalse" category if the correct answer
// is false.
//
// Parameters:
//   - question: The question for which to get the invalid answer phrase.
//   - userName: The name of the player who answered the question.
//
// Returns:
//
//	Array of strings representing the invalid answer phrase.
func (question *Question) GetInvalidAnswerPhraseParts(userName string) []string {
	var msg []string

	if question == nil {
		return msg
	}

	tokens := map[string]string{
		"ANSWER": question.Correct,
		"NAME":   userName,
	}

	switch qt := question.Type; qt {
	case MultipleChoice:

		msg = append(msg, util.ApplyTokens(util.RandomElement(messages.Messages.Trivia.InvalidAnswer.MultipleLeft), tokens))

		if util.IsValidArray(question.IncorrectAnswerMessages) {
			msg = append(msg, util.ApplyTokens(util.RandomElement(question.IncorrectAnswerMessages), tokens))
		}
	case TrueFalse:
		if question.Correct == True {
			msg = append(msg, util.ApplyTokens(util.RandomElement(messages.Messages.Trivia.InvalidAnswer.BooleanTrue), tokens))
		} else {
			msg = append(msg, util.ApplyTokens(util.RandomElement(messages.Messages.Trivia.InvalidAnswer.BooleanFalse), tokens))
		}
	}

	return msg
}

func (question *Question) GetValidAnswerMessages() (m []string) {
	if question.Type == TrueFalse {
		m = messages.Messages.Trivia.ValidAnswer.Boolean
	} else {
		m = messages.Messages.Trivia.ValidAnswer.Multiple
	}

	return m
}
