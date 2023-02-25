export const createMockChannel = <T>(messages: T[]) => ({
  isText: () => true,
  send: jest.fn(),
  messages: {
    fetch: jest.fn().mockResolvedValue({
      values: () => messages,
    }),
  },
});
