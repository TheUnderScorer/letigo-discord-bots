import { applyTokens, TokensRecord } from './tokens';

describe('applyTokens', () => {
  interface TestCase {
    text: string;
    tokens: TokensRecord;
    expected: string;
  }

  const testCases: TestCase[] = [
    {
      text: 'Hello, {{name}}!',
      tokens: {
        name: 'John',
      },
      expected: 'Hello, John!',
    },
    {
      text: 'Hello, {{name}}!',
      tokens: {
        name: (value, match, rawMatch) => {
          expect(rawMatch).toEqual('{{name}}');
          expect(match).toEqual('name');

          return 'John';
        },
      },
      expected: 'Hello, John!',
    },
  ];

  testCases.forEach((testCase, index) => {
    it(`should apply tokens to text - #${index}`, () => {
      const result = applyTokens(testCase.text, testCase.tokens);

      expect(result).toBe(testCase.expected);
    });
  });
});
