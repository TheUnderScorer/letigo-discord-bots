package bots

import "context"

var Bots = []BotName{BotNameWojciech, BotNameTadeuszSznuk}

func GetAllFromContext(ctx context.Context) []*Bot {
	var result []*Bot

	for _, v := range Bots {
		instance := ctx.Value(v).(*Bot)

		result = append(result, instance)
	}

	return result
}
