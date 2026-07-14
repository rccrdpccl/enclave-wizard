export function stripAnsi(text: string): string {
  // biome-ignore lint/suspicious/noControlCharactersInRegex: ANSI escape sequences require the ESC control character
  return text.replace(/\x1b\[[0-9;]*m/g, "").replace(/\r/g, "");
}
