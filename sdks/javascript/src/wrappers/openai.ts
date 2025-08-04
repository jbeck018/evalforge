/**
 * OpenAI wrapper for automatic tracing
 */

import { 
  OpenAICompletionRequest,
  OpenAICompletionResponse,
  OpenAIEmbeddingRequest,
  OpenAIEmbeddingResponse,
  TokenUsage 
} from '../types';
import { getClient } from '../client';
import { generateId } from '../utils/id-generator';

/**
 * OpenAI pricing per 1K tokens (as of late 2023)
 */
const OPENAI_PRICING: Record<string, { prompt: number; completion: number }> = {
  // GPT-4 models
  'gpt-4': { prompt: 0.03, completion: 0.06 },
  'gpt-4-32k': { prompt: 0.06, completion: 0.12 },
  'gpt-4-1106-preview': { prompt: 0.01, completion: 0.03 },
  'gpt-4-0125-preview': { prompt: 0.01, completion: 0.03 },
  'gpt-4-turbo-preview': { prompt: 0.01, completion: 0.03 },
  
  // GPT-3.5 models
  'gpt-3.5-turbo': { prompt: 0.001, completion: 0.002 },
  'gpt-3.5-turbo-16k': { prompt: 0.003, completion: 0.004 },
  'gpt-3.5-turbo-1106': { prompt: 0.001, completion: 0.002 },
  'gpt-3.5-turbo-0125': { prompt: 0.0005, completion: 0.0015 },
  
  // Embedding models
  'text-embedding-ada-002': { prompt: 0.0001, completion: 0 },
  'text-embedding-3-small': { prompt: 0.00002, completion: 0 },
  'text-embedding-3-large': { prompt: 0.00013, completion: 0 },
};

function calculateOpenAICost(model: string, tokens: TokenUsage): number {
  const pricing = OPENAI_PRICING[model] || OPENAI_PRICING['gpt-3.5-turbo'];
  const promptCost = (tokens.prompt / 1000) * pricing.prompt;
  const completionCost = (tokens.completion / 1000) * pricing.completion;
  return promptCost + completionCost;
}

function extractTokenUsage(response: any): TokenUsage {
  if (response?.usage) {
    return {
      prompt: response.usage.prompt_tokens || 0,
      completion: response.usage.completion_tokens || 0,
      total: response.usage.total_tokens || 0,
    };
  }
  return { prompt: 0, completion: 0, total: 0 };
}

/**
 * Wrap OpenAI client for automatic tracing
 */
export function wrapOpenAI<T extends any>(OpenAIClass: new (...args: any[]) => T): new (...args: any[]) => T {
  return class WrappedOpenAI extends OpenAIClass {
    constructor(...args: any[]) {
      super(...args);
      
      // Wrap chat completions
      if (this.chat?.completions?.create) {
        const originalCreate = this.chat.completions.create.bind(this.chat.completions);
        this.chat.completions.create = wrapChatCompletion(originalCreate);
      }
      
      // Wrap embeddings
      if (this.embeddings?.create) {
        const originalCreate = this.embeddings.create.bind(this.embeddings);
        this.embeddings.create = wrapEmbedding(originalCreate);
      }
    }
  };
}

function wrapChatCompletion(originalMethod: Function) {
  return async function(this: any, request: OpenAICompletionRequest, options?: any): Promise<OpenAICompletionResponse> {
    const client = getClient();
    if (!client) {
      return originalMethod.call(this, request, options);
    }

    const traceId = generateId();
    const spanId = generateId();
    const startTime = new Date();

    try {
      const response = await originalMethod.call(this, request, options);
      const endTime = new Date();

      // Extract response data
      const tokens = extractTokenUsage(response);
      const cost = calculateOpenAICost(request.model, tokens);

      // Get response content
      let outputContent = '';
      let finishReason = null;
      if (response.choices && response.choices.length > 0) {
        const choice = response.choices[0];
        outputContent = choice.message?.content || '';
        finishReason = choice.finish_reason;
      }

      client.trace({
        operationType: 'llm_call',
        input: {
          model: request.model,
          messages: request.messages,
          temperature: request.temperature,
          max_tokens: request.max_tokens,
          top_p: request.top_p,
          frequency_penalty: request.frequency_penalty,
          presence_penalty: request.presence_penalty,
          stop: request.stop,
        },
        output: {
          content: outputContent,
          finish_reason: finishReason,
        },
        metadata: {
          provider: 'openai',
          response_id: response.id,
          system_fingerprint: response.system_fingerprint,
          created: response.created,
        },
        traceId,
        spanId,
        startTime,
        endTime,
        status: 'success',
        tokens,
        cost,
        provider: 'openai',
        model: request.model,
      });

      return response;
    } catch (error) {
      const endTime = new Date();

      client.trace({
        operationType: 'llm_call',
        input: {
          model: request.model,
          messages: request.messages,
          temperature: request.temperature,
          max_tokens: request.max_tokens,
          top_p: request.top_p,
          frequency_penalty: request.frequency_penalty,
          presence_penalty: request.presence_penalty,
          stop: request.stop,
        },
        output: {},
        metadata: {
          provider: 'openai',
          error: error instanceof Error ? error.message : 'Unknown error',
          error_type: error instanceof Error ? error.constructor.name : 'UnknownError',
        },
        traceId,
        spanId,
        startTime,
        endTime,
        status: 'error',
        tokens: { prompt: 0, completion: 0, total: 0 },
        cost: 0,
        provider: 'openai',
        model: request.model,
      });

      throw error;
    }
  };
}

function wrapEmbedding(originalMethod: Function) {
  return async function(this: any, request: OpenAIEmbeddingRequest, options?: any): Promise<OpenAIEmbeddingResponse> {
    const client = getClient();
    if (!client) {
      return originalMethod.call(this, request, options);
    }

    const traceId = generateId();
    const spanId = generateId();
    const startTime = new Date();

    try {
      const response = await originalMethod.call(this, request, options);
      const endTime = new Date();

      // Extract response data
      const tokens = extractTokenUsage(response);
      const cost = calculateOpenAICost(request.model, tokens);

      client.trace({
        operationType: 'embedding',
        input: {
          model: request.model,
          input: request.input,
          encoding_format: request.encoding_format,
          dimensions: request.dimensions,
          user: request.user,
        },
        output: {
          embedding_count: response.data?.length || 0,
        },
        metadata: {
          provider: 'openai',
          object: response.object,
        },
        traceId,
        spanId,
        startTime,
        endTime,
        status: 'success',
        tokens,
        cost,
        provider: 'openai',
        model: request.model,
      });

      return response;
    } catch (error) {
      const endTime = new Date();

      client.trace({
        operationType: 'embedding',
        input: {
          model: request.model,
          input: request.input,
          encoding_format: request.encoding_format,
          dimensions: request.dimensions,
          user: request.user,
        },
        output: {},
        metadata: {
          provider: 'openai',
          error: error instanceof Error ? error.message : 'Unknown error',
          error_type: error instanceof Error ? error.constructor.name : 'UnknownError',
        },
        traceId,
        spanId,
        startTime,
        endTime,
        status: 'error',
        tokens: { prompt: 0, completion: 0, total: 0 },
        cost: 0,
        provider: 'openai',
        model: request.model,
      });

      throw error;
    }
  };
}