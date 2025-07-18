import DOMPurify from 'isomorphic-dompurify';
import { parse as YAMLParse } from 'yaml';

const tokens = ['---', '---'];
const regexHeader = new RegExp('^' + tokens[0] + '([\\s|\\S]*?)' + tokens[1]);
const regexContent = new RegExp('^ *?\\' + tokens[0] + '[^]*?' + tokens[1] + '*');

export function getHeader<T = any>(data: string): T {
  const header = getRawHeader(data);
  return YAMLParse(header) as T;
}

export function getRawHeader(data: string): string {
  const header = regexHeader.exec(data);
  if (!header) {
    throw new Error("Can't get the header");
  }
  return header[1];
}

export async function getContent(data: string): Promise<string> {
  const marked = await import('marked')

  const content = data.replace(regexContent, '').replace(/<!--(.*?)-->/gm, '');
  if (!content) {
    throw new Error("Can't get the content");
  }
  return DOMPurify.sanitize(marked.marked(content) as string);
}
