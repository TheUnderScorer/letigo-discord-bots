export type TokenEntry =
  | string
  | ((text: string, match: string, rawMatch: string) => string);

export type TokensRecord = Record<string, TokenEntry>;

export const applyTokens = (text: string, tokens: TokensRecord) =>
  Object.keys(tokens).reduce((text, token) => {
    const tokenEntry = tokens[token];

    return text.replace(new RegExp(`{{${token}}}`, 'g'), rawMatch => {
      if (typeof tokenEntry === 'string') {
        return tokenEntry;
      }

      const match = rawMatch.replace(/{{|}}/g, '');

      return tokenEntry(text, match, rawMatch);
    });
  }, text);
