package discord

import "github.com/bwmarrin/discordgo"

// ResolvedCommandOption represents a command option with a name and associated value.
// Name specifies the unique identifier of the command option.
// Value holds the actual data or argument associated with the option.
type ResolvedCommandOption struct {
	Name  string
	Value any
}

// String returns the string value of the ResolvedCommandOption if its type is ApplicationCommandOptionString, otherwise an empty string.
func (r *ResolvedCommandOption) String() string {
	if r.Value != nil {
		if str, ok := r.Value.(string); ok {
			return str
		}
	}

	return ""
}

// CommandInteractionOptions represents a collection of resolved command options for interaction handling.
// It enables retrieval and storage of options within a specified command execution context.
// The struct integrates a map for quick lookup and a slice for sequential retention of interaction options.
type CommandInteractionOptions struct {
	optionsMap         map[string]*ResolvedCommandOption
	interactionOptions []*discordgo.ApplicationCommandInteractionDataOption
}

// Option retrieves the ResolvedCommandOption associated with the specified name from optionsMap or interactionOptions.
// If the option is found in interactionOptions, it registers the option in optionsMap before returning it.
// Returns nil if no option is found with the given name.
func (c *CommandInteractionOptions) Option(name string) *ResolvedCommandOption {
	if option, ok := c.optionsMap[name]; ok {
		return option
	}

	for _, option := range c.interactionOptions {
		if option.Name == name {
			resolvedCommandOption := ResolvedCommandOption{
				Name:  option.Name,
				Value: option.Value,
			}

			c.optionsMap[name] = &resolvedCommandOption
			return &resolvedCommandOption
		}
	}

	return &ResolvedCommandOption{
		Name:  name,
		Value: nil,
	}
}

// CommandOption represents an option for a command in a Discord application.
// Name is the name of the command option.
// Description provides a brief explanation of the option's purpose.
// Type specifies the type of the option, such as string, integer, or boolean.
// Required determines if the option is mandatory for the command.
type CommandOption struct {
	Name        string
	Description string
	Type        discordgo.ApplicationCommandOptionType
	Required    bool
}

// ToApplicationCommandOption converts a CommandOption to a discordgo.ApplicationCommandOption for API usage.
func (o *CommandOption) ToApplicationCommandOption() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Name:        o.Name,
		Description: o.Description,
		Required:    o.Required,
		Type:        o.Type,
	}
}
