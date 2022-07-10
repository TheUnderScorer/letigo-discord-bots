export function mapCommandsForHelp(commands: string[]) {
  return commands.map(quoteCommand).join(', ');
}

export function quoteCommand(command: string) {
  return `\`${command}\``;
}
