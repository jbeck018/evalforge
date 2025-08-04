/**
 * EvalForge JavaScript/TypeScript SDK
 * 
 * Zero-overhead LLM observability and evaluation platform.
 */

export { EvalForge } from './client';
export { wrapOpenAI } from './wrappers/openai';
export { trace, TraceContext } from './decorators';
export { 
  TraceEvent, 
  TokenUsage, 
  EvalForgeConfig,
  OpenAICompletionResponse,
  OpenAIEmbeddingResponse 
} from './types';
export { configure, getConfig } from './config';

// Re-export for convenience
import { EvalForge } from './client';
import { configure } from './config';

export default {
  EvalForge,
  configure,
};