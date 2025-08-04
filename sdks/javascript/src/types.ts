/**
 * Type definitions for EvalForge SDK
 */

export interface TokenUsage {
  prompt: number;
  completion: number;
  total: number;
}

export interface TraceEvent {
  id: string;
  project_id: number;
  trace_id: string;
  span_id: string;
  parent_span_id?: string;
  operation_type: string;
  start_time: string; // ISO 8601 format
  end_time: string;   // ISO 8601 format
  duration_ms: number;
  status: 'success' | 'error' | 'timeout';
  input: Record<string, any>;
  output: Record<string, any>;
  metadata: Record<string, any>;
  tokens: TokenUsage;
  cost: number;
  provider: string;
  model: string;
}

export interface EvalForgeConfig {
  apiKey: string;
  projectId: number;
  baseUrl?: string;
  batchSize?: number;
  flushInterval?: number;
  maxRetries?: number;
  timeout?: number;
  debug?: boolean;
}

export interface EventBatch {
  events: TraceEvent[];
}

// OpenAI types for wrapper integration
export interface OpenAIMessage {
  role: 'system' | 'user' | 'assistant';
  content: string;
}

export interface OpenAICompletionRequest {
  model: string;
  messages: OpenAIMessage[];
  temperature?: number;
  max_tokens?: number;
  top_p?: number;
  frequency_penalty?: number;
  presence_penalty?: number;
  stop?: string | string[];
  stream?: boolean;
}

export interface OpenAIChoice {
  index: number;
  message: OpenAIMessage;
  finish_reason: string | null;
}

export interface OpenAIUsage {
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
}

export interface OpenAICompletionResponse {
  id: string;
  object: string;
  created: number;
  model: string;
  choices: OpenAIChoice[];
  usage: OpenAIUsage;
  system_fingerprint?: string;
}

export interface OpenAIEmbeddingRequest {
  model: string;
  input: string | string[];
  encoding_format?: 'float' | 'base64';
  dimensions?: number;
  user?: string;
}

export interface OpenAIEmbeddingData {
  object: string;
  embedding: number[];
  index: number;
}

export interface OpenAIEmbeddingResponse {
  object: string;
  data: OpenAIEmbeddingData[];
  model: string;
  usage: OpenAIUsage;
}

// Rate limiting and error handling
export interface RetryOptions {
  maxRetries: number;
  baseDelay: number;
  maxDelay: number;
  jitter: boolean;
}

export interface RateLimiter {
  checkLimit(): Promise<void>;
  reset(): void;
}

// Utility types
export type EventHandler<T = any> = (event: T) => void;
export type ErrorHandler = (error: Error) => void;

export interface Logger {
  debug(message: string, ...args: any[]): void;
  info(message: string, ...args: any[]): void;
  warn(message: string, ...args: any[]): void;
  error(message: string, ...args: any[]): void;
}