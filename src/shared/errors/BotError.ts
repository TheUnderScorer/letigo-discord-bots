/**
 * BotError indicates error that will contain message that will be sent directly to the channel
 * */
export class BotError extends Error {
  constructor(message: string, readonly errorContext?: string) {
    super(message);
  }

  get messageContent() {
    const parts = [this.message];

    if (this.errorContext) {
      parts.push(`\`${this.errorContext}\``);
    }

    return parts.join('\n');
  }
}
