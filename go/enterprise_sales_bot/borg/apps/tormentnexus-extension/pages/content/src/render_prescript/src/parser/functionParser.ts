import type { FunctionInfo } from '../core/types';
import { extractLanguageTag } from './languageParser';
import { containsJSONFunctionCalls, extractJSONFunctionInfo } from './jsonFunctionParser';
import * as cheerio from 'cheerio';

/**
 * Analyzes content to determine if it contains function calls
 * and related information about their completeness
 *
 * @param block The HTML element containing potential function call content
 * @returns Information about the detected function calls
 */
export const containsFunctionCalls = (block: HTMLElement): FunctionInfo => {
  const content = block.textContent?.trim() || '';
  const result: FunctionInfo = {
    hasFunctionCalls: false,
    isComplete: false,
    hasInvoke: false,
    hasParameters: false,
    hasClosingTags: false,
    languageTag: null,
    detectedBlockType: null,
    partialTagDetected: false,
  };

  // First, check for JSON function calls
  const jsonResult = containsJSONFunctionCalls(block);
  if (jsonResult.hasFunctionCalls) {
    // Extract description for JSON format
    const { description } = extractJSONFunctionInfo(content);
    return {
      ...jsonResult,
      description: description ?? undefined,
    };
  }

  // Check for XML function call content
  if (
    !content.includes('<') &&
    !content.includes('<function_calls>') &&
    !content.includes('<invoke') &&
    !content.includes('</invoke>') &&
    !content.includes('<parameter')
  ) {
    return result;
  }

  // Detect language tag and update content to examine
  const langTagResult = extractLanguageTag(content);
  if (langTagResult.tag) {
    result.languageTag = langTagResult.tag;
  }

  // The content to analyze (with or without language tag)
  const contentToExamine = langTagResult.content || content;

  // Use cheerio for resilient AST parsing
  const $ = cheerio.load(contentToExamine, null, false);
  const invokeNode = $('invoke');
  
  if (invokeNode.length > 0) {
    result.hasFunctionCalls = true;
    result.detectedBlockType = 'antml';
    result.hasInvoke = true;
    result.invokeName = invokeNode.attr('name');

    const paramsNode = $('parameter');
    result.hasParameters = paramsNode.length > 0;

    // Check for complete structure by seeing if we can find closing tags in raw text
    // (Since parse5 auto-closes, we do a quick raw check for strictly the closing element)
    const hasOpeningTag = contentToExamine.includes('<function_calls>') || contentToExamine.includes('<invoke');
    const hasClosingTag = contentToExamine.includes('</function_calls>') || contentToExamine.includes('</invoke>');

    result.hasClosingTags = hasOpeningTag && hasClosingTag;
    result.isComplete = result.hasClosingTags;
  }

  return result;
};
