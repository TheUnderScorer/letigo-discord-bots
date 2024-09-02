package discord

type MemberDefinition struct {
	Nickname  string
	FirstName string
	LastName  string
}

// Friends contains static map of discord id to a member definition
var Friends = map[string]MemberDefinition{
	"611471485218848768": {
		Nickname:  "Amaterasu",
		FirstName: "Paulina",
		LastName:  "Grzanek",
	},
	"144797044236615680": {
		Nickname:  "Mawgan",
		FirstName: "Artur",
		LastName:  "Wieczorek",
	},
	"300692223769575425": {
		Nickname:  "Ravutto",
		FirstName: "Rafał",
		LastName:  "Czarnecki",
	},
	"498528474408157232": {
		Nickname:  "Tamako",
		FirstName: "Marek",
		LastName:  "Heintzel",
	},
	"600656821568405508": {
		Nickname:  "Walter_441",
		FirstName: "Zachary",
		LastName:  "Bujok",
	},
	"691271489907064922": {
		Nickname:  "PureGold",
		FirstName: "Wojtek",
		LastName:  "Górny",
	},
	"417397822372052992": {
		Nickname:  "Szat265",
		FirstName: "Emila",
		LastName:  "Gerlicka",
	},
	"170268157494296576": {
		Nickname:  "Taliön",
		FirstName: "Przemek",
		LastName:  "Żydek",
	},
}

func GetFriend(id string) (*MemberDefinition, bool) {
	friend, ok := Friends[id]

	return &friend, ok
}
