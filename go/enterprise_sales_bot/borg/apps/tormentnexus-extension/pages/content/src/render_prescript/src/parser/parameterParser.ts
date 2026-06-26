import { CONFIG } from '../core/config';
import type { Parameter, PartialParameterState } from '../core/types';
import { extractJSONParameters } from './jsonFunctionParser';
import * as cheerio from 'cheerio';

// State storage for streaming parameters
export const partialParameterState = new Map<string, PartialParameterState>();
export const streamingContentLengths = new Map<string, number>();

/**
 * Detect if content is JSON format
 */
const isJSONFormat = (content: string): boolean => {
  const trimmed = content.trim();
  return trimmed.includes('"type"') && (trimmed.includes('function_call_start') || trimmed.includes('parameter'));
};

/**
 * Extract parameters from JSON format with state tracking for real-time streaming
 */
const extractParametersFromJSON = (content: string, blockId: string | null = null): Parameter[] => {
  const parameters: Parameter[] = [];
  const jsonParams = extractJSONParameters(content);

  // Get previous state for tracking changes
  const partialParams: PartialParameterState = blockId ? partialParameterState.get(blockId) || {} : {};
  const newPartialState: PartialParameterState = {};

  // Check if streaming (no function_call_end)
  const isStreaming =
    !content.includes('"type":"function_call_end"') && !content.includes('"type": "function_call_end"');

  Object.entries(jsonParams).forEach(([name, value]) => {
    const displayValue = typeof value === 'object' ? JSON.stringify(value, null, 2) : String(value);

    // Track content length for large content handling
    if (blockId && displayValue.length > CONFIG.largeContentThreshold) {
      streamingContentLengths.set(`${blockId}-${name}`, displayValue.length);
    }

    // Determine if this is a new or changed parameter
    const isNew = !partialParams[name] || partialParams[name] !== displayValue;

    // Store current state for next iteration
    newPartialState[name] = displayValue;

    parameters.push({
      name,
      value: displayValue,
      isComplete: !isStreaming,
      isStreaming,
      isNew,
      originalContent: displayValue,
      contentLength: displayValue.length,
      isLargeContent: displayValue.length > CONFIG.largeContentThreshold,
    });
  });

  // Update state for next iteration
  if (blockId) {
    partialParameterState.set(blockId, newPartialState);
  }

  return parameters;
};

/**
 * Extract parameters from function call content
 *
 * @param content The content to extract parameters from
 * @param blockId Optional block ID for tracking streaming parameters
 * @returns Array of extracted parameters
 */
export const extractParameters = (content: string, blockId: string | null = null): Parameter[] => {
  // Check if JSON format
  if (isJSONFormat(content)) {
    return extractParametersFromJSON(content, blockId);
  }

  // XML format extraction using Cheerio AST
  const parameters: Parameter[] = [];
  const partialParams: PartialParameterState = blockId ? partialParameterState.get(blockId) || {} : {};
  const newPartialState: PartialParameterState = {};

  const $ = cheerio.load(content, null, false);
  const paramNodes = $('parameter');

  // To determine if a parameter is complete, we check if there are matching closing tags.
  // Cheerio auto-closes, so we count occurrences of `</parameter>`.
  // This is a heuristic that works well because parameters are sequential.
  const closingTagMatches = content.match(/<\/parameter>/g);
  const closingTagCount = closingTagMatches ? closingTagMatches.length : 0;

  paramNodes.each((index, el) => {
    // Determine if this specific parameter has been closed
    const isComplete = index < closingTagCount;
    
    // Check if the opening tag is incomplete (e.g. streaming `<parameter name="co`)
    // If the element has no name attribute but the raw content ends with an unclosed attribute
    let paramName = $(el).attr('name');
    let isIncompleteTag = false;
    
    if (!paramName) {
      if (index === paramNodes.length - 1 && !content.endsWith('>')) {
        // Last node, and content doesn't end with >, might be streaming `<parameter name="...`
        const partialTagMatch = content.match(/<parameter(?:\s+(?:name="([^"]*)")?[^>]*)?$/);
        if (partialTagMatch) {
          paramName = partialTagMatch[1] || 'unnamed_parameter';
          isIncompleteTag = true;
        } else {
          paramName = 'unnamed_parameter';
        }
      } else {
        paramName = 'unnamed_parameter';
      }
    }

    // Extract value
    let paramValue = $(el).text();
    
    // For incomplete tag, the value is streaming (empty in AST but we preserve the partial tag state)
    if (isIncompleteTag) {
      const partialTagMatch = content.match(/<parameter(?:\s+(?:name="([^"]*)")?[^>]*)?$/);
      paramValue = partialTagMatch ? partialTagMatch[0] : '';
      newPartialState[`__partial_tag_${Date.now()}`] = paramValue;
    }

    if (blockId && paramValue.length > CONFIG.largeContentThreshold && !isIncompleteTag) {
      streamingContentLengths.set(`${blockId}-${paramName}`, paramValue.length);
    }
    
    if (isComplete && !isIncompleteTag) {
      parameters.push({
        name: paramName,
        value: paramValue,
        isComplete: true,
      });
    } else {
      newPartialState[paramName] = paramValue;
      if (blockId && !isIncompleteTag) {
        const key = `${blockId}-${paramName}`;
        const prevLength = streamingContentLengths.get(key) || 0;
        const newLength = paramValue.length;
        streamingContentLengths.set(key, newLength);

        const isLargeContent = newLength > CONFIG.largeContentThreshold;
        const hasGrown = newLength > prevLength;
        const isNew = !partialParams[paramName] || partialParams[paramName] !== paramValue;

        parameters.push({
          name: paramName,
          value: paramValue,
          isComplete: false,
          isNew: isNew || hasGrown,
          isStreaming: true,
          originalContent: paramValue,
          isLargeContent: isLargeContent,
          contentLength: newLength,
          truncated: isLargeContent,
        });
      } else {
        parameters.push({
          name: paramName,
          value: isIncompleteTag ? '(streaming...)' : paramValue,
          isComplete: false,
          isNew: !partialParams[paramName] || partialParams[paramName] !== paramValue,
          isStreaming: true,
          originalContent: paramValue,
          isIncompleteTag,
        });
      }
    }
  });

  // Check for the specific streaming scenario where an attribute is unquoted `<parameter name="` (cheerio drops it)
  // or just `<parameter `
  if (paramNodes.length === 0 && content.includes('<parameter')) {
      const partialTagRegex = /<parameter(?:\s+(?:name="([^"]*)")?[^>]*)?$/;
      const partialTagMatch = content.match(partialTagRegex);
  
      if (partialTagMatch) {
        const paramName = partialTagMatch[1] || 'unnamed_parameter';
        const partialTag = partialTagMatch[0];
  
        // Store the partial tag with timestamp to avoid collisions
        newPartialState[`__partial_tag_${Date.now()}`] = partialTag;
  
        if (paramName && paramName !== 'unnamed_parameter') {
          const existingParam = parameters.find(p => p.name === paramName);
          if (!existingParam) {
            parameters.push({
              name: paramName,
              value: '(streaming...)',
              isComplete: false,
              isStreaming: true,
              isIncompleteTag: true,
            });
          }
        }
      }
  }

  // Update the partial state for this block if we have an ID
  if (blockId) {
    partialParameterState.set(blockId, newPartialState);
  }

  return parameters;
};
