export function arrayToOrRegex(arr: string[]) {
  return new RegExp(arr.join('|'), 'g');
}
